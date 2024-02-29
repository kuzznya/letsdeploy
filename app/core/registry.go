package core

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"github.com/kuzznya/letsdeploy/app/apperrors"
	"github.com/kuzznya/letsdeploy/app/middleware"
	"github.com/kuzznya/letsdeploy/app/storage"
	"github.com/kuzznya/letsdeploy/internal/openapi"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	applyConfigsCoreV1 "k8s.io/client-go/applyconfigurations/core/v1"
	"k8s.io/client-go/kubernetes"
	"net/url"
	"strings"
)

const regcredSecretName = managedSecretPrefix + "regcred"

type ContainerRegistries interface {
	projectSynchronizable
	GetProjectContainerRegistries(project string, auth middleware.Authentication) ([]openapi.ContainerRegistry, error)
	AddContainerRegistry(ctx context.Context, project string, registry openapi.ContainerRegistry, auth middleware.Authentication) (openapi.ContainerRegistry, error)
	DeleteContainerRegistry(ctx context.Context, project string, id int, auth middleware.Authentication) error
}

type containerRegistriesImpl struct {
	projects  Projects
	storage   *storage.Storage
	clientset *kubernetes.Clientset
}

var _ ContainerRegistries = (*containerRegistriesImpl)(nil)

func InitContainerRegistries(projects Projects, storage *storage.Storage, clientset *kubernetes.Clientset) ContainerRegistries {
	return &containerRegistriesImpl{
		projects:  projects,
		storage:   storage,
		clientset: clientset,
	}
}

func (r containerRegistriesImpl) GetProjectContainerRegistries(project string, auth middleware.Authentication) ([]openapi.ContainerRegistry, error) {
	return r.getProjectContainerRegistries(r.storage, project, false, auth)
}

func (r containerRegistriesImpl) AddContainerRegistry(ctx context.Context, project string, registry openapi.ContainerRegistry, auth middleware.Authentication) (openapi.ContainerRegistry, error) {
	if err := r.projects.checkAccess(project, auth); err != nil {
		return openapi.ContainerRegistry{}, err
	}

	if !isRegistryUrlValid(registry.Url) {
		return openapi.ContainerRegistry{}, apperrors.BadRequest("Invalid container registry URL")
	}

	entity := storage.ContainerRegistryEntity{
		ProjectId: project,
		Url:       registry.Url,
		Username:  registry.Username,
		Password:  *registry.Password,
	}
	err := r.storage.ExecTx(ctx, func(s *storage.Storage) error {
		id, err := s.ContainerRegistryRepository().CreateNew(entity)
		if err != nil {
			return err
		}
		registry.Id = &id

		return r.syncProjectSecret(ctx, s, project, auth)
	})
	if err != nil {
		return openapi.ContainerRegistry{}, errors.Wrap(err, "failed to create container registry")
	}
	log.Infof("Added container registry %d (%s) to project %s", registry.Id, registry.Url, project)
	return registry, nil
}

func (r containerRegistriesImpl) DeleteContainerRegistry(ctx context.Context, project string, id int, auth middleware.Authentication) error {
	if err := r.projects.checkAccess(project, auth); err != nil {
		return err
	}
	err := r.storage.ExecTx(ctx, func(s *storage.Storage) error {
		if err := s.ContainerRegistryRepository().Delete(id); err != nil {
			return err
		}

		return r.syncProjectSecret(ctx, s, project, auth)
	})
	if err != nil {
		return errors.Wrap(err, "failed to delete container registry")
	}
	log.Infof("Deleted container registry %d from project %s", id, project)
	return nil
}

func (r containerRegistriesImpl) syncKubernetes(ctx context.Context, projectId string) error {
	return r.syncProjectSecret(ctx, r.storage, projectId, middleware.ServiceAccount)
}

func (r containerRegistriesImpl) syncProjectSecret(
	ctx context.Context,
	s *storage.Storage,
	project string,
	auth middleware.Authentication,
) error {
	regs, err := r.getProjectContainerRegistries(s, project, true, auth)
	if err != nil {
		return err
	}
	return r.createRegistriesSecret(ctx, project, regs)
}

func (r containerRegistriesImpl) getProjectContainerRegistries(
	s *storage.Storage,
	project string,
	withPwd bool,
	auth middleware.Authentication,
) ([]openapi.ContainerRegistry, error) {
	if err := r.projects.checkAccess(project, auth); err != nil {
		return nil, err
	}
	entities, err := s.ContainerRegistryRepository().FindByProjectId(project)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get project container registries")
	}
	registries := mapItems(entities, func(e storage.ContainerRegistryEntity) openapi.ContainerRegistry {
		var pwd *string = nil
		if withPwd {
			pwd = &e.Password
		}
		return openapi.ContainerRegistry{
			Id:       &e.Id,
			Url:      e.Url,
			Username: e.Username,
			Password: pwd,
		}
	})
	return registries, nil
}

func (r containerRegistriesImpl) createRegistriesSecret(ctx context.Context, project string, registries []openapi.ContainerRegistry) error {
	type regAuth struct {
		Auth string `json:"auth"`
	}
	type dockerConfig struct {
		Auths map[string]regAuth `json:"auths"`
	}
	auths := make(map[string]regAuth)
	for _, r := range registries {
		auth := base64.StdEncoding.EncodeToString([]byte(r.Username + ":" + *r.Password))
		auths[r.Url] = regAuth{Auth: auth}
	}
	configJson, err := json.Marshal(dockerConfig{Auths: auths})
	if err != nil {
		return errors.Wrap(err, "failed to serialize dockerconfigjson")
	}

	secret := applyConfigsCoreV1.Secret(regcredSecretName, project).
		WithType("kubernetes.io/dockerconfigjson").
		WithData(map[string][]byte{".dockerconfigjson": configJson})

	_, err = r.clientset.CoreV1().Secrets(project).Apply(ctx, secret, metav1.ApplyOptions{FieldManager: "letsdeploy"})
	if err != nil {
		return errors.Wrap(err, "failed to apply registry auth secret")
	}
	log.Debugf("Applied registry secret configuration to project %s", project)
	return nil
}

func isRegistryUrlValid(regUrl string) bool {
	if !strings.HasPrefix(regUrl, "http:") && !strings.HasPrefix(regUrl, "https:") {
		regUrl = "https://" + regUrl
	}
	_, err := url.Parse(regUrl)
	return err == nil
}
