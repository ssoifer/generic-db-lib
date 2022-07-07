package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database"
	"github.com/golang-migrate/migrate/database/postgres"
	"log"
	"os"
	"reflect"
	"strconv"
	"time"
)

const (
	databaseExists     = `SELECT EXISTS(SELECT datname FROM pg_catalog.pg_database WHERE datname = '%s')`
	createDatabase     = `CREATE DATABASE %s`
	setTimezone        = `SET TIME ZONE 'UTC'`
	maxOpenConnections = 25
	maxIdleConnections = 25
	connMaxLifetime    = 5 // in minutes
)

// Config holds the configuration used for instantiating a new postgres.
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
}

var driverName = os.Getenv("DRIVER-NAME")

func NewDatabase(dbConfig Config) (*sql.DB, error) {
	ctx := context.Background()
	err := validateDBConfig(dbConfig)
	if err != nil {
		log.Fatal(ctx, err, "Error in validate db config")
		return nil, err
	}
	dbServer, err := waitForDatabase(dbConfig, connectionTimeout)
	if err != nil {
		log.Fatal(ctx, err, "Error while wait for database")
		return nil, err
	}
	return createDBAndConnect(dbServer, dbConfig)
}

func Migrate(db *sql.DB, config Config) error {
	ctx := context.Background()
	var driver database.Driver
	var err error
	switch driverName {

	case "postgres":
		driver, err = postgres.WithInstance(db, &postgres.Config{})

		//to support more drivers add more cases here
	}

	if err != nil {
		log.Print(ctx, err, "Error with migrate ")
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		migrationFolderPath,
		config.Database,
		driver,
	)
	if err != nil {
		log.Fatal(ctx, err, "Error with migrate")
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal(ctx, err, "Error with migrate")
		return err
	}

	ver, _, _ := m.Version()
	log.Print(ctx, "successfully applied migrations for database. current version %s", strconv.FormatUint(uint64(ver), 10))

	return nil
}

//
func createDBAndConnect(dbServer *sql.DB, dbConfig Config) (*sql.DB, error) {
	ctx := context.Background()
	err := createDB(dbServer, dbConfig)
	if err != nil {
		log.Fatal(ctx, err, "Error with creating db")
		return nil, err
	}
	err = dbServer.Close()
	if err != nil {
		log.Fatal(ctx, err, "Error with closing db server")
		return nil, err
	}

	db, err := openConnection(dbConfig)
	if err != nil {
		log.Fatal(ctx, err, "Error with open connection")
		return nil, err
	}
	return connect(db, dbConfig)
}

//
func validateDBConfig(dbConfig Config) error {
	ctx := context.Background()
	v := reflect.ValueOf(dbConfig)
	for i := 0; i < v.NumField(); i++ {
		if v.Field(i).Interface() == "" {
			log.Fatal(ctx, "all repository configuration fields must be set. Missing configuration for %s", v.Field(i).Interface())
			return errors.New("all repository configuration fields must be set")
		}
	}
	return nil
}

//
func toCreateDB(db *sql.DB, dbConfig Config) bool {
	ctx := context.Background()
	row := db.QueryRow(fmt.Sprintf(databaseExists, dbConfig.Database))
	var exists bool
	err := row.Scan(&exists)
	if err != nil {
		log.Fatal(ctx, err, "Error no rows ")
	}
	return err == nil && !exists
}

func createDB(db *sql.DB, dbConfig Config) error {
	ctx := context.Background()
	toCreateDB := toCreateDB(db, dbConfig)
	if toCreateDB {
		_, err := db.Exec(fmt.Sprintf(createDatabase, dbConfig.Database))
		if err != nil {
			log.Fatal(ctx, err, "unable to create database %s", dbConfig.Database)
			return err
		}
		log.Print(ctx, "created new database %s", dbConfig.Database)
	}
	log.Print(ctx, "setting timezone of database - %s", setTimezone)
	_, err := db.Exec(setTimezone)
	if err != nil {
		log.Fatal(ctx, err, "error in setting timezone of database - %s", setTimezone)
		return err
	}
	return nil
}

func retryConnection(dbConfig Config, timeout int) (*sql.DB, error) {
	ctx := context.Background()
	ticker := time.NewTicker(dbReconnectionTime * time.Second)
	defer ticker.Stop()

	timeoutExceeded := time.After(time.Duration(timeout) * time.Second)
	for {
		select {
		case <-timeoutExceeded:
			log.Fatal(ctx, "connection failed after %ds timeout", timeout)
			return nil, fmt.Errorf(" connection failed after %ds timeout", timeout)

		case <-ticker.C:
			db, err := openConnection(dbConfig)
			if err == nil {
				log.Print(ctx, "successfully opened database - %s", dbConfig.Database)
				return db, nil
			}
			log.Fatal(ctx, err, "unable to open database - %s, trying again", dbConfig.Database)
		}
	}
}

//
func waitForDatabase(dbConfig Config, timeout int) (*sql.DB, error) {
	ctx := context.Background()
	db, err := openConnection(dbConfig)

	// If connection fails, it may be because the repository container is not up yet.
	if err != nil {
		log.Fatal(ctx, err, "Connection  fails")
		db, err = retryConnection(dbConfig, timeout)
	}
	return db, err
}

//
func openConnection(dbConfig Config) (*sql.DB, error) {
	ctx := context.Background()
	if driverName == "" {
		//error - missing env var
	}
	db, err := sql.Open(driverName, fmt.Sprintf(
		"user=%s password=%s dbname=%s host=%s port=%s sslmode=disable",
		dbConfig.User, dbConfig.Password, dbConfig.Database, dbConfig.Host, dbConfig.Port))
	if err != nil {
		log.Print(ctx, err, "Incorrect formatting string")
		return nil, errors.New("incorrect formatting string")
	}
	err = db.Ping()
	if err != nil {
		log.Print(ctx, err, "Can't ping server")
		return nil, errors.New("can't ping server")
	}
	return nil, nil
}

func connect(db *sql.DB, dbConfig Config) (*sql.DB, error) {
	ctx := context.Background()
	// Ping verifies if the connection to the repository is alive
	err := db.Ping()
	if err != nil {
		log.Fatal(ctx, err, "database %s ping failed", dbConfig.Database)
		return nil, err
	}
	db.SetMaxOpenConns(maxOpenConnections)
	db.SetMaxIdleConns(maxIdleConnections)
	db.SetConnMaxLifetime(connMaxLifetime * time.Minute)
	log.Print(ctx, "connected to database", dbConfig.Database)
	return db, nil
}
