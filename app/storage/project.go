package storage

import (
	"database/sql"
	"github.com/kuzznya/letsdeploy/app/apperrors"
	"github.com/pkg/errors"
)

type ProjectEntity struct {
	Id string `db:"id"`
}

type ProjectRepository interface {
	CrudRepository[ProjectEntity, string]
	FindAll(limit int, offset int) ([]ProjectEntity, error)
	FindUserProjects(username string) ([]ProjectEntity, error)
	GetParticipants(id string) ([]string, error)
	IsParticipant(id string, username string) (bool, error)
	AddParticipant(id string, username string) error
	RemoveParticipant(id string, username string) error
}

type projectRepositoryImpl struct {
	db QueryExecDB
}

func (r *projectRepositoryImpl) CreateNew(project ProjectEntity) (string, error) {
	var result string
	err := r.db.Get(&result, "INSERT INTO project (id) VALUES ($1) RETURNING id", project.Id)
	if err != nil {
		return "", errors.Wrap(err, "cannot save new project")
	}
	return result, nil
}

func (r *projectRepositoryImpl) ExistsByID(id string) (bool, error) {
	var exists bool
	err := r.db.Get(&exists, "SELECT exists(SELECT * FROM project WHERE id = $1)", id)
	if err != nil {
		return false, errors.Wrap(err, "cannot check if project exists")
	}
	return exists, nil
}

func (r *projectRepositoryImpl) FindByID(id string) (*ProjectEntity, error) {
	var project ProjectEntity
	err := r.db.Get(&project, "SELECT * FROM project WHERE id = $1", id)
	if err == sql.ErrNoRows {
		return nil, apperrors.NotFound("Project not found")
	} else if err != nil {
		return nil, errors.Wrap(err, "cannot find project by id")
	}
	return &project, nil
}

func (r *projectRepositoryImpl) Update(_ ProjectEntity) error {
	panic("Unsupported operation")
}

func (r *projectRepositoryImpl) Delete(id string) error {
	_, err := r.db.Exec("DELETE FROM project WHERE id = $1", id)
	if err != nil {
		return errors.Wrap(err, "failed to delete project")
	}
	return nil
}

func (r *projectRepositoryImpl) FindAll(limit int, offset int) ([]ProjectEntity, error) {
	projects := []ProjectEntity{}
	err := r.db.Select(&projects, "SELECT * FROM project LIMIT $1 OFFSET $2", limit, offset)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve projects")
	}
	return projects, nil
}

func (r *projectRepositoryImpl) FindUserProjects(username string) ([]ProjectEntity, error) {
	projects := []ProjectEntity{}
	err := r.db.Select(&projects,
		`SELECT p.* FROM project p 
    	JOIN project_participant pp ON pp.project_id = p.id 
        WHERE pp.username = $1`,
		username)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve user's projects")
	}
	return projects, nil
}

func (r *projectRepositoryImpl) GetParticipants(id string) ([]string, error) {
	exists, err := r.ExistsByID(id)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.Errorf("project with id %d does not exist", id)
	}
	participants := []string{}
	err = r.db.Select(&participants, "SELECT username FROM project_participant WHERE project_id = $1", id)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get project participants")
	}
	return participants, nil
}

func (r *projectRepositoryImpl) IsParticipant(id string, username string) (bool, error) {
	exists, err := r.ExistsByID(id)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, nil
	}
	var isParticipant bool
	err = r.db.Get(
		&isParticipant,
		`SELECT exists(
    		SELECT * FROM project_participant 
    		WHERE project_id = $1 AND username = $2
    	)`,
		id, username)
	if err != nil {
		return false, errors.Wrap(err, "failed to check if user is participant of a project")
	}
	return isParticipant, nil
}

func (r *projectRepositoryImpl) AddParticipant(id string, username string) error {
	exists, err := r.ExistsByID(id)
	if err != nil {
		return err
	}
	if !exists {
		return errors.Errorf("project with id %d does not exist", id)
	}
	_, err = r.db.Exec("INSERT INTO project_participant (project_id, username) VALUES ($1, $2)", id, username)
	if err != nil {
		return errors.Wrap(err, "failed to add new participant")
	}
	return nil
}

func (r *projectRepositoryImpl) RemoveParticipant(id string, username string) error {
	exists, err := r.ExistsByID(id)
	if err != nil {
		return err
	}
	if !exists {
		return errors.Errorf("project with id %d does not exist", id)
	}
	_, err = r.db.Exec("DELETE FROM project_participant WHERE project_id = $1 AND username = $2", id, username)
	if err != nil {
		return errors.Wrap(err, "failed to remove participant")
	}
	return nil
}
