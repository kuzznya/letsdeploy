package core

import (
	"context"
	"fmt"
	"github.com/kuzznya/letsdeploy/app/apperrors"
	"github.com/kuzznya/letsdeploy/app/middleware"
	"github.com/kuzznya/letsdeploy/app/storage"
	"github.com/kuzznya/letsdeploy/app/util/promise"
	"github.com/kuzznya/letsdeploy/internal/openapi"
	"github.com/pkg/errors"
	"github.com/procyon-projects/chrono"
	log "github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	applyConfigsV1 "k8s.io/client-go/applyconfigurations/core/v1"
	"k8s.io/client-go/kubernetes"
	"strings"
)

const namespaceLabel = "letsdeploy.space/project-namespace"
const secretKey = "value"

type Projects interface {
	projectSynchronizable
	FindAll(limit int, offset int) ([]openapi.Project, error)
	CreateProject(ctx context.Context, project openapi.Project, auth middleware.Authentication) (*openapi.Project, error)
	GetProject(id string, auth middleware.Authentication) (*openapi.Project, error)
	GetProjectInfo(id string, auth middleware.Authentication) (*openapi.ProjectInfo, error)
	UpdateProject(project openapi.Project, auth middleware.Authentication) error
	DeleteProject(ctx context.Context, id string, auth middleware.Authentication) error
	GetUserProjects(auth middleware.Authentication) ([]openapi.Project, error)
	GetParticipants(id string, auth middleware.Authentication) ([]string, error)
	AddParticipant(id string, username string, auth middleware.Authentication) error
	RemoveParticipant(id string, username string, auth middleware.Authentication) error
	GetSecrets(projectId string, auth middleware.Authentication) ([]openapi.Secret, error)
	CreateSecret(ctx context.Context, projectId string, secret openapi.Secret, value string, auth middleware.Authentication) (*openapi.Secret, error)
	DeleteSecret(ctx context.Context, projectId string, name string, auth middleware.Authentication) error
	checkAccess(id string, auth middleware.Authentication) error
}

type projectsImpl struct {
	services        Services
	managedServices ManagedServices
	storage         *storage.Storage
	clientset       *kubernetes.Clientset
	scheduler       chrono.TaskScheduler
}

func InitProjects(
	storage *storage.Storage,
	clientset *kubernetes.Clientset,
	core promise.Promise[Core],
) Projects {
	p := &projectsImpl{storage: storage, clientset: clientset}
	core.OnProvided(func(core Core) {
		p.services = core.Services
		p.managedServices = core.ManagedServices
	})
	return p
}

func (p projectsImpl) FindAll(limit int, offset int) ([]openapi.Project, error) {
	entities, err := p.storage.ProjectRepository().FindAll(limit, offset)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve projects")
	}
	projects := make([]openapi.Project, len(entities))
	for i, entity := range entities {
		projects[i] = openapi.Project{
			Id: entity.Id,
		}
	}
	return projects, nil
}

func (p projectsImpl) CreateProject(ctx context.Context, project openapi.Project, auth middleware.Authentication) (*openapi.Project, error) {
	exists, err := p.storage.ProjectRepository().ExistsByID(project.Id)
	if err != nil {
		return nil, errors.Wrap(err, "cannot check if project with this name already exists")
	} else if exists {
		return nil, apperrors.BadRequest("project with this name already exists")
	}
	record := storage.ProjectEntity{Id: project.Id}
	err = p.storage.ExecTx(ctx, func(s *storage.Storage) error {
		id, err := s.ProjectRepository().CreateNew(record)
		if err != nil {
			return err
		}
		project.Id = id

		err = s.ProjectRepository().AddParticipant(id, auth.Username)
		if err != nil {
			return err
		}

		err = p.createProjectNamespace(ctx, project)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new project")
	}
	return &project, nil
}

func (p projectsImpl) GetProject(id string, auth middleware.Authentication) (*openapi.Project, error) {
	if err := p.checkAccess(id, auth); err != nil {
		return nil, err
	}
	record, err := p.storage.ProjectRepository().FindByID(id)
	if err != nil {
		return nil, apperrors.WrapNonAppError(err, "cannot find project by id")
	}
	return &openapi.Project{Id: record.Id}, nil
}

func (p projectsImpl) GetProjectInfo(id string, auth middleware.Authentication) (*openapi.ProjectInfo, error) {
	if err := p.checkAccess(id, auth); err != nil {
		return nil, err
	}
	record, err := p.storage.ProjectRepository().FindByID(id)
	if err != nil {
		return nil, apperrors.WrapNonAppError(err, "cannot find project by id")
	}
	participants, err := p.storage.ProjectRepository().GetParticipants(id)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve project participants")
	}
	fmt.Printf("services: %+v\n", p.services)
	services, err := p.services.GetProjectServices(id, auth)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve project services")
	}
	managedServices, err := p.managedServices.GetProjectManagedServices(id, auth)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve project managed services")
	}

	return &openapi.ProjectInfo{
		Id:              record.Id,
		Participants:    participants,
		Services:        services,
		ManagedServices: managedServices,
	}, nil
}

func (p projectsImpl) UpdateProject(project openapi.Project, auth middleware.Authentication) error {
	if err := p.checkAccess(project.Id, auth); err != nil {
		return err
	}
	record := storage.ProjectEntity{Id: project.Id}
	err := p.storage.ProjectRepository().Update(record)
	if err != nil {
		return errors.Wrap(err, "failed to update project")
	}
	return nil
}

func (p projectsImpl) DeleteProject(ctx context.Context, id string, auth middleware.Authentication) error {
	if err := p.checkAccess(id, auth); err != nil {
		return err
	}
	err := p.storage.ExecTx(ctx, func(s *storage.Storage) error {
		err := p.clientset.CoreV1().Namespaces().
			Delete(ctx, id, metav1.DeleteOptions{})
		if err != nil && !apierrors.IsNotFound(err) {
			return err
		}
		err = s.ProjectRepository().Delete(id)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return errors.Wrap(err, "failed to delete project")
	}
	return nil
}

func (p projectsImpl) GetUserProjects(auth middleware.Authentication) ([]openapi.Project, error) {
	projects, err := p.storage.ProjectRepository().FindUserProjects(auth.Username)
	if err != nil {
		return nil, errors.Wrap(err, "failed to find user's projects")
	}
	result := make([]openapi.Project, len(projects))
	for i, record := range projects {
		result[i] = openapi.Project{Id: record.Id}
	}
	return result, nil
}

func (p projectsImpl) GetParticipants(id string, auth middleware.Authentication) ([]string, error) {
	if err := p.checkAccess(id, auth); err != nil {
		return nil, err
	}
	participants, err := p.storage.ProjectRepository().GetParticipants(id)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get project participants")
	}
	return participants, nil
}

func (p projectsImpl) AddParticipant(id string, username string, auth middleware.Authentication) error {
	if err := p.checkAccess(id, auth); err != nil {
		return err
	}
	err := p.storage.ProjectRepository().AddParticipant(id, username)
	if err != nil {
		return errors.Wrap(err, "failed to add participant")
	}
	return nil
}

func (p projectsImpl) RemoveParticipant(id string, username string, auth middleware.Authentication) error {
	if err := p.checkAccess(id, auth); err != nil {
		return err
	}
	err := p.storage.ProjectRepository().RemoveParticipant(id, username)
	if err != nil {
		return errors.Wrap(err, "failed to add participant")
	}
	return nil
}

func (p projectsImpl) GetSecrets(projectId string, auth middleware.Authentication) ([]openapi.Secret, error) {
	if err := p.checkAccess(projectId, auth); err != nil {
		return nil, err
	}
	entities, err := p.storage.SecretRepository().FindByProjectId(projectId)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get project secrets")
	}
	secrets := make([]openapi.Secret, len(entities))
	for i, entity := range entities {
		secrets[i] = openapi.Secret{
			Name:             entity.Name,
			ManagedServiceId: entity.ManagedServiceId,
		}
	}
	return secrets, nil
}

func (p projectsImpl) CreateSecret(ctx context.Context, projectId string, secret openapi.Secret, value string, auth middleware.Authentication) (*openapi.Secret, error) {
	if err := p.checkAccess(projectId, auth); err != nil {
		return nil, err
	}
	exists, err := p.storage.SecretRepository().ExistsByName(secret.Name)
	if err != nil {
		return nil, errors.Wrap(err, "failed to check if secret already exists")
	}
	if exists {
		return nil, apperrors.BadRequest(fmt.Sprintf("Secret %s already exists in the project", secret.Name))
	}
	entity := storage.SecretEntity{
		ProjectId: projectId,
		Name:      secret.Name,
		Value:     value,
	}
	err = p.storage.ExecTx(ctx, func(s *storage.Storage) error {
		err := s.SecretRepository().CreateNew(entity)
		if err != nil {
			return err
		}

		config := applyConfigsV1.Secret(strings.ReplaceAll(strings.ToLower(secret.Name), "_", "-"), projectId).
			WithStringData(map[string]string{secretKey: value})
		_, err = p.clientset.CoreV1().Secrets(projectId).Apply(ctx, config, metav1.ApplyOptions{FieldManager: "letsdeploy"})
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new secret")
	}
	return &secret, nil
}

func (p projectsImpl) DeleteSecret(ctx context.Context, projectId string, name string, auth middleware.Authentication) error {
	if err := p.checkAccess(projectId, auth); err != nil {
		return err
	}
	secret, err := p.storage.SecretRepository().FindByName(name)
	if err != nil && !apperrors.IsNotFound(err) {
		return errors.Wrap(err, "failed to check access to existing secret")
	}
	if secret != nil && secret.ManagedServiceId != nil {
		return apperrors.Forbidden("Managed service password secret deletion is forbidden")
	}
	err = p.storage.SecretRepository().DeleteByName(name)
	if err != nil {
		return errors.Wrap(err, "failed to delete secret")
	}
	err = p.clientset.CoreV1().Secrets(projectId).Delete(ctx, strings.ReplaceAll(strings.ToLower(secret.Name), "_", "-"), metav1.DeleteOptions{})
	if err != nil {
		log.WithError(err).Warnln("Failed to delete secret from Kubernetes")
	}
	return nil
}

func (p projectsImpl) checkAccess(id string, auth middleware.Authentication) error {
	if auth == middleware.ServiceAccount {
		return nil
	}
	isParticipant, err := p.storage.ProjectRepository().IsParticipant(id, auth.Username)
	if err != nil {
		return errors.Wrap(err, "project access check unexpected failure")
	}
	if !isParticipant {
		return apperrors.NotFound(fmt.Sprintf("cannot find project with id %d", id))
	}
	return nil
}

func (p projectsImpl) createProjectNamespace(ctx context.Context, project openapi.Project) error {
	config := applyConfigsV1.Namespace(project.Id).WithLabels(map[string]string{namespaceLabel: "true"})
	_, err := p.clientset.CoreV1().Namespaces().Apply(ctx, config, metav1.ApplyOptions{FieldManager: "letsdeploy"})
	if err != nil {
		return err
	}
	log.Infof("Namespace %s created/updated for a project", project.Id)
	return nil
}

func (p projectsImpl) syncKubernetes(ctx context.Context, projectId string) error {
	err := p.createProjectNamespace(ctx, openapi.Project{Id: projectId})
	if err != nil {
		return errors.Wrap(err, "failed to create project namespace")
	}

	// TODO sync secrets

	log.Debugf("Project %s checked, namespace exists or was created", projectId)
	return nil
}

func (p projectsImpl) removeExcessNamespaces(ctx context.Context, checkedProjects map[string]bool) {
	namespaces, err := p.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		log.WithError(err).Errorln("Failed to retrieve namespaces")
		return
	}
	for _, namespace := range namespaces.Items {
		if namespace.Labels[namespaceLabel] != "true" {
			continue
		}
		if checkedProjects[namespace.Name] == false {
			err := p.clientset.CoreV1().Namespaces().Delete(ctx, namespace.Name, metav1.DeleteOptions{})
			if err != nil {
				log.WithError(err).Errorln("Failed to delete namespace without project, skipping")
				continue
			}
			log.Debugf("Namespace %s without corresponding project was deleted", namespace.Name)
		}
	}
}
