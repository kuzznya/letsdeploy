package core

import (
	"context"
	"fmt"
	"github.com/kuzznya/letsdeploy/app/apperrors"
	"github.com/kuzznya/letsdeploy/app/storage"
	"github.com/kuzznya/letsdeploy/internal/openapi"
	"github.com/pkg/errors"
	"github.com/procyon-projects/chrono"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"time"
)

const namespaceLabel = "letsdeploy.space/project-namespace"

type Projects interface {
	CreateProject(ctx context.Context, project openapi.Project, author string) (*openapi.Project, error)
	GetProject(id string, requester string) (*openapi.Project, error)
	GetProjectInfo(id string, requester string) (*openapi.ProjectInfo, error)
	UpdateProject(project openapi.Project, requester string) error
	DeleteProject(ctx context.Context, id string, requester string) error
	GetUserProjects(username string) ([]openapi.Project, error)
	GetParticipants(id string, requester string) ([]string, error)
	AddParticipant(id string, username string, requester string) error
	RemoveParticipant(id string, username string, requester string) error
	checkAccess(id string, user string) error
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
	scheduler chrono.TaskScheduler,
) Projects {
	projects := projectsImpl{storage: storage, clientset: clientset, scheduler: scheduler}
	_, err := projects.scheduler.ScheduleWithFixedDelay(projects.syncKubernetes, 1*time.Minute)
	if err != nil {
		log.WithError(err).Panicln("Unable to schedule k8s synchronization")
	}
	return &projects
}

func (p projectsImpl) setServices(services Services) {
	p.services = services
}

func (p projectsImpl) setManagedServices(managedServices ManagedServices) {
	p.managedServices = managedServices
}

func (p projectsImpl) CreateProject(ctx context.Context, project openapi.Project, author string) (*openapi.Project, error) {
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

		err = s.ProjectRepository().AddParticipant(id, author)
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

func (p projectsImpl) GetProject(id string, requester string) (*openapi.Project, error) {
	if err := p.checkAccess(id, requester); err != nil {
		return nil, err
	}
	record, err := p.storage.ProjectRepository().FindByID(id)
	if err != nil {
		return nil, apperrors.WrapNonAppError(err, "cannot find project by id")
	}
	return &openapi.Project{Id: record.Id}, nil
}

func (p projectsImpl) GetProjectInfo(id string, requester string) (*openapi.ProjectInfo, error) {
	if err := p.checkAccess(id, requester); err != nil {
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
	services, err := p.services.GetProjectServices(id, requester)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve project services")
	}
	managedServices, err := p.managedServices.GetProjectManagedServices(id, requester)
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

func (p projectsImpl) UpdateProject(project openapi.Project, requester string) error {
	if err := p.checkAccess(project.Id, requester); err != nil {
		return err
	}
	record := storage.ProjectEntity{Id: project.Id}
	err := p.storage.ProjectRepository().Update(record)
	if err != nil {
		return errors.Wrap(err, "failed to update project")
	}
	return nil
}

func (p projectsImpl) DeleteProject(ctx context.Context, id string, requester string) error {
	if err := p.checkAccess(id, requester); err != nil {
		return err
	}
	err := p.storage.ExecTx(ctx, func(s *storage.Storage) error {
		err := p.clientset.CoreV1().Namespaces().
			Delete(ctx, id, metav1.DeleteOptions{})
		if err != nil && !apierrors.IsNotFound(err) {
			return err
		}
		err = p.storage.ProjectRepository().Delete(id)
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

func (p projectsImpl) GetUserProjects(username string) ([]openapi.Project, error) {
	projects, err := p.storage.ProjectRepository().FindUserProjects(username)
	if err != nil {
		return nil, errors.Wrap(err, "failed to find user's projects")
	}
	result := make([]openapi.Project, len(projects))
	for i, record := range projects {
		result[i] = openapi.Project{Id: record.Id}
	}
	return result, nil
}

func (p projectsImpl) GetParticipants(id string, requester string) ([]string, error) {
	if err := p.checkAccess(id, requester); err != nil {
		return nil, err
	}
	participants, err := p.storage.ProjectRepository().GetParticipants(id)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get project participants")
	}
	return participants, nil
}

func (p projectsImpl) AddParticipant(id string, username string, requester string) error {
	if err := p.checkAccess(id, requester); err != nil {
		return err
	}
	err := p.storage.ProjectRepository().AddParticipant(id, username)
	if err != nil {
		return errors.Wrap(err, "failed to add participant")
	}
	return nil
}

func (p projectsImpl) RemoveParticipant(id string, username string, requester string) error {
	if err := p.checkAccess(id, requester); err != nil {
		return err
	}
	err := p.storage.ProjectRepository().RemoveParticipant(id, username)
	if err != nil {
		return errors.Wrap(err, "failed to add participant")
	}
	return nil
}

func (p projectsImpl) checkAccess(id string, user string) error {
	isParticipant, err := p.storage.ProjectRepository().IsParticipant(id, user)
	if err != nil {
		return errors.Wrap(err, "project access check unexpected failure")
	}
	if !isParticipant {
		return apperrors.NotFound(fmt.Sprintf("cannot find project with id %d", id))
	}
	return nil
}

func (p projectsImpl) createProjectNamespace(ctx context.Context, project openapi.Project) error {
	namespace, err := p.clientset.CoreV1().Namespaces().Get(ctx, project.Id, metav1.GetOptions{})
	if err == nil {
		log.Infof("Namespace %s already exists, not creating new one", namespace.Name)
		return nil
	}
	namespace, err = p.clientset.CoreV1().Namespaces().Create(
		ctx,
		&v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: project.Id,
			},
		},
		metav1.CreateOptions{})
	if err != nil {
		return err
	}
	log.Infof("Namespace %s created for a project", namespace.Name)
	return nil
}

func (p projectsImpl) syncKubernetes(ctx context.Context) {
	log.Infoln("Projects sync started")

	limit := 1000
	offset := 0
	checkedProjects := make(map[string]bool)
	for {
		projects, err := p.storage.ProjectRepository().FindAll(limit, offset)
		if err != nil {
			log.WithError(err).Errorln("Failed to retrieve projects")
			return
		}
		for _, project := range projects {
			checkedProjects[project.Id] = true
			err := p.createProjectNamespace(ctx, openapi.Project{Id: project.Id})
			if err != nil {
				log.WithError(err).Errorf("Failed to create namespace for project %d to synchronize, skipping", project.Id)
			}

			// TODO call Services and ManagedServices to check Deployments and StatefulSets

			log.Debugf("Project %s checked, namespace exists or was created", project.Id)
		}
		if len(projects) < limit {
			break
		}
	}

	p.removeExcessNamespaces(ctx, checkedProjects)

	log.Infof("Projects sync finished")
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
