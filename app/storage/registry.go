package storage

import "github.com/pkg/errors"

type ContainerRegistryEntity struct {
	Id        int    `db:"id"`
	ProjectId string `db:"project_id"`
	Url       string `db:"url"`
	Username  string `db:"username"`
	Password  string `db:"password"`
}

type ContainerRegistryRepository interface {
	FindByProjectId(projectId string) ([]ContainerRegistryEntity, error)
	CreateNew(registry ContainerRegistryEntity) (int, error)
	ExistsByProjectIdAndUrl(projectId string, url string) (bool, error)
	Delete(id int) error
}

type containerRegistryRepositoryImpl struct {
	db QueryExecDB
}

func (r containerRegistryRepositoryImpl) FindByProjectId(projectId string) ([]ContainerRegistryEntity, error) {
	registries := make([]ContainerRegistryEntity, 0)
	err := r.db.Select(&registries, "SELECT * FROM container_registry WHERE project_id = $1", projectId)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get project container registries")
	}
	return registries, nil
}

func (r containerRegistryRepositoryImpl) CreateNew(registry ContainerRegistryEntity) (int, error) {
	var id int
	err := r.db.Get(&id, "INSERT INTO container_registry (project_id, url, username, password) VALUES ($1, $2, $3, $4) RETURNING id",
		registry.ProjectId, registry.Url, registry.Username, registry.Password)
	if err != nil {
		return -1, errors.Wrap(err, "failed to create container registry")
	}
	return id, nil
}

func (r containerRegistryRepositoryImpl) ExistsByProjectIdAndUrl(projectId string, url string) (bool, error) {
	var exists bool
	err := r.db.Get(&exists, "SELECT exists(SELECT * FROM container_registry WHERE project_id = $1 AND url = $2)",
		projectId, url)
	if err != nil {
		return false, errors.Wrap(err, "failed to check if container registry exists")
	}
	return exists, nil
}

func (r containerRegistryRepositoryImpl) Delete(id int) error {
	_, err := r.db.Exec("DELETE FROM container_registry WHERE id = $1", id)
	if err != nil {
		return errors.Wrap(err, "failed to delete container registry")
	}
	return nil
}
