package core

import (
	"context"
	"github.com/kuzznya/letsdeploy/app/apperrors"
	"github.com/kuzznya/letsdeploy/app/middleware"
	"github.com/kuzznya/letsdeploy/app/storage"
	"github.com/kuzznya/letsdeploy/internal/openapi"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	applyConfigsAppsV1 "k8s.io/client-go/applyconfigurations/apps/v1"
	applyConfigsCoreV1 "k8s.io/client-go/applyconfigurations/core/v1"
	applyConfigsMetaV1 "k8s.io/client-go/applyconfigurations/meta/v1"
	"k8s.io/client-go/kubernetes"
	"math/rand"
)

type managedServiceParams struct {
	image    string
	username string
	podPort  int
}

var managedServices = map[openapi.ManagedServiceType]managedServiceParams{
	openapi.Postgres: {image: "postgres:14", username: "postgres", podPort: 5432},
	openapi.Mysql:    {image: "mysql:8", username: "root", podPort: 3306},
	openapi.Redis:    {image: "redis:7", username: "", podPort: 6379},
	openapi.Rabbitmq: {image: "rabbitmq:3-management", username: "guest", podPort: 5672},
}

type ManagedServices interface {
	projectSynchronizable
	GetProjectManagedServices(project string, auth middleware.Authentication) ([]openapi.ManagedService, error)
	CreateManagedService(ctx context.Context, service openapi.ManagedService, auth middleware.Authentication) (*openapi.ManagedService, error)
	GetManagedService(id int, auth middleware.Authentication) (*openapi.ManagedService, error)
	DeleteManagedService(ctx context.Context, id int, auth middleware.Authentication) error
}

type managedServicesImpl struct {
	projects  Projects
	storage   *storage.Storage
	clientset *kubernetes.Clientset
}

var _ ManagedServices = (*managedServicesImpl)(nil)

func InitManagedServices(
	projects Projects,
	storage *storage.Storage,
	clientset *kubernetes.Clientset,
) ManagedServices {
	return &managedServicesImpl{projects: projects, storage: storage, clientset: clientset}
}

func (m managedServicesImpl) GetProjectManagedServices(project string, auth middleware.Authentication) ([]openapi.ManagedService, error) {
	if err := m.projects.checkAccess(project, auth); err != nil {
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

func (m managedServicesImpl) CreateManagedService(ctx context.Context, service openapi.ManagedService, auth middleware.Authentication) (*openapi.ManagedService, error) {
	if err := m.projects.checkAccess(service.Project, auth); err != nil {
		return nil, err
	}
	entity := storage.ManagedServiceEntity{ProjectId: service.Project, Name: service.Name, Type: string(service.Type)}
	err := m.storage.ExecTx(ctx, func(s *storage.Storage) error {
		id, err := s.ManagedServiceRepository().CreateNew(entity)
		if err != nil {
			return err
		}
		service.Id = &id

		err = m.createManagedServiceDeployment(ctx, s, service)
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

func (m managedServicesImpl) GetManagedService(id int, auth middleware.Authentication) (*openapi.ManagedService, error) {
	entity, err := m.storage.ManagedServiceRepository().FindByID(id)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get managed service by id")
	}
	if err := m.projects.checkAccess(entity.ProjectId, auth); err != nil {
		return nil, err
	}
	return &openapi.ManagedService{
		Id:      &entity.Id,
		Name:    entity.Name,
		Project: entity.ProjectId,
		Type:    openapi.ManagedServiceType(entity.Type),
	}, nil
}

func (m managedServicesImpl) DeleteManagedService(ctx context.Context, id int, auth middleware.Authentication) error {
	entity, err := m.storage.ManagedServiceRepository().FindByID(id)
	if apperrors.IsNotFound(err) {
		return nil
	} else if err != nil {
		return errors.Wrap(err, "failed to get managed service by id")
	}
	if err := m.projects.checkAccess(entity.ProjectId, auth); err != nil {
		return err
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

func (m managedServicesImpl) createManagedServiceDeployment(ctx context.Context, store *storage.Storage, service openapi.ManagedService) error {
	err := m.createK8sService(ctx, service)
	if err != nil {
		return errors.Wrap(err, "failed to create K8s Service for managed service")
	}
	err = m.createPasswordSecret(ctx, store, service)
	if err != nil {
		return errors.Wrap(err, "failed to create secret for managed service")
	}
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

func (m managedServicesImpl) createK8sService(ctx context.Context, service openapi.ManagedService) error {
	port := applyConfigsCoreV1.ServicePort().
		WithPort(80).
		WithTargetPort(intstr.FromInt(managedServices[service.Type].podPort))
	serviceConfig := applyConfigsCoreV1.Service(service.Name, service.Project).
		WithSpec(applyConfigsCoreV1.ServiceSpec().WithPorts(port))
	_, err := m.clientset.CoreV1().Services(service.Project).Apply(ctx, serviceConfig, metav1.ApplyOptions{FieldManager: "letsdeploy"})
	if err != nil {
		return errors.Wrap(err, "failed to create K8s service for managed service")
	}
	return nil
}

func (m managedServicesImpl) createPasswordSecret(ctx context.Context, store *storage.Storage, service openapi.ManagedService) error {
	exists, err := store.SecretRepository().ExistsByProjectIdAndName(service.Project, service.Name+"-password")
	if err != nil {
		return errors.Wrap(err, "failed to check if password secret already exists")
	}
	if exists {
		return nil
	}

	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	passwordLen := 16
	b := make([]rune, passwordLen)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	password := string(b)

	secret := applyConfigsCoreV1.Secret(service.Name+"-password", service.Project).
		WithLabels(map[string]string{"letsdeploy.space/managed": "true"}).
		WithStringData(map[string]string{secretKey: password})

	secretEntity := storage.SecretEntity{
		ProjectId:        service.Project,
		Name:             service.Name + "-password",
		Value:            password,
		ManagedServiceId: service.Id,
	}

	err = store.ExecTx(ctx, func(s *storage.Storage) error {
		err := s.SecretRepository().CreateNew(secretEntity)
		if err != nil {
			return err
		}

		_, err = m.clientset.CoreV1().Secrets(service.Project).Apply(ctx, secret, metav1.ApplyOptions{FieldManager: "letsdeploy"})
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return errors.Wrap(err, "failed to create secret for managed service")
	}
	return nil
}

func (m managedServicesImpl) createPasswordEnvVarSource(service openapi.ManagedService) *applyConfigsCoreV1.EnvVarSourceApplyConfiguration {
	return applyConfigsCoreV1.EnvVarSource().
		WithSecretKeyRef(applyConfigsCoreV1.SecretKeySelector().
			WithName(service.Name + "-password").
			WithKey(secretKey))
}

func (m managedServicesImpl) createPostgresDeployment(ctx context.Context, service openapi.ManagedService) error {
	container := applyConfigsCoreV1.Container().
		WithName(containerName).
		WithImage(managedServices[service.Type].image).
		WithPorts(applyConfigsCoreV1.ContainerPort().WithContainerPort(int32(managedServices[service.Type].podPort))).
		WithVolumeMounts(applyConfigsCoreV1.VolumeMount().WithName("data").WithMountPath("/var/lib/postgresql")).
		WithEnv(applyConfigsCoreV1.EnvVar().
			WithName("POSTGRES_PASSWORD").
			WithValueFrom(m.createPasswordEnvVarSource(service)))

	podTemplate := applyConfigsCoreV1.PodTemplateSpec().
		WithLabels(map[string]string{"app": service.Name}).
		WithSpec(applyConfigsCoreV1.PodSpec().WithContainers(container).WithTerminationGracePeriodSeconds(10))

	pvClaim := applyConfigsCoreV1.PersistentVolumeClaim("data", service.Project).
		WithSpec(applyConfigsCoreV1.PersistentVolumeClaimSpec().
			WithAccessModes(v1.ReadWriteOnce).
			WithResources(applyConfigsCoreV1.ResourceRequirements().
				WithRequests(v1.ResourceList{v1.ResourceStorage: resource.MustParse("1Gi")})))

	statefulSet := applyConfigsAppsV1.StatefulSet(service.Name, service.Project).
		WithLabels(map[string]string{"letsdeploy.space/managed": "true"}).
		WithSpec(applyConfigsAppsV1.StatefulSetSpec().
			WithSelector(applyConfigsMetaV1.LabelSelector().
				WithMatchLabels(map[string]string{"app": service.Name})).
			WithServiceName(service.Name).
			WithTemplate(podTemplate).
			WithVolumeClaimTemplates(pvClaim))

	_, err := m.clientset.AppsV1().StatefulSets(service.Project).Apply(ctx, statefulSet, metav1.ApplyOptions{FieldManager: "letsdeploy"})
	if err != nil {
		return errors.Wrap(err, "failed to create K8s deployment for managed service")
	}
	return nil
}

func (m managedServicesImpl) createMySqlDeployment(ctx context.Context, service openapi.ManagedService) error {
	container := applyConfigsCoreV1.Container().
		WithName(containerName).
		WithImage(managedServices[service.Type].image).
		WithPorts(applyConfigsCoreV1.ContainerPort().WithContainerPort(int32(managedServices[service.Type].podPort))).
		WithVolumeMounts(applyConfigsCoreV1.VolumeMount().WithName("data").WithMountPath("/var/lib/postgresql")).
		WithEnv(applyConfigsCoreV1.EnvVar().
			WithName("MYSQL_ROOT_PASSWORD").
			WithValueFrom(m.createPasswordEnvVarSource(service)),
			applyConfigsCoreV1.EnvVar().
				WithName("MYSQL_DATABASE").
				WithValue("db"))

	podTemplate := applyConfigsCoreV1.PodTemplateSpec().
		WithLabels(map[string]string{"app": service.Name}).
		WithSpec(applyConfigsCoreV1.PodSpec().WithContainers(container).WithTerminationGracePeriodSeconds(10))

	pvClaim := applyConfigsCoreV1.PersistentVolumeClaim("data", service.Project).
		WithSpec(applyConfigsCoreV1.PersistentVolumeClaimSpec().
			WithAccessModes(v1.ReadWriteOnce).
			WithResources(applyConfigsCoreV1.ResourceRequirements().
				WithRequests(v1.ResourceList{v1.ResourceStorage: resource.MustParse("1Gi")})))

	statefulSet := applyConfigsAppsV1.StatefulSet(service.Name, service.Project).
		WithLabels(map[string]string{"letsdeploy.space/managed": "true"}).
		WithSpec(applyConfigsAppsV1.StatefulSetSpec().
			WithSelector(applyConfigsMetaV1.LabelSelector().
				WithMatchLabels(map[string]string{"app": service.Name})).
			WithServiceName(service.Name).
			WithTemplate(podTemplate).
			WithVolumeClaimTemplates(pvClaim))

	_, err := m.clientset.AppsV1().StatefulSets(service.Project).Apply(ctx, statefulSet, metav1.ApplyOptions{FieldManager: "letsdeploy"})
	if err != nil {
		return errors.Wrap(err, "failed to create K8s deployment for managed service")
	}
	return nil
}

func (m managedServicesImpl) createRedisDeployment(ctx context.Context, service openapi.ManagedService) error {
	container := applyConfigsCoreV1.Container().
		WithName(containerName).
		WithImage(managedServices[service.Type].image).
		WithPorts(applyConfigsCoreV1.ContainerPort().WithContainerPort(int32(managedServices[service.Type].podPort))).
		WithVolumeMounts(applyConfigsCoreV1.VolumeMount().WithName("data").WithMountPath("/data")).
		WithCommand("/bin/sh", "-c", "redis-server --appendonly yes --requirepass ${REDIS_PASSWORD}").
		WithEnv(applyConfigsCoreV1.EnvVar().
			WithName("REDIS_PASSWORD").
			WithValueFrom(m.createPasswordEnvVarSource(service)))

	podTemplate := applyConfigsCoreV1.PodTemplateSpec().
		WithLabels(map[string]string{"app": service.Name}).
		WithSpec(applyConfigsCoreV1.PodSpec().WithContainers(container).WithTerminationGracePeriodSeconds(10))

	pvClaim := applyConfigsCoreV1.PersistentVolumeClaim("data", service.Project).
		WithSpec(applyConfigsCoreV1.PersistentVolumeClaimSpec().
			WithAccessModes(v1.ReadWriteOnce).
			WithResources(applyConfigsCoreV1.ResourceRequirements().
				WithRequests(v1.ResourceList{v1.ResourceStorage: resource.MustParse("500Mi")})))

	statefulSet := applyConfigsAppsV1.StatefulSet(service.Name, service.Project).
		WithLabels(map[string]string{"letsdeploy.space/managed": "true"}).
		WithSpec(applyConfigsAppsV1.StatefulSetSpec().
			WithSelector(applyConfigsMetaV1.LabelSelector().
				WithMatchLabels(map[string]string{"app": service.Name})).
			WithServiceName(service.Name).
			WithTemplate(podTemplate).
			WithVolumeClaimTemplates(pvClaim))

	_, err := m.clientset.AppsV1().StatefulSets(service.Project).Apply(ctx, statefulSet, metav1.ApplyOptions{FieldManager: "letsdeploy"})
	if err != nil {
		return errors.Wrap(err, "failed to create K8s deployment for managed service")
	}
	return nil
}

func (m managedServicesImpl) createRabbitMQDeployment(ctx context.Context, service openapi.ManagedService) error {
	container := applyConfigsCoreV1.Container().
		WithName(containerName).
		WithImage(managedServices[service.Type].image).
		WithPorts(
			applyConfigsCoreV1.ContainerPort().WithContainerPort(5672).WithName("amqp"),
			applyConfigsCoreV1.ContainerPort().WithContainerPort(15672).WithName("http"),
			applyConfigsCoreV1.ContainerPort().WithContainerPort(4369)).
		WithEnv(
			applyConfigsCoreV1.EnvVar().
				WithName("HOSTNAME").
				WithValueFrom(applyConfigsCoreV1.EnvVarSource().WithFieldRef(
					applyConfigsCoreV1.ObjectFieldSelector().WithFieldPath("metadata.name"))),
			applyConfigsCoreV1.EnvVar().
				WithName("NODE_NAME").
				WithValueFrom(applyConfigsCoreV1.EnvVarSource().WithFieldRef(
					applyConfigsCoreV1.ObjectFieldSelector().WithFieldPath("metadata.name"))),
			applyConfigsCoreV1.EnvVar().
				WithName("NAMESPACE").
				WithValueFrom(applyConfigsCoreV1.EnvVarSource().WithFieldRef(
					applyConfigsCoreV1.ObjectFieldSelector().WithFieldPath("metadata.namespace"))),
			applyConfigsCoreV1.EnvVar().WithName("RABBITMQ_USE_LONGNAME").WithValue("true"),
			applyConfigsCoreV1.EnvVar().WithName("RABBITMQ_NODENAME").
				WithValue("rabbit@$(HOSTNAME).rabbitmq.$(NAMESPACE).svc.cluster.local"),
			applyConfigsCoreV1.EnvVar().WithName("RABBITMQ_DEFAULT_USER").
				WithValue(managedServices[service.Type].username),
			applyConfigsCoreV1.EnvVar().WithName("RABBITMQ_DEFAULT_PASS").
				WithValueFrom(m.createPasswordEnvVarSource(service)),
			applyConfigsCoreV1.EnvVar().WithName("RABBITMQ_ERLANG_COOKIE").
				WithValue("secret_cookie_12345678"), // TODO refactor
		).
		WithLivenessProbe(applyConfigsCoreV1.Probe().
			WithExec(applyConfigsCoreV1.ExecAction().WithCommand("rabbitmq-diagnostics", "status")).
			WithInitialDelaySeconds(20).
			WithPeriodSeconds(60).
			WithTimeoutSeconds(15).
			WithFailureThreshold(3)).
		WithReadinessProbe(applyConfigsCoreV1.Probe().
			WithExec(applyConfigsCoreV1.ExecAction().WithCommand("rabbitmq-diagnostics", "ping")).
			WithInitialDelaySeconds(30).
			WithPeriodSeconds(60).
			WithTimeoutSeconds(10).
			WithFailureThreshold(3))

	podTemplate := applyConfigsCoreV1.PodTemplateSpec().
		WithLabels(map[string]string{"app": service.Name}).
		WithSpec(applyConfigsCoreV1.PodSpec().WithContainers(container).WithTerminationGracePeriodSeconds(10))

	statefulSet := applyConfigsAppsV1.StatefulSet(service.Name, service.Project).
		WithLabels(map[string]string{"letsdeploy.space/managed": "true"}).
		WithSpec(applyConfigsAppsV1.StatefulSetSpec().
			WithSelector(applyConfigsMetaV1.LabelSelector().
				WithMatchLabels(map[string]string{"app": service.Name})).
			WithServiceName(service.Name).
			WithTemplate(podTemplate))

	_, err := m.clientset.AppsV1().StatefulSets(service.Project).Apply(ctx, statefulSet, metav1.ApplyOptions{FieldManager: "letsdeploy"})
	if err != nil {
		return errors.Wrap(err, "failed to create K8s deployment for managed service")
	}
	return nil
}

func (m managedServicesImpl) deleteManagedServiceDeployment(ctx context.Context, namespace string, name string) error {
	err := m.clientset.AppsV1().StatefulSets(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return errors.Wrap(err, "failed to delete managed service StatefulSet")
	}
	return nil
}

func (m managedServicesImpl) syncKubernetes(ctx context.Context, projectId string) error {
	services, err := m.GetProjectManagedServices(projectId, middleware.ServiceAccount)
	if err != nil {
		return errors.Wrap(err, "failed to get project managed services")
	}
	servicesMap := toMapSelf(services, func(service openapi.ManagedService) string { return service.Name })
	for _, service := range services {
		err := m.createManagedServiceDeployment(ctx, m.storage, service)
		if err != nil {
			log.WithError(err).Errorf("Failed to create managed service deployment %s, skipping\n", service.Name)
		}
	}
	statefulSets, err := m.clientset.AppsV1().StatefulSets(projectId).List(ctx, metav1.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "failed to get statefulsets list")
	}
	for _, statefulSet := range statefulSets.Items {
		if !contains(servicesMap, statefulSet.Name) && statefulSet.Labels["letsdeploy.space/managed"] == "true" {
			err := m.deleteManagedServiceDeployment(ctx, projectId, statefulSet.Name)
			if err != nil {
				log.WithError(err).Errorf("Failed to delete managed service statefulset %s, skipping\n", statefulSet.Name)
			}
		}
	}
	return nil
}
