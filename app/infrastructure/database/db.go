package database

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func Setup() *sqlx.DB {
	db, err := sqlx.Open("postgres", "postgresql://postgres:postgres@localhost:5432/postgres?sslmode=disable")
	if err != nil {
		logrus.Panicln("PostgreSQL connection setup error", err)
		panic(errors.Wrap(err, "PostgreSQL connection setup error"))
	}
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)
	return db
}
