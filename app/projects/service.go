package projects

import (
	"fmt"
	"github.com/pkg/errors"
)

func GetUserProjects(username string) ([]Project, error) {
	records, _ := FindUserProjects(username)
	projects := []Project{}
	for _, record := range records {
		participants, err := FindProjectParticipants(record.Id)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("Cannot load project %d participants", record.Id))
		}
		project := Project{
			Id:           record.Id,
			Name:         record.Name,
			Participants: participants,
		}
		projects = append(projects, project)
	}
	return projects, nil
}
