package projects

import (
	"github.com/kuzznya/letsdeploy/common/database"
	"github.com/pkg/errors"
)

type ProjectRecord struct {
	Id     int    `db:"id" json:"id,omitempty"`
	Name   string `db:"name" json:"name"`
	Author string `db:"author" json:"author"`
}

func FindUserProjects(username string) ([]ProjectRecord, error) {
	db := database.GetDB()
	records := []ProjectRecord{}
	err := db.Select(
		&records,
		"SELECT p.* FROM project p "+
			"JOIN project_participant pp ON pp.project_id = p.id "+
			"WHERE pp.username = $1",
		username)
	if err != nil {
		return nil, errors.Wrap(err, "Cannot execute query")
	}
	return records, nil
}

func FindProjectParticipants(projectId int) ([]string, error) {
	db := database.GetDB()
	participants := []string{}
	err := db.Select(participants,
		"SELECT pp.username FROM project_participant pp WHERE pp.project_id = $1",
		projectId)
	if err != nil {
		return nil, errors.Wrap(err, "Cannot execute query")
	}
	return participants, nil
}
