package db

import (
	"dcfs/util/logger"
	"reflect"
)

// MigrateAll - migrate all dbo models to database
//
// This function performs AutoMigrate() for every dbo model in this project.
// Needs to be updated for every new dbo model.
//
// return type:
//   - error
func (db *DatabaseConnection) MigrateAll() error {
	for _, value := range db.Tables {
		err := db.DatabaseHandle.AutoMigrate(value)
		if err != nil {
			logger.Logger.Error("db", "Failed to migrate the table: ", reflect.TypeOf(value).Name())
			return err
		}
	}

	logger.Logger.Debug("db", "Successfully migrated all the tables to the db.")
	return nil
}

// Respawn - drop the entire database and create everything anew
//
// return type:
//   - error
func (db *DatabaseConnection) Respawn() error {
	db.DatabaseHandle.Exec("DROP DATABASE IF EXISTS " + db.connectionInfo.DbName)
	db.DatabaseHandle.Exec("CREATE DATABASE " + db.connectionInfo.DbName)

	err := db.MigrateAll()
	if err != nil {
		logger.Logger.Error("db", "Failed to migrate the tables into the database with error: ", err.Error())
		return err
	}

	logger.Logger.Debug("db", "Successfully respawned the db.")
	return nil
}
