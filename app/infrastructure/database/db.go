package database

import (
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/pkger"
	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/kuzznya/letsdeploy/app/util/validations"
	"github.com/markbates/pkger"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type DbConfig struct {
	Host     string
	Username string
	Password string
	Database string
}

func Setup(cfg *viper.Viper) *sqlx.DB {
	config := DbConfig{
		Host:     cfg.GetString("postgres.host"),
		Username: cfg.GetString("postgres.username"),
		Password: cfg.GetString("postgres.password"),
		Database: cfg.GetString("postgres.database"),
	}

	validations.MustBe(validations.NotEmptyString)(config.Host).
		OrPanicWithMessage("postgres.host is not defined")
	validations.MustBe(validations.NotEmptyString)(config.Username).
		OrPanicWithMessage("postgres.username is not defined")
	validations.MustBe(validations.NotEmptyString)(config.Password).
		OrPanicWithMessage("postgres.password is not defined")
	validations.MustBe(validations.NotEmptyString)(config.Database).
		OrPanicWithMessage("postgres.database is not defined")

	dbUrl := fmt.Sprintf("postgresql://%s:%s@%s/%s?sslmode=disable",
		config.Username, config.Password, config.Host, config.Database)
	log.Infof("Connecting to postgresql://%s:<password>@%s/%s\n",
		config.Username, config.Host, config.Database)
	db := sqlx.MustOpen("pgx", dbUrl)
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)

	log.Infoln("Successfully connected to PostgreSQL")

	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		log.WithError(err).Panicln("cannot create migration driver from existing sqlx DB")
	}

	migrationsModule := "/app/infrastructure/database/migrations"
	pkger.Include(migrationsModule)
	m, err := migrate.NewWithDatabaseInstance("pkger://"+migrationsModule, "postgres", driver)
	if err != nil {
		log.Panicln(fmt.Errorf("error creating migrate instance: %w", err))
	}
	err = m.Up()
	if err != nil {
		if err == migrate.ErrNoChange {
			log.Infoln("Database is up to date, no migrations executed")
		} else {
			log.WithError(err).Panicln("error running migrations")
		}
	} else {
		log.Infoln("Migrations executed")
	}
	return db
}
