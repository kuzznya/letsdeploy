package core

import (
	"context"
	"fmt"
	"github.com/kuzznya/letsdeploy/app/apperrors"
	"github.com/kuzznya/letsdeploy/app/models"
	"github.com/kuzznya/letsdeploy/app/storage"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Projects struct {
	storage   *storage.Storage
	clientset *kubernetes.Clientset
}

func (p *Projects) CreateProject(ctx context.Context, project models.Project, author string) (*models.Project, error) {
	record := storage.ProjectDBModel{Id: project.Id, Name: project.Name}
	err := p.storage.ExecTx(ctx, func(s *storage.Storage) error {
		id, err := s.ProjectRepository().CreateNew(record)
		if err != nil {
			return err
		}
		project.Id = id

		err = s.ProjectRepository().AddParticipant(id, author)
		if err != nil {
			return err
		}

		namespace, err := p.clientset.CoreV1().Namespaces().Create(
			ctx,
			&v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: fmt.Sprintf("project-%d", project.Id),
				},
			},
			metav1.CreateOptions{})
		if err != nil {
			return err
		}
		log.Infof("Namespace %s created for a new project", namespace.Name)

		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new project")
	}
	return &project, nil
}

func (p *Projects) GetProject(id int, requester string) (*models.Project, error) {
	if err := p.checkAccess(id, requester); err != nil {
		return nil, err
	}
	record, err := p.storage.ProjectRepository().FindByID(id)
	if err != nil {
		return nil, apperrors.NotFoundWrap(err, "cannot find project by id")
	}
	return &models.Project{Id: record.Id, Name: record.Name}, nil
}

func (p *Projects) UpdateProject(project models.Project, requester string) error {
	if err := p.checkAccess(project.Id, requester); err != nil {
		return err
	}
	record := storage.ProjectDBModel{Id: project.Id, Name: project.Name}
	err := p.storage.ProjectRepository().Update(record)
	if err != nil {
		return errors.Wrap(err, "failed to update project")
	}
	return nil
}

func (p *Projects) DeleteProject(ctx context.Context, id int, requester string) error {
	if err := p.checkAccess(id, requester); err != nil {
		return err
	}
	err := p.storage.ExecTx(ctx, func(s *storage.Storage) error {
		err := p.clientset.CoreV1().Namespaces().Delete(ctx, fmt.Sprintf("project-%d", id), metav1.DeleteOptions{})
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

func (p *Projects) GetUserProjects(username string) ([]models.Project, error) {
	projects, err := p.storage.ProjectRepository().FindUserProjects(username)
	if err != nil {
		return nil, errors.Wrap(err, "failed to find user's projects")
	}
	result := make([]models.Project, len(projects))
	for i, record := range projects {
		result[i] = models.Project{Id: record.Id, Name: record.Name}
	}
	return result, nil
}

func (p *Projects) GetParticipants(id int, requester string) ([]string, error) {
	if err := p.checkAccess(id, requester); err != nil {
		return nil, err
	}
	participants, err := p.storage.ProjectRepository().GetParticipants(id)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get project participants")
	}
	return participants, nil
}

func (p *Projects) AddParticipant(id int, username string, requester string) error {
	if err := p.checkAccess(id, requester); err != nil {
		return err
	}
	err := p.storage.ProjectRepository().AddParticipant(id, username)
	if err != nil {
		return errors.Wrap(err, "failed to add participant")
	}
	return nil
}

func (p *Projects) RemoveParticipant(id int, username string, requester string) error {
	if err := p.checkAccess(id, requester); err != nil {
		return err
	}
	err := p.storage.ProjectRepository().RemoveParticipant(id, username)
	if err != nil {
		return errors.Wrap(err, "failed to add participant")
	}
	return nil
}

func (p *Projects) checkAccess(id int, user string) error {
	isParticipant, err := p.storage.ProjectRepository().IsParticipant(id, user)
	if err != nil {
		return errors.Wrap(err, "project access check unexpected failure")
	}
	if !isParticipant {
		return apperrors.NotFound(fmt.Sprintf("cannot find project with id %d", id))
	}
	return nil
}
