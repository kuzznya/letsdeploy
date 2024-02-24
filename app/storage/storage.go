package storage

import (
	"context"
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type QueryExecDB interface {
	sqlx.Execer
	sqlx.Queryer
	sqlx.ExecerContext
	sqlx.QueryerContext
	Select(dest interface{}, query string, args ...interface{}) error
	Get(dest interface{}, query string, args ...interface{}) error
	NamedExec(query string, args interface{}) (sql.Result, error)
	NamedQuery(query string, args interface{}) (*sqlx.Rows, error)
}

type Storage struct {
	db QueryExecDB
}

func (s *Storage) ProjectRepository() ProjectRepository {
	return &projectRepositoryImpl{db: s.db}
}

func (s *Storage) ServiceRepository() ServiceRepository {
	return &serviceRepositoryImpl{db: s.db}
}

func (s *Storage) ManagedServiceRepository() ManagedServiceRepository {
	return &managedServiceRepositoryImpl{db: s.db}
}

func (s *Storage) SecretRepository() SecretRepository {
	return &secretRepositoryImpl{db: s.db}
}

func (s *Storage) ApiKeyRepository() ApiKeyRepository {
	return &apiKeyRepositoryImpl{db: s.db}
}

func (s *Storage) ExecTx(ctx context.Context, f func(*Storage) error) error {
	var tx *sqlx.Tx

	nestedTx := false

	switch s.db.(type) {
	case *sqlx.DB:
		var err error
		tx, err = s.db.(*sqlx.DB).BeginTxx(ctx, &sql.TxOptions{})
		if err != nil {
			return errors.Wrap(err, "tx creation failed")
		}
	case *sqlx.Tx:
		tx = s.db.(*sqlx.Tx)
		nestedTx = true
	}

	err := f(&Storage{db: tx})
	if err != nil {
		if err := tx.Rollback(); err != nil {
			log.WithError(err).Panicln("cannot rollback transaction")
		}
		return err
	}
	// do not commit tx to let parent tx decide whether to commit
	if nestedTx {
		return nil
	}
	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "cannot commit transaction")
	}
	return nil
}

type CrudRepository[T any, ID any] interface {
	CreateNew(record T) (ID, error)
	ExistsByID(id ID) (bool, error)
	FindByID(id ID) (*T, error)
	Update(record T) error
	Delete(id ID) error
}

func New(db *sqlx.DB) *Storage {
	return &Storage{db: db}
}
