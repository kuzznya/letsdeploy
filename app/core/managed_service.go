package core

import (
	"context"
	"fmt"
	"github.com/kuzznya/letsdeploy/app/apperrors"
	"github.com/kuzznya/letsdeploy/app/middleware"
	"github.com/kuzznya/letsdeploy/app/storage"
	"github.com/kuzznya/letsdeploy/internal/openapi"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
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
	openapi.Postgres: {image: "postgres:15", username: "postgres", podPort: 5432},
	openapi.Mysql:    {image: "mysql:8", username: "root", podPort: 3306},
	openapi.Mongo:    {image: "mongo:6", username: "root", podPort: 27017},
	openapi.Redis:    {image: "redis:7", username: "", podPort: 6379},
	openapi.Rabbitmq: {image: "rabbitmq:3-management", username: "guest", podPort: 5672},
}

type ManagedServices interface {
	projectSynchronizable
	GetProjectManagedServices(project string, auth middleware.Authentication) ([]openapi.ManagedService, error)
	CreateManagedService(ctx context.Context, service openapi.ManagedService, auth middleware.Authentication) (*openapi.ManagedService, error)
	GetManagedService(id int, auth middleware.Authentication) (*openapi.ManagedService, error)
	DeleteManagedService(ctx context.Context, id int, auth middleware.Authentication) error
	GetManagedServiceStatus(ctx context.Context, id int, auth middleware.Authentication) (*openapi.ServiceStatus, error)
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
		id := entity.Id
		services[i] = openapi.ManagedService{
			Id:      &id,
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
	log.Infof("Created managed service %s in project %s", service.Name, service.Project)
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
	log.Infof("Deleted managed service %s in project %s", entity.Name, entity.ProjectId)
	return nil
}

func (m managedServicesImpl) GetManagedServiceStatus(ctx context.Context, id int, auth middleware.Authentication) (*openapi.ServiceStatus, error) {
	service, err := m.GetManagedService(id, auth)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get managed service")
	}

	set, err := m.clientset.AppsV1().StatefulSets(service.Project).Get(ctx, service.Name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get managed service stateful set")
	}

	if set.Generation > set.Status.ObservedGeneration {
		log.Debugf("Managed service %s generation is greater than observed generation, deployment is progressing", service.Name)
		return &openapi.ServiceStatus{Id: id, Status: openapi.Progressing}, nil
	}
	if set.Spec.Replicas != nil && set.Status.UpdatedReplicas < *set.Spec.Replicas {
		log.Debugf("Managed service %s updated replicas is less than expected, deployment is progressing", service.Name)
		return &openapi.ServiceStatus{Id: id, Status: openapi.Progressing}, nil
	}
	if set.Status.Replicas > set.Status.UpdatedReplicas {
		list, err := m.clientset.CoreV1().Pods(service.Project).List(ctx, metav1.ListOptions{LabelSelector: "app=" + service.Name})
		if err != nil {
			return nil, errors.Wrap(err, "failed to find a pod for managed service "+service.Name)
		}
		if len(list.Items) == 0 {
			return nil, apperrors.InternalServerError("failed to find a pod for managed service " + service.Name)
		}

		log.Debugf("Managed service %s old replicas are waiting termination", service.Name)

		newestPod := list.Items[0]
		for _, pod := range list.Items {
			if pod.CreationTimestamp.After(newestPod.CreationTimestamp.Time) {
				newestPod = pod
			}
		}

		if len(newestPod.Status.ContainerStatuses) == 0 {
			return &openapi.ServiceStatus{Id: id, Status: openapi.Progressing}, nil
		}

		podState := newestPod.Status.ContainerStatuses[0].State
		if podState.Waiting != nil && podState.Waiting.Reason == "CrashLoopBackOff" {
			log.Debugf("Managed service %s pod %s is unhealthy", service.Name, newestPod.Name)
			return &openapi.ServiceStatus{Id: id, Status: openapi.Unhealthy}, nil
		}

		return &openapi.ServiceStatus{Id: id, Status: openapi.Progressing}, nil
	}
	if set.Status.AvailableReplicas < set.Status.UpdatedReplicas {
		log.Debugf("Managed service %s %d of %d updated replicas are available",
			service.Name, set.Status.AvailableReplicas, set.Status.UpdatedReplicas)
		return &openapi.ServiceStatus{Id: id, Status: openapi.Progressing}, nil
	}
	return &openapi.ServiceStatus{Id: id, Status: openapi.Available}, nil
}

func (m managedServicesImpl) createManagedServiceDeployment(ctx context.Context, store *storage.Storage, service openapi.ManagedService) error {
	err := m.createK8sService(ctx, service)
	if err != nil {
		return errors.Wrap(err, "failed to create K8s Service for managed service")
	}
	err = m.createPasswordSecret(ctx, store, service)
	if err != nil {
		if err := m.deleteK8sService(ctx, service.Project, service.Name); err != nil {
			log.WithError(err).Errorln("Failed to delete K8s service after password secret creation failure")
		}
		return errors.Wrap(err, "failed to create secret for managed service")
	}
	switch service.Type {
	case openapi.Postgres:
		err = m.createPostgresDeployment(ctx, service)
	case openapi.Mysql:
		err = m.createMySqlDeployment(ctx, service)
	case openapi.Mongo:
		err = m.createMongoDeployment(ctx, service)
	case openapi.Redis:
		err = m.createRedisDeployment(ctx, service)
	case openapi.Rabbitmq:
		err = m.createRabbitMQDeployment(ctx, service)
	default:
		return errors.Errorf("Unknown managed service type %s", service.Type)
	}
	if err != nil {
		if err := m.deletePasswordSecret(ctx, service.Project, service.Name); err != nil {
			log.WithError(err).Errorln("Failed to delete password secret after managed service deployment failure")
		}
		if err := m.deleteK8sService(ctx, service.Project, service.Name); err != nil {
			log.WithError(err).Errorln("Failed to delete K8s service after managed service deployment failure")
		}
		return err
	}
	return nil
}

func (m managedServicesImpl) createK8sService(ctx context.Context, service openapi.ManagedService) error {
	port := applyConfigsCoreV1.ServicePort().
		WithPort(int32(managedServices[service.Type].podPort)).
		WithTargetPort(intstr.FromInt32(int32(managedServices[service.Type].podPort)))
	serviceConfig := applyConfigsCoreV1.Service(service.Name, service.Project).
		WithLabels(map[string]string{
			"letsdeploy.space/managed":      "true",
			"letsdeploy.space/service-type": "managed",
			"app":                           service.Name,
		}).
		WithSpec(applyConfigsCoreV1.ServiceSpec().WithPorts(port).
			WithSelector(map[string]string{"app": service.Name}))
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
	portArg := fmt.Sprintf("--port=%d", managedServices[service.Type].podPort)
	container := applyConfigsCoreV1.Container().
		WithName(containerName).
		WithImage(managedServices[service.Type].image).
		WithPorts(applyConfigsCoreV1.ContainerPort().WithContainerPort(int32(managedServices[service.Type].podPort))).
		WithVolumeMounts(applyConfigsCoreV1.VolumeMount().WithName("data").WithMountPath("/var/lib/postgresql")).
		WithEnv(applyConfigsCoreV1.EnvVar().
			WithName("POSTGRES_PASSWORD").
			WithValueFrom(m.createPasswordEnvVarSource(service))).
		WithLivenessProbe(applyConfigsCoreV1.Probe().
			WithExec(applyConfigsCoreV1.ExecAction().WithCommand("pg_isready", portArg)).
			WithInitialDelaySeconds(20).
			WithPeriodSeconds(20).
			WithTimeoutSeconds(5).
			WithFailureThreshold(3)).
		WithReadinessProbe(applyConfigsCoreV1.Probe().
			WithExec(applyConfigsCoreV1.ExecAction().WithCommand("pg_isready", portArg)).
			WithInitialDelaySeconds(30).
			WithPeriodSeconds(20).
			WithTimeoutSeconds(5).
			WithFailureThreshold(3))

	podTemplate := applyConfigsCoreV1.PodTemplateSpec().
		WithLabels(map[string]string{"app": service.Name}).
		WithSpec(applyConfigsCoreV1.PodSpec().WithContainers(container).WithTerminationGracePeriodSeconds(10))

	pvClaim := applyConfigsCoreV1.PersistentVolumeClaim("data", service.Project).
		WithSpec(applyConfigsCoreV1.PersistentVolumeClaimSpec().
			WithAccessModes(v1.ReadWriteOnce).
			WithResources(applyConfigsCoreV1.VolumeResourceRequirements().
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
	livenessCmd := fmt.Sprintf("mysqladmin -u%s -p$MYSQL_ROOT_PASSWORD ping",
		managedServices[service.Type].username)
	readinessCmd := fmt.Sprintf("mysql -h 127.0.0.1 -u%s -p$MYSQL_ROOT_PASSWORD -e 'SELECT 1'",
		managedServices[service.Type].username)

	container := applyConfigsCoreV1.Container().
		WithName(containerName).
		WithImage(managedServices[service.Type].image).
		WithPorts(applyConfigsCoreV1.ContainerPort().WithContainerPort(int32(managedServices[service.Type].podPort))).
		WithVolumeMounts(applyConfigsCoreV1.VolumeMount().WithName("data").WithMountPath("/var/lib/mysql")).
		WithEnv(applyConfigsCoreV1.EnvVar().
			WithName("MYSQL_ROOT_PASSWORD").
			WithValueFrom(m.createPasswordEnvVarSource(service)),
			applyConfigsCoreV1.EnvVar().
				WithName("MYSQL_DATABASE").
				WithValue("db")).
		WithLivenessProbe(applyConfigsCoreV1.Probe().
			WithExec(applyConfigsCoreV1.ExecAction().
				WithCommand("/bin/sh", "-c", livenessCmd)).
			WithInitialDelaySeconds(20).
			WithPeriodSeconds(20).
			WithTimeoutSeconds(5).
			WithFailureThreshold(3)).
		WithReadinessProbe(applyConfigsCoreV1.Probe().
			WithExec(applyConfigsCoreV1.ExecAction().
				WithCommand("/bin/sh", "-c", readinessCmd)).
			WithInitialDelaySeconds(30).
			WithPeriodSeconds(20).
			WithTimeoutSeconds(5).
			WithFailureThreshold(3))

	podTemplate := applyConfigsCoreV1.PodTemplateSpec().
		WithLabels(map[string]string{"app": service.Name}).
		WithSpec(applyConfigsCoreV1.PodSpec().WithContainers(container).WithTerminationGracePeriodSeconds(10))

	pvClaim := applyConfigsCoreV1.PersistentVolumeClaim("data", service.Project).
		WithSpec(applyConfigsCoreV1.PersistentVolumeClaimSpec().
			WithAccessModes(v1.ReadWriteOnce).
			WithResources(applyConfigsCoreV1.VolumeResourceRequirements().
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

func (m managedServicesImpl) createMongoDeployment(ctx context.Context, service openapi.ManagedService) error {
	livenessCmd := fmt.Sprintf("mongosh --port 27017 --username %s --password $MONGO_INITDB_ROOT_PASSWORD "+
		"--eval 'db.runCommand({ping: 1})' --quiet",
		managedServices[service.Type].username)
	readinessCmd := fmt.Sprintf("mongosh --port 27017 --username %s --password $MONGO_INITDB_ROOT_PASSWORD "+
		"--eval 'db.serverStatus().ok' --quiet | grep -q 1",
		managedServices[service.Type].username)

	container := applyConfigsCoreV1.Container().
		WithName(containerName).
		WithImage(managedServices[service.Type].image).
		WithPorts(applyConfigsCoreV1.ContainerPort().WithContainerPort(int32(managedServices[service.Type].podPort))).
		WithVolumeMounts(applyConfigsCoreV1.VolumeMount().WithName("data").WithMountPath("/var/lib/mongodb")).
		WithEnv(applyConfigsCoreV1.EnvVar().
			WithName("MONGO_INITDB_ROOT_PASSWORD").
			WithValueFrom(m.createPasswordEnvVarSource(service)),
			applyConfigsCoreV1.EnvVar().
				WithName("MONGO_INITDB_ROOT_USERNAME").
				WithValue(managedServices[service.Type].username)).
		WithLivenessProbe(applyConfigsCoreV1.Probe().
			WithExec(applyConfigsCoreV1.ExecAction().
				WithCommand("/bin/sh", "-c", livenessCmd)).
			WithInitialDelaySeconds(20).
			WithPeriodSeconds(20).
			WithTimeoutSeconds(5).
			WithFailureThreshold(3)).
		WithReadinessProbe(applyConfigsCoreV1.Probe().
			WithExec(applyConfigsCoreV1.ExecAction().
				WithCommand("/bin/sh", "-c", readinessCmd)).
			WithInitialDelaySeconds(30).
			WithPeriodSeconds(20).
			WithTimeoutSeconds(5).
			WithFailureThreshold(3))

	podTemplate := applyConfigsCoreV1.PodTemplateSpec().
		WithLabels(map[string]string{"app": service.Name}).
		WithSpec(applyConfigsCoreV1.PodSpec().WithContainers(container).WithTerminationGracePeriodSeconds(10))

	pvClaim := applyConfigsCoreV1.PersistentVolumeClaim("data", service.Project).
		WithSpec(applyConfigsCoreV1.PersistentVolumeClaimSpec().
			WithAccessModes(v1.ReadWriteOnce).
			WithResources(applyConfigsCoreV1.VolumeResourceRequirements().
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
	livenessCmd := "redis-cli --pass $REDIS_PASSWORD ping | grep -q PONG"
	readinessCmd := "redis-cli --pass $REDIS_PASSWORD ping | grep -q PONG"

	container := applyConfigsCoreV1.Container().
		WithName(containerName).
		WithImage(managedServices[service.Type].image).
		WithPorts(applyConfigsCoreV1.ContainerPort().WithContainerPort(int32(managedServices[service.Type].podPort))).
		WithVolumeMounts(applyConfigsCoreV1.VolumeMount().WithName("data").WithMountPath("/data")).
		WithCommand("/bin/sh", "-c", "redis-server --appendonly yes --requirepass ${REDIS_PASSWORD}").
		WithEnv(applyConfigsCoreV1.EnvVar().
			WithName("REDIS_PASSWORD").
			WithValueFrom(m.createPasswordEnvVarSource(service))).
		WithLivenessProbe(applyConfigsCoreV1.Probe().
			WithExec(applyConfigsCoreV1.ExecAction().
				WithCommand("/bin/sh", "-c", livenessCmd)).
			WithInitialDelaySeconds(20).
			WithPeriodSeconds(20).
			WithTimeoutSeconds(5).
			WithFailureThreshold(3)).
		WithReadinessProbe(applyConfigsCoreV1.Probe().
			WithExec(applyConfigsCoreV1.ExecAction().
				WithCommand("/bin/sh", "-c", readinessCmd)).
			WithInitialDelaySeconds(30).
			WithPeriodSeconds(20).
			WithTimeoutSeconds(5).
			WithFailureThreshold(3))

	podTemplate := applyConfigsCoreV1.PodTemplateSpec().
		WithLabels(map[string]string{"app": service.Name}).
		WithSpec(applyConfigsCoreV1.PodSpec().WithContainers(container).WithTerminationGracePeriodSeconds(10))

	pvClaim := applyConfigsCoreV1.PersistentVolumeClaim("data", service.Project).
		WithSpec(applyConfigsCoreV1.PersistentVolumeClaimSpec().
			WithAccessModes(v1.ReadWriteOnce).
			WithResources(applyConfigsCoreV1.VolumeResourceRequirements().
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
				WithValue("rabbit@$(HOSTNAME)."+service.Name+".$(NAMESPACE).svc.cluster.local"),
			applyConfigsCoreV1.EnvVar().WithName("RABBITMQ_DEFAULT_USER").
				WithValue(managedServices[service.Type].username),
			applyConfigsCoreV1.EnvVar().WithName("RABBITMQ_DEFAULT_PASS").
				WithValueFrom(m.createPasswordEnvVarSource(service)),
			applyConfigsCoreV1.EnvVar().WithName("RABBITMQ_ERLANG_COOKIE").
				WithValue("secret_cookie_12345678"), // TODO refactor
		).
		WithLivenessProbe(applyConfigsCoreV1.Probe().
			WithExec(applyConfigsCoreV1.ExecAction().WithCommand("rabbitmq-diagnostics", "status", "--timeout", "10")).
			WithInitialDelaySeconds(20).
			WithPeriodSeconds(20).
			WithTimeoutSeconds(15).
			WithFailureThreshold(3)).
		WithReadinessProbe(applyConfigsCoreV1.Probe().
			WithExec(applyConfigsCoreV1.ExecAction().WithCommand("rabbitmq-diagnostics", "ping", "--timeout", "10")).
			WithInitialDelaySeconds(30).
			WithPeriodSeconds(20).
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
	if err != nil && !apierrors.IsNotFound(err) {
		return errors.Wrap(err, "failed to delete managed service StatefulSet")
	}

	err = m.deleteK8sService(ctx, namespace, name)
	if err != nil && !apierrors.IsNotFound(err) {
		log.WithError(err).Errorln("Failed to delete K8s service after deleting managed service, skipping")
	}

	err = m.deletePasswordSecret(ctx, namespace, name)
	if err != nil && !apierrors.IsNotFound(err) {
		log.WithError(err).Errorln("Failed to delete password secret after deleting managed service, skipping")
	}
	return nil
}

func (m managedServicesImpl) deleteK8sService(ctx context.Context, namespace string, name string) error {
	err := m.clientset.CoreV1().Services(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return errors.Wrap(err, "failed to delete K8s service for managed service")
	}
	return nil
}

func (m managedServicesImpl) deletePasswordSecret(ctx context.Context, namespace string, name string) error {
	err := m.clientset.CoreV1().Secrets(namespace).Delete(ctx, name+"-password", metav1.DeleteOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return errors.Wrap(err, "failed to delete secret for managed service")
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

	ssOptions := metav1.ListOptions{
		LabelSelector: "letsdeploy.space/managed=true",
	}
	statefulSets, err := m.clientset.AppsV1().StatefulSets(projectId).List(ctx, ssOptions)
	if err != nil {
		return errors.Wrap(err, "failed to get statefulsets list")
	}
	for _, statefulSet := range statefulSets.Items {
		if !contains(servicesMap, statefulSet.Name) {
			err := m.deleteManagedServiceDeployment(ctx, projectId, statefulSet.Name)
			if err != nil {
				log.WithError(err).Errorf("Failed to delete managed service statefulset %s, skipping\n", statefulSet.Name)
			}
		}
	}

	serviceOptions := metav1.ListOptions{
		LabelSelector: "letsdeploy.space/managed=true,letsdeploy.space/service-type=managed",
	}
	k8sServices, err := m.clientset.CoreV1().Services(projectId).List(ctx, serviceOptions)
	if err != nil {
		return errors.Wrap(err, "failed to get K8s services")
	}
	for _, k8sService := range k8sServices.Items {
		if !contains(servicesMap, k8sService.Name) {
			err := m.deleteK8sService(ctx, projectId, k8sService.Name)
			if err != nil {
				log.WithError(err).Errorf("Failed to delete k8s service %s, skipping\n", k8sService.Name)
			}
		}
	}

	return nil
}
