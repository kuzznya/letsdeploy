package core

import (
	"context"
	"database/sql"
	"github.com/kuzznya/letsdeploy/app/apperrors"
	"github.com/kuzznya/letsdeploy/app/middleware"
	"github.com/kuzznya/letsdeploy/app/storage"
	"github.com/kuzznya/letsdeploy/internal/openapi"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	networkingV1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	applyConfigsAppsV1 "k8s.io/client-go/applyconfigurations/apps/v1"
	applyConfigsCoreV1 "k8s.io/client-go/applyconfigurations/core/v1"
	applyConfigsMetaV1 "k8s.io/client-go/applyconfigurations/meta/v1"
	applyConfigsNetworkingV1 "k8s.io/client-go/applyconfigurations/networking/v1"
	"k8s.io/client-go/kubernetes"
	"strings"
)

const containerName = "container-0"

type Services interface {
	projectSynchronizable
	GetProjectServices(project string, auth middleware.Authentication) ([]openapi.Service, error)
	CreateService(ctx context.Context, service openapi.Service, auth middleware.Authentication) (*openapi.Service, error)
	GetService(id int, auth middleware.Authentication) (*openapi.Service, error)
	UpdateService(ctx context.Context, service openapi.Service, auth middleware.Authentication) (*openapi.Service, error)
	DeleteService(ctx context.Context, id int, auth middleware.Authentication) error
	GetServiceEnvVars(id int, auth middleware.Authentication) ([]openapi.EnvVar, error)
	SetServiceEnvVar(ctx context.Context, serviceId int, envVar openapi.EnvVar, auth middleware.Authentication) (*openapi.EnvVar, error)
	DeleteServiceEnvVar(ctx context.Context, serviceId int, envVarName string, auth middleware.Authentication) error
}

type servicesImpl struct {
	projects  Projects
	storage   *storage.Storage
	clientset *kubernetes.Clientset
}

var _ Services = (*servicesImpl)(nil)

func InitServices(
	projects Projects,
	storage *storage.Storage,
	clientset *kubernetes.Clientset,
) Services {
	s := servicesImpl{projects: projects, storage: storage, clientset: clientset}
	return &s
}

func (s servicesImpl) GetProjectServices(project string, auth middleware.Authentication) ([]openapi.Service, error) {
	if err := s.projects.checkAccess(project, auth); err != nil {
		return nil, err
	}
	entities, err := s.storage.ServiceRepository().FindByProjectId(project)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get project services")
	}
	services := make([]openapi.Service, len(entities))
	for i, entity := range entities {
		id := entity.Id
		envVarEntities, err := s.storage.EnvVarRepository().FindByServiceId(id)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get service env vars")
		}
		envVars, err := mapEnvVarEntities(envVarEntities)
		if err != nil {
			return nil, err
		}
		services[i] = openapi.Service{
			Id:              &id,
			Image:           entity.Image,
			Name:            entity.Name,
			Port:            entity.Port,
			Project:         entity.ProjectId,
			PublicApiPrefix: fromNullString(entity.PublicApiPrefix),
			EnvVars:         envVars,
		}
	}
	return services, nil
}

func (s servicesImpl) CreateService(ctx context.Context, service openapi.Service, auth middleware.Authentication) (*openapi.Service, error) {
	if err := s.projects.checkAccess(service.Project, auth); err != nil {
		return nil, err
	}
	record := storage.ServiceEntity{
		ProjectId:       service.Project,
		Name:            service.Name,
		Image:           service.Image,
		Port:            service.Port,
		PublicApiPrefix: toNullString(service.PublicApiPrefix),
	}
	err := s.storage.ExecTx(ctx, func(store *storage.Storage) error {
		id, err := store.ServiceRepository().CreateNew(record)
		if err != nil {
			return err
		}
		service.Id = &id

		envVars := mapItems(service.EnvVars, func(v openapi.EnvVar) storage.EnvVarEntity {
			varEntity := storage.EnvVarEntity{
				Name:      v.Name,
				ServiceId: id,
			}
			processEnvVar(v,
				func(e openapi.EnvVar0) { varEntity.Value = sql.NullString{String: e.Value, Valid: true} },
				func(e openapi.EnvVar1) { varEntity.Secret = sql.NullString{String: e.Secret, Valid: true} })
			return varEntity
		})

		if len(envVars) > 0 {
			_, err = store.EnvVarRepository().CreateOrUpdateAll(envVars)
			if err != nil {
				return err
			}
		}

		err = s.applyServiceDeployment(ctx, service)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new service")
	}
	log.Infof("Created service %s in project %s", service.Name, service.Project)
	return &service, nil
}

func (s servicesImpl) GetService(id int, auth middleware.Authentication) (*openapi.Service, error) {
	entity, err := s.storage.ServiceRepository().FindByID(id)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get service by id")
	}
	if err := s.projects.checkAccess(entity.ProjectId, auth); err != nil {
		return nil, err
	}
	varEntities, err := s.storage.EnvVarRepository().FindByServiceId(id)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get service env vars")
	}

	envVars, err := mapEnvVarEntities(varEntities)
	if err != nil {
		return nil, err
	}

	return &openapi.Service{
		Id:              &entity.Id,
		Image:           entity.Image,
		Name:            entity.Name,
		Port:            entity.Port,
		Project:         entity.ProjectId,
		EnvVars:         envVars,
		PublicApiPrefix: fromNullString(entity.PublicApiPrefix),
	}, nil
}

func (s servicesImpl) UpdateService(ctx context.Context, service openapi.Service, auth middleware.Authentication) (*openapi.Service, error) {
	retrieved, err := s.GetService(*service.Id, auth)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get service by id")
	}
	if retrieved.Project != service.Project {
		return nil, apperrors.BadRequest("Project field cannot be updated")
	}
	updated := storage.ServiceEntity{
		Id:              *service.Id,
		ProjectId:       retrieved.Project,
		Name:            service.Name,
		Image:           service.Image,
		Port:            service.Port,
		PublicApiPrefix: toNullString(service.PublicApiPrefix),
	}
	err = s.storage.ExecTx(ctx, func(store *storage.Storage) error {
		err := store.ServiceRepository().Update(updated)
		if err != nil {
			return err
		}

		prevEnvVars, err := store.EnvVarRepository().FindByServiceId(*service.Id)
		if err != nil {
			return err
		}

		envVars := mapItems(service.EnvVars, func(v openapi.EnvVar) storage.EnvVarEntity {
			varEntity := storage.EnvVarEntity{
				Name:      v.Name,
				ServiceId: *service.Id,
			}
			processEnvVar(v,
				func(e openapi.EnvVar0) { varEntity.Value = sql.NullString{String: e.Value, Valid: true} },
				func(e openapi.EnvVar1) { varEntity.Secret = sql.NullString{String: e.Secret, Valid: true} })
			return varEntity
		})

		_, err = store.EnvVarRepository().CreateOrUpdateAll(envVars)
		if err != nil {
			return err
		}

		envVarNames := make(map[string]bool)
		for _, v := range envVars {
			envVarNames[v.Name] = true
		}
		for _, prevVar := range prevEnvVars {
			if _, ok := envVarNames[prevVar.Name]; !ok {
				err := store.EnvVarRepository().DeleteByServiceIdAndName(*service.Id, prevVar.Name)
				if err != nil {
					return err
				}
			}
		}

		err = s.applyServiceDeployment(ctx, service)
		if err != nil {
			return err
		}

		if updated.Name != retrieved.Name {
			err := s.deleteServiceDeployment(ctx, retrieved.Project, retrieved.Name)
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
		Id:              &updated.Id,
		Image:           updated.Image,
		Name:            updated.Name,
		Port:            updated.Port,
		Project:         updated.ProjectId,
		EnvVars:         service.EnvVars,
		PublicApiPrefix: fromNullString(updated.PublicApiPrefix),
	}
	log.Infof("Updated service %s in project %s", service.Name, service.Project)
	return &result, nil
}

func (s servicesImpl) DeleteService(ctx context.Context, id int, auth middleware.Authentication) error {
	service, err := s.GetService(id, auth)
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
	log.Infof("Deleted service %s in project %s", service.Name, service.Project)
	return nil
}

func (s servicesImpl) GetServiceEnvVars(id int, auth middleware.Authentication) ([]openapi.EnvVar, error) {
	_, err := s.GetService(id, auth)
	if err != nil {
		return nil, errors.Wrap(err, "failed to find service by id")
	}
	entities, err := s.storage.EnvVarRepository().FindByServiceId(id)
	if err != nil {
		return nil, err
	}
	envVars := make([]openapi.EnvVar, len(entities))
	for i, entity := range entities {
		envVars[i] = openapi.EnvVar{Name: entity.Name}
		if entity.Value.Valid {
			err := envVars[i].FromEnvVar0(openapi.EnvVar0{Value: entity.Value.String})
			if err != nil {
				return nil, errors.Wrap(err, "failed to map env var entity to model")
			}
		} else if entity.Secret.Valid {
			err := envVars[i].FromEnvVar1(openapi.EnvVar1{Secret: entity.Secret.String})
			if err != nil {
				return nil, errors.Wrap(err, "failed to map env var entity to model")
			}
		} else {
			return nil, errors.New("Unknown env var type: both Value and Secret are null")
		}
	}
	return envVars, nil
}

func (s servicesImpl) SetServiceEnvVar(ctx context.Context, serviceId int, envVar openapi.EnvVar, auth middleware.Authentication) (*openapi.EnvVar, error) {
	service, err := s.GetService(serviceId, auth)
	if err != nil {
		return nil, errors.Wrap(err, "failed to find service by id")
	}
	entity := storage.EnvVarEntity{
		ServiceId: serviceId,
		Name:      envVar.Name,
	}

	processEnvVar(envVar,
		func(e openapi.EnvVar0) { entity.Value = sql.NullString{String: e.Value, Valid: true} },
		func(e openapi.EnvVar1) { entity.Secret = sql.NullString{String: e.Secret, Valid: true} })

	newEnv := removeFirstByPredicate(service.EnvVars, func(e openapi.EnvVar) bool { return e.Name == envVar.Name })
	newEnv = append(newEnv, envVar)
	service.EnvVars = newEnv

	err = s.storage.ExecTx(ctx, func(store *storage.Storage) error {
		_, err = store.EnvVarRepository().CreateOrUpdate(entity)
		if err != nil {
			return err
		}

		err := s.applyServiceDeployment(ctx, *service)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to save env var")
	}
	log.Infof("Set service %s env var %s", service.Name, envVar.Name)
	return &envVar, nil
}

func (s servicesImpl) DeleteServiceEnvVar(ctx context.Context, serviceId int, envVarName string, auth middleware.Authentication) error {
	service, err := s.GetService(serviceId, auth)
	if err != nil {
		return errors.Wrap(err, "failed to find service by id")
	}
	err = s.storage.ExecTx(ctx, func(store *storage.Storage) error {
		err := store.EnvVarRepository().DeleteByServiceIdAndName(serviceId, envVarName)
		if err != nil {
			return err
		}
		service.EnvVars = filter(service.EnvVars, func(v openapi.EnvVar) bool { return v.Name != envVarName })
		err = s.applyServiceDeployment(ctx, *service)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return errors.Wrap(err, "failed to delete service env var")
	}
	log.Infof("Deleted service %s env var %s", service.Name, envVarName)
	return nil
}

func (s servicesImpl) applyServiceDeployment(ctx context.Context, service openapi.Service) error {
	if err := s.createK8sService(ctx, service); err != nil {
		return err
	}
	if err := s.createIngress(ctx, service); err != nil {
		if err := s.deleteK8sService(ctx, service.Project, service.Name); err != nil {
			log.WithError(err).Errorln("Failed to delete K8s service for user service after ingress deployment failure, skipping")
		}
		return err
	}

	limits := v1.ResourceList{}
	limits.Cpu().SetMilli(250)
	limits.Memory().SetScaled(512, resource.Mega)
	container := applyConfigsCoreV1.Container().
		WithName(containerName).
		WithImage(service.Image).
		WithImagePullPolicy(v1.PullAlways).
		WithPorts(applyConfigsCoreV1.ContainerPort().WithContainerPort(int32(service.Port))).
		WithResources(applyConfigsCoreV1.ResourceRequirements().WithLimits(limits))
	if service.EnvVars == nil {
		service.EnvVars = []openapi.EnvVar{}
	}
	for _, envVar := range service.EnvVars {
		processEnvVar(envVar,
			func(e openapi.EnvVar0) {
				container = container.WithEnv(applyConfigsCoreV1.EnvVar().WithName(envVar.Name).WithValue(e.Value))
			},
			func(e openapi.EnvVar1) {
				source := applyConfigsCoreV1.EnvVarSource().
					WithSecretKeyRef(applyConfigsCoreV1.SecretKeySelector().WithName(e.Secret).WithKey(secretKey))
				container = container.WithEnv(applyConfigsCoreV1.EnvVar().WithName(envVar.Name).WithValueFrom(source))
			})
	}
	podTemplate := applyConfigsCoreV1.PodTemplateSpec().
		WithLabels(map[string]string{"app": service.Name}).
		WithSpec(applyConfigsCoreV1.PodSpec().WithContainers(container))
	deployment := applyConfigsAppsV1.Deployment(service.Name, service.Project).
		WithLabels(map[string]string{"letsdeploy.space/managed": "true"}).
		WithSpec(applyConfigsAppsV1.DeploymentSpec().
			WithSelector(applyConfigsMetaV1.LabelSelector().
				WithMatchLabels(map[string]string{"app": service.Name})).
			WithTemplate(podTemplate))

	_, err := s.clientset.AppsV1().Deployments(service.Project).
		Apply(ctx, deployment, metav1.ApplyOptions{FieldManager: "letsdeploy"})
	if err != nil {
		if err := s.deleteIngress(ctx, service.Project, service.Name); err != nil {
			log.WithError(err).Errorln("Failed to delete ingress after deployment failure, skipping")
		}
		if err := s.deleteK8sService(ctx, service.Project, service.Name); err != nil {
			log.WithError(err).Errorln("Failed to delete K8s service after deployment failure, skipping")
		}
		return errors.Wrap(err, "failed to create service deployment")
	}
	return nil
}

func (s servicesImpl) createK8sService(ctx context.Context, service openapi.Service) error {
	port := applyConfigsCoreV1.ServicePort().
		WithPort(80).
		WithTargetPort(intstr.FromInt(service.Port))
	svc := applyConfigsCoreV1.Service(service.Name, service.Project).
		WithLabels(map[string]string{
			"letsdeploy.space/managed":      "true",
			"letsdeploy.space/service-type": "service",
			"app":                           service.Name,
		}).
		WithSpec(applyConfigsCoreV1.ServiceSpec().WithPorts(port).
			WithSelector(map[string]string{"app": service.Name}))
	_, err := s.clientset.CoreV1().Services(service.Project).Apply(ctx, svc, metav1.ApplyOptions{FieldManager: "letsdeploy"})
	if err != nil {
		return errors.Wrap(err, "failed to create K8s service for user service")
	}
	return nil
}

func (s servicesImpl) createIngress(ctx context.Context, service openapi.Service) error {
	if service.PublicApiPrefix == nil {
		return nil
	}
	backend := applyConfigsNetworkingV1.IngressBackend().
		WithService(applyConfigsNetworkingV1.IngressServiceBackend().
			WithName(service.Name).
			WithPort(applyConfigsNetworkingV1.ServiceBackendPort().WithNumber(80)))
	path := applyConfigsNetworkingV1.HTTPIngressPath().
		WithPathType(networkingV1.PathTypePrefix).
		WithPath(*service.PublicApiPrefix).
		WithBackend(backend)
	rule := applyConfigsNetworkingV1.IngressRule().
		WithHost(service.Project + ".letsdeploy.space").
		WithHTTP(applyConfigsNetworkingV1.HTTPIngressRuleValue().WithPaths(path))
	ingress := applyConfigsNetworkingV1.Ingress(service.Name+"-ingress", service.Project).
		WithLabels(map[string]string{
			"letsdeploy.space/managed":      "true",
			"letsdeploy.space/service-type": "service",
		}).
		WithSpec(applyConfigsNetworkingV1.IngressSpec().WithRules(rule))
	_, err := s.clientset.NetworkingV1().Ingresses(service.Project).Apply(ctx, ingress, metav1.ApplyOptions{FieldManager: "letsdeploy"})
	if err != nil {
		return errors.Wrap(err, "failed to create Ingress for service")
	}
	return nil
}

func (s servicesImpl) deleteServiceDeployment(ctx context.Context, project string, service string) error {
	err := s.clientset.AppsV1().Deployments(project).Delete(ctx, service, metav1.DeleteOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return errors.Wrap(err, "failed to delete service deployment")
	}

	err = s.deleteK8sService(ctx, project, service)
	if err != nil && !apierrors.IsNotFound(err) {
		log.WithError(err).Errorf("Failed to delete K8s service %s after deleting service deployment in namespace %s\n", service, project)
	}

	err = s.deleteIngress(ctx, project, service)
	if err != nil && !apierrors.IsNotFound(err) {
		log.WithError(err).Errorf("Failed to delete ingress %s after deleting service deployment in namespace %s\n", service, project)
	}
	return nil
}

func (s servicesImpl) deleteK8sService(ctx context.Context, project string, service string) error {
	err := s.clientset.CoreV1().Services(project).Delete(ctx, service, metav1.DeleteOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return errors.Wrap(err, "failed to delete K8s service for user service")
	}
	log.Debugf("Deleted K8s service %s in namespace %s", service, project)
	return nil
}

func (s servicesImpl) deleteIngress(ctx context.Context, project string, service string) error {
	err := s.clientset.NetworkingV1().Ingresses(project).Delete(ctx, service+"-ingress", metav1.DeleteOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return errors.Wrap(err, "failed to delete ingress for service")
	}
	log.Debugf("Deleted ingress %s in namespace %s", service, project)
	return nil
}

func (s servicesImpl) syncKubernetes(ctx context.Context, projectId string) error {
	services, err := s.GetProjectServices(projectId, middleware.ServiceAccount)
	if err != nil {
		return errors.Wrap(err, "failed to get project services")
	}
	servicesMap := toMapSelf(services, func(item openapi.Service) string { return item.Name })
	for _, service := range services {
		err := s.applyServiceDeployment(ctx, service)
		if err != nil {
			log.WithError(err).Errorf("Failed to create service deployment %s, skipping\n", service.Name)
		}
	}

	deploymentOptions := metav1.ListOptions{
		LabelSelector: "letsdeploy.space/managed=true",
	}
	deployments, err := s.clientset.AppsV1().Deployments(projectId).List(ctx, deploymentOptions)
	if err != nil {
		return errors.Wrap(err, "failed to get deployments list")
	}
	for _, deployment := range deployments.Items {
		if !contains(servicesMap, deployment.Name) {
			err := s.deleteServiceDeployment(ctx, projectId, deployment.Name)
			if err != nil {
				log.WithError(err).Errorf("Failed to delete deployment %s, skipping\n", deployment.Name)
			}
		} else {
			log.Debugf("Checked deployment %s", deployment.Name)
		}
	}

	serviceOptions := metav1.ListOptions{
		LabelSelector: "letsdeploy.space/managed=true,letsdeploy.space/service-type=service",
	}
	k8sServices, err := s.clientset.CoreV1().Services(projectId).List(ctx, serviceOptions)
	if err != nil {
		return errors.Wrap(err, "failed to get K8s services")
	}
	for _, k8sService := range k8sServices.Items {
		if !contains(servicesMap, k8sService.Name) {
			err := s.deleteK8sService(ctx, projectId, k8sService.Name)
			if err != nil {
				log.WithError(err).Errorf("Failed to delete k8s service %s, skipping\n", k8sService.Name)
			}
		} else {
			log.Debugf("Checked K8s service %s", k8sService.Name)
		}
	}

	ingresses, err := s.clientset.CoreV1().Services(projectId).List(ctx, deploymentOptions)
	if err != nil {
		return errors.Wrap(err, "failed to get ingresses")
	}
	for _, ingress := range ingresses.Items {
		name, _ := strings.CutSuffix(ingress.Name, "-ingress")
		if !contains(servicesMap, name) {
			err := s.deleteIngress(ctx, projectId, name)
			if err != nil {
				log.WithError(err).Errorf("Failed to delete ingress %s, skipping\n", ingress.Name)
			}
		}
	}

	return nil
}

func processEnvVar(envVar openapi.EnvVar, onValue func(openapi.EnvVar0), onSecret func(openapi.EnvVar1)) {
	withValue, _ := envVar.AsEnvVar0()
	if withValue.Value != "" {
		onValue(withValue)
		return
	}
	withSecret, _ := envVar.AsEnvVar1()
	if withSecret.Secret != "" {
		onSecret(withSecret)
		return
	}
}

func mapEnvVarEntities(entities []storage.EnvVarEntity) ([]openapi.EnvVar, error) {
	envVars := make([]openapi.EnvVar, len(entities))
	for i, entity := range entities {
		envVars[i] = openapi.EnvVar{Name: entity.Name}
		if entity.Value.Valid {
			err := envVars[i].FromEnvVar0(openapi.EnvVar0{Value: entity.Value.String})
			if err != nil {
				return nil, errors.Wrap(err, "failed to map env var entity to model")
			}
		} else if entity.Secret.Valid {
			err := envVars[i].FromEnvVar1(openapi.EnvVar1{Secret: entity.Secret.String})
			if err != nil {
				return nil, errors.Wrap(err, "failed to map env var entity to model")
			}
		} else {
			return nil, errors.New("Unknown env var type: both Value and Secret are null")
		}
	}
	return envVars, nil
}
