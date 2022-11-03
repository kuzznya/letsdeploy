package core

import (
	"context"
	"github.com/kuzznya/letsdeploy/app/apperrors"
	"github.com/kuzznya/letsdeploy/app/storage"
	"github.com/kuzznya/letsdeploy/internal/openapi"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	applyConfigsAppsV1 "k8s.io/client-go/applyconfigurations/apps/v1"
	applyConfigsCoreV1 "k8s.io/client-go/applyconfigurations/core/v1"
	applyConfigsMetaV1 "k8s.io/client-go/applyconfigurations/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type managedServiceParams struct {
	Image string
	Ports []int32
}

var managedServices = map[openapi.ManagedServiceType]managedServiceParams{
	openapi.Postgres: {Image: "postgres:14", Ports: []int32{5432}},
	openapi.Mysql:    {Image: "mysql:8.0", Ports: []int32{3306}},
}

type ManagedServices interface {
	GetProjectManagedServices(project string, requester string) ([]openapi.ManagedService, error)
	CreateManagedService(ctx context.Context, service openapi.ManagedService, author string) (*openapi.ManagedService, error)
	GetManagedService(id int, requester string) (*openapi.ManagedService, error)
	DeleteManagedService(ctx context.Context, id int, requester string) error
}

type managedServicesImpl struct {
	projects  Projects
	storage   *storage.Storage
	clientset *kubernetes.Clientset
}

func InitManagedServices(
	projects Projects,
	storage *storage.Storage,
	clientset *kubernetes.Clientset,
) ManagedServices {
	return &managedServicesImpl{projects: projects, storage: storage, clientset: clientset}
}

func (m managedServicesImpl) GetProjectManagedServices(project string, requester string) ([]openapi.ManagedService, error) {
	if err := m.projects.checkAccess(project, requester); err != nil {
		return nil, err
	}
	entities, err := m.storage.ManagedServiceRepository().FindByProjectId(project)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get managed services of a project")
	}
	services := make([]openapi.ManagedService, len(entities))
	for i, entity := range entities {
		services[i] = openapi.ManagedService{
			Id:      &entity.Id,
			Name:    entity.Name,
			Project: entity.ProjectId,
			Type:    openapi.ManagedServiceType(entity.Type),
		}
	}
	return services, nil
}

func (m managedServicesImpl) CreateManagedService(ctx context.Context, service openapi.ManagedService, author string) (*openapi.ManagedService, error) {
	if err := m.projects.checkAccess(service.Project, author); err != nil {
		return nil, err
	}
	entity := storage.ManagedServiceEntity{ProjectId: service.Project, Name: service.Name, Type: string(service.Type)}
	err := m.storage.ExecTx(ctx, func(s *storage.Storage) error {
		id, err := s.ManagedServiceRepository().CreateNew(entity)
		if err != nil {
			return err
		}
		service.Id = &id

		err = m.createManagedServiceDeployment(ctx, service)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create managed service")
	}
	return &service, nil
}

func (m managedServicesImpl) GetManagedService(id int, requester string) (*openapi.ManagedService, error) {
	entity, err := m.storage.ManagedServiceRepository().FindByID(id)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get managed service by id")
	}
	if err := m.projects.checkAccess(entity.ProjectId, requester); err != nil {
		return nil, err
	}
	return &openapi.ManagedService{
		Id:      &entity.Id,
		Name:    entity.Name,
		Project: entity.ProjectId,
		Type:    openapi.ManagedServiceType(entity.Type),
	}, nil
}

func (m managedServicesImpl) DeleteManagedService(ctx context.Context, id int, requester string) error {
	entity, err := m.storage.ManagedServiceRepository().FindByID(id)
	if apperrors.IsNotFound(err) {
		return nil
	} else if err != nil {
		return errors.Wrap(err, "failed to get managed service by id")
	}
	err = m.storage.ExecTx(ctx, func(s *storage.Storage) error {
		err := s.ManagedServiceRepository().Delete(id)
		if err != nil {
			return err
		}
		err = m.deleteManagedServiceDeployment(ctx, entity.ProjectId, entity.Name)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return errors.Wrap(err, "failed to delete managed service")
	}
	return nil
}

func (m managedServicesImpl) createManagedServiceDeployment(ctx context.Context, service openapi.ManagedService) error {
	port := applyConfigsCoreV1.ServicePort().WithPort(80).WithTargetPort(intstr.FromInt(80)) // TODO set real target port
	serviceConfig := applyConfigsCoreV1.Service(service.Name, service.Project).
		WithSpec(applyConfigsCoreV1.ServiceSpec().WithPorts(port))
	_, err := m.clientset.CoreV1().Services(service.Project).Apply(ctx, serviceConfig, metav1.ApplyOptions{})
	if err != nil {
		return errors.Wrap(err, "failed to create K8s service for managed service")
	}

	// TODO form real spec
	container := applyConfigsCoreV1.Container().
		WithName(containerName).WithImage("")
	podTemplate := applyConfigsCoreV1.PodTemplateSpec().
		WithLabels(map[string]string{"app": service.Name}).
		WithSpec(applyConfigsCoreV1.PodSpec().WithContainers(container).WithTerminationGracePeriodSeconds(10))
	statefulSet := applyConfigsAppsV1.StatefulSet(service.Name, service.Project).
		WithSpec(applyConfigsAppsV1.StatefulSetSpec().
			WithSelector(applyConfigsMetaV1.LabelSelector().
				WithMatchLabels(map[string]string{"app": service.Name})).
			WithServiceName(service.Name).
			WithTemplate(podTemplate))
	statefulSet.Labels["letsdeploy.space/managed"] = "true"

	switch service.Type {
	case openapi.Postgres:
		return m.createPostgresDeployment(ctx, service)
	case openapi.Mysql:
		return m.createMySqlDeployment(ctx, service)
	case openapi.Redis:
		return m.createRedisDeployment(ctx, service)
	case openapi.Rabbitmq:
		return m.createRabbitMQDeployment(ctx, service)
	default:
		return errors.Errorf("Unknown managed service type %s", service.Type)
	}
}

func (m managedServicesImpl) createPostgresDeployment(ctx context.Context, service openapi.ManagedService) error {
	// TODO implement
	panic("not implemented")
}

func (m managedServicesImpl) createMySqlDeployment(ctx context.Context, service openapi.ManagedService) error {
	// TODO implement
	panic("not implemented")
}

func (m managedServicesImpl) createRedisDeployment(ctx context.Context, service openapi.ManagedService) error {
	// TODO implement
	panic("not implemented")
}

func (m managedServicesImpl) createRabbitMQDeployment(ctx context.Context, service openapi.ManagedService) error {
	// TODO implement
	panic("not implemented")
}

func (m managedServicesImpl) deleteManagedServiceDeployment(ctx context.Context, namespace string, name string) error {
	err := m.clientset.AppsV1().StatefulSets(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return errors.Wrap(err, "failed to delete managed service StatefulSet")
	}
	return nil
}
