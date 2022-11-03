package core

import (
	"context"
	"github.com/kuzznya/letsdeploy/app/apperrors"
	"github.com/kuzznya/letsdeploy/app/storage"
	"github.com/kuzznya/letsdeploy/internal/openapi"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	applyConfigsAppsV1 "k8s.io/client-go/applyconfigurations/apps/v1"
	applyConfigsCoreV1 "k8s.io/client-go/applyconfigurations/core/v1"
	applyConfigsMetaV1 "k8s.io/client-go/applyconfigurations/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const containerName = "container-0"

type Services interface {
	GetProjectServices(project string, requester string) ([]openapi.Service, error)
	CreateService(ctx context.Context, service openapi.Service, author string) (*openapi.Service, error)
	GetService(id int, requester string) (*openapi.Service, error)
	UpdateService(ctx context.Context, service openapi.Service, requester string) (*openapi.Service, error)
	DeleteService(ctx context.Context, id int, requester string) error
}

type servicesImpl struct {
	projects  Projects
	storage   *storage.Storage
	clientset *kubernetes.Clientset
}

func InitServices(
	projects Projects,
	storage *storage.Storage,
	clientset *kubernetes.Clientset,
) Services {
	s := servicesImpl{projects: projects, storage: storage, clientset: clientset}
	return &s
}

func (s servicesImpl) GetProjectServices(project string, requester string) ([]openapi.Service, error) {
	if err := s.projects.checkAccess(project, requester); err != nil {
		return nil, err
	}
	entities, err := s.storage.ServiceRepository().FindByProjectId(project)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get project services")
	}
	services := make([]openapi.Service, len(entities))
	for i, entity := range entities {
		services[i] = openapi.Service{
			Id:      &entity.Id,
			Image:   entity.Image,
			Name:    entity.Name,
			Port:    entity.Port,
			Project: entity.ProjectId,
		}
	}
	return services, nil
}

func (s servicesImpl) CreateService(ctx context.Context, service openapi.Service, author string) (*openapi.Service, error) {
	if err := s.projects.checkAccess(service.Project, author); err != nil {
		return nil, err
	}
	record := storage.ServiceEntity{Name: service.Name, Image: service.Image, Port: service.Port}
	err := s.storage.ExecTx(ctx, func(store *storage.Storage) error {
		id, err := store.ServiceRepository().CreateNew(record)
		if err != nil {
			return err
		}
		service.Id = &id

		err = s.createServiceDeployment(ctx, service)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new service")
	}
	return &service, nil
}

func (s servicesImpl) GetService(id int, requester string) (*openapi.Service, error) {
	entity, err := s.storage.ServiceRepository().FindByID(id)
	if err != nil {
		return nil, apperrors.WrapNonAppError(err, "failed to get service by id")
	}
	if err := s.projects.checkAccess(entity.ProjectId, requester); err != nil {
		return nil, err
	}
	return &openapi.Service{
		Id:      &entity.Id,
		Image:   entity.Image,
		Name:    entity.Name,
		Port:    entity.Port,
		Project: entity.ProjectId,
	}, nil
}

func (s servicesImpl) UpdateService(ctx context.Context, service openapi.Service, requester string) (*openapi.Service, error) {
	entity, err := s.storage.ServiceRepository().FindByID(*service.Id)
	if err != nil {
		return nil, apperrors.WrapNonAppError(err, "failed to get service by id")
	}
	if err := s.projects.checkAccess(entity.ProjectId, requester); err != nil {
		return nil, err
	}
	if entity.ProjectId != service.Project {
		return nil, apperrors.BadRequest("Project field cannot be updated")
	}
	updated := storage.ServiceEntity{
		Id:        *service.Id,
		ProjectId: entity.ProjectId,
		Name:      service.Name,
		Image:     service.Image,
		Port:      service.Port,
	}
	err = s.storage.ExecTx(ctx, func(store *storage.Storage) error {
		err := store.ServiceRepository().Update(updated)
		if err != nil {
			return err
		}

		err = s.createServiceDeployment(ctx, service)
		if err != nil {
			return err
		}

		if updated.Name != entity.Name {
			err := s.deleteServiceDeployment(ctx, entity.ProjectId, entity.Name)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to update service")
	}
	result := openapi.Service{
		Id:      &updated.Id,
		Image:   updated.Image,
		Name:    updated.Name,
		Port:    updated.Port,
		Project: updated.ProjectId,
	}
	return &result, nil
}

func (s servicesImpl) DeleteService(ctx context.Context, id int, requester string) error {
	service, err := s.GetService(id, requester)
	if apperrors.IsNotFound(err) {
		return nil
	} else if err != nil {
		return errors.Wrap(err, "failed to get service by id")
	}
	err = s.storage.ExecTx(ctx, func(store *storage.Storage) error {
		err = store.ServiceRepository().Delete(id)
		if err != nil {
			return err
		}
		err := s.deleteServiceDeployment(ctx, service.Project, service.Name)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return errors.Wrap(err, "failed to delete service")
	}
	return nil
}

func (s servicesImpl) createServiceDeployment(ctx context.Context, service openapi.Service) error {
	limits := v1.ResourceList{}
	limits.Cpu().SetMilli(250)
	limits.Memory().SetScaled(512, resource.Mega)
	container := applyConfigsCoreV1.Container().
		WithName(containerName).
		WithImage(service.Image).
		WithImagePullPolicy(v1.PullAlways).
		WithPorts(applyConfigsCoreV1.ContainerPort().WithContainerPort(int32(service.Port))).
		WithResources(applyConfigsCoreV1.ResourceRequirements().WithLimits(limits))
	podTemplate := applyConfigsCoreV1.PodTemplateSpec().
		WithLabels(map[string]string{"app": service.Name}).
		WithSpec(applyConfigsCoreV1.PodSpec().WithContainers(container))
	deployment := applyConfigsAppsV1.Deployment(service.Name, "namespace").
		WithLabels(map[string]string{"app": service.Name}).
		WithSpec(applyConfigsAppsV1.DeploymentSpec().
			WithSelector(applyConfigsMetaV1.LabelSelector().
				WithMatchLabels(map[string]string{"app": service.Name})).
			WithTemplate(podTemplate))

	_, err := s.clientset.AppsV1().Deployments(service.Project).
		Apply(ctx, deployment, metav1.ApplyOptions{})
	if err != nil {
		return errors.Wrap(err, "failed to create service deployment")
	}
	return nil
}

func (s servicesImpl) deleteServiceDeployment(ctx context.Context, project string, service string) error {
	err := s.clientset.AppsV1().Deployments(project).Delete(ctx, service, metav1.DeleteOptions{})
	if err != nil {
		return errors.Wrap(err, "failed to delete service deployment")
	}
	return nil
}
