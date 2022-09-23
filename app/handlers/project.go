package handlers

import (
	"context"
	"fmt"
	"github.com/kuzznya/letsdeploy/app/appErrors"
	"github.com/kuzznya/letsdeploy/app/models"
	"github.com/kuzznya/letsdeploy/app/storage"
	"github.com/pkg/errors"
)

func CreateProject(ctx context.Context, s *storage.Storage, project models.Project, author string) (*models.Project, error) {
	record := storage.ProjectDBModel{Id: project.Id, Name: project.Name}
	err := s.ExecTx(ctx, func(s *storage.Storage) error {
		id, err := s.ProjectRepository().CreateNew(record)
		if err != nil {
			return err
		}
		project.Id = id

		err = s.ProjectRepository().AddParticipant(id, author)
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

func GetProject(s *storage.Storage, id int, requester string) (*models.Project, error) {
	if err := checkAccess(s, id, requester); err != nil {
		return nil, err
	}
	record, err := s.ProjectRepository().FindByID(id)
	if err != nil {
		return nil, appErrors.NotFoundWrap(err, "cannot find project by id")
	}
	return &models.Project{Id: record.Id, Name: record.Name}, nil
}

func UpdateProject(s *storage.Storage, project models.Project, requester string) error {
	if err := checkAccess(s, project.Id, requester); err != nil {
		return err
	}
	record := storage.ProjectDBModel{Id: project.Id, Name: project.Name}
	err := s.ProjectRepository().Update(record)
	if err != nil {
		return errors.Wrap(err, "failed to update project")
	}
	return nil
}

func DeleteProject(s *storage.Storage, id int, requester string) error {
	if err := checkAccess(s, id, requester); err != nil {
		return err
	}
	err := s.ProjectRepository().Delete(id)
	if err != nil {
		return errors.Wrap(err, "failed to delete project")
	}
	return nil
}

func GetUserProjects(s *storage.Storage, username string) ([]models.Project, error) {
	projects, err := s.ProjectRepository().FindUserProjects(username)
	if err != nil {
		return nil, errors.Wrap(err, "failed to find user's projects")
	}
	result := make([]models.Project, len(projects))
	for i, record := range projects {
		result[i] = models.Project{Id: record.Id, Name: record.Name}
	}
	return result, nil
}

func GetParticipants(s *storage.Storage, id int, requester string) ([]string, error) {
	if err := checkAccess(s, id, requester); err != nil {
		return nil, err
	}
	participants, err := s.ProjectRepository().GetParticipants(id)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get project participants")
	}
	return participants, nil
}

func AddParticipant(s *storage.Storage, id int, username string, requester string) error {
	if err := checkAccess(s, id, requester); err != nil {
		return err
	}
	err := s.ProjectRepository().AddParticipant(id, username)
	if err != nil {
		return errors.Wrap(err, "failed to add participant")
	}
	return nil
}

func RemoveParticipant(s *storage.Storage, id int, username string, requester string) error {
	if err := checkAccess(s, id, requester); err != nil {
		return err
	}
	err := s.ProjectRepository().RemoveParticipant(id, username)
	if err != nil {
		return errors.Wrap(err, "failed to add participant")
	}
	return nil
}

func checkAccess(s *storage.Storage, id int, user string) error {
	isParticipant, err := s.ProjectRepository().IsParticipant(id, user)
	if err != nil {
		return errors.Wrap(err, "project access check unexpected failure")
	}
	if !isParticipant {
		return appErrors.NotFound(fmt.Sprintf("cannot find project with id %d", id))
	}
	return nil
}
