package storage

import "github.com/pkg/errors"

type ProjectDBModel struct {
	Id   int    `db:"id"`
	Name string `db:"name"`
}

type ProjectRepository interface {
	CrudRepository[ProjectDBModel, int]
	FindUserProjects(username string) ([]ProjectDBModel, error)
	GetParticipants(id int) ([]string, error)
	IsParticipant(id int, username string) (bool, error)
	AddParticipant(id int, username string) error
	RemoveParticipant(id int, username string) error
}

type projectRepositoryImpl struct {
	db QueryExecDB
}

func (p *projectRepositoryImpl) CreateNew(project ProjectDBModel) (int, error) {
	var result int
	err := p.db.Get(&result, "INSERT INTO project (name) VALUES ($1) RETURNING id", project.Name)
	if err != nil {
		return 0, errors.Wrap(err, "cannot save new project")
	}
	return result, nil
}

func (p *projectRepositoryImpl) ExistsByID(id int) (bool, error) {
	var exists bool
	err := p.db.Get(&exists, "SELECT exists(SELECT * FROM project WHERE id = $1)", id)
	if err != nil {
		return false, errors.Wrap(err, "cannot check if project exists")
	}
	return exists, nil
}

func (p *projectRepositoryImpl) FindByID(id int) (*ProjectDBModel, error) {
	var project ProjectDBModel
	err := p.db.Get(&project, "SELECT * FROM project WHERE id = $1", id)
	if err != nil {
		return nil, errors.Wrap(err, "cannot find project by id")
	}
	return &project, nil
}

func (p *projectRepositoryImpl) Update(project ProjectDBModel) error {
	_, err := p.db.Exec("UPDATE project SET name = $1 WHERE id = $2", project.Name, project.Id)
	if err != nil {
		return errors.Wrap(err, "failed to update project")
	}
	return nil
}

func (p *projectRepositoryImpl) Delete(id int) error {
	_, err := p.db.Exec("DELETE FROM project WHERE id = $1", id)
	if err != nil {
		return errors.Wrap(err, "failed to delete project")
	}
	return nil
}

func (p *projectRepositoryImpl) FindUserProjects(username string) ([]ProjectDBModel, error) {
	projects := []ProjectDBModel{}
	err := p.db.Select(&projects,
		`SELECT p.* FROM project p 
    	JOIN project_participant pp ON pp.project_id = p.id 
        WHERE pp.username = $1`,
		username)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve user's projects")
	}
	return projects, nil
}

func (p *projectRepositoryImpl) GetParticipants(id int) ([]string, error) {
	exists, err := p.ExistsByID(id)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.Errorf("project with id %d does not exist", id)
	}
	participants := []string{}
	err = p.db.Select(&participants, "SELECT username FROM project_participant WHERE project_id = $1", id)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get project participants")
	}
	return participants, nil
}

func (p *projectRepositoryImpl) IsParticipant(id int, username string) (bool, error) {
	exists, err := p.ExistsByID(id)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, nil
	}
	var isParticipant bool
	err = p.db.Get(
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

func (p *projectRepositoryImpl) AddParticipant(id int, username string) error {
	exists, err := p.ExistsByID(id)
	if err != nil {
		return err
	}
	if !exists {
		return errors.Errorf("project with id %d does not exist", id)
	}
	_, err = p.db.Exec("INSERT INTO project_participant (project_id, username) VALUES ($1, $2)", id, username)
	if err != nil {
		return errors.Wrap(err, "failed to add new participant")
	}
	return nil
}

func (p *projectRepositoryImpl) RemoveParticipant(id int, username string) error {
	exists, err := p.ExistsByID(id)
	if err != nil {
		return err
	}
	if !exists {
		return errors.Errorf("project with id %d does not exist", id)
	}
	_, err = p.db.Exec("DELETE FROM project_participant WHERE project_id = $1 AND username = $2", id, username)
	if err != nil {
		return errors.Wrap(err, "failed to remove participant")
	}
	return nil
}
