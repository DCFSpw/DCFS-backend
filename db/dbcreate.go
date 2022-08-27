package db

import (
	"fmt"
	"reflect"
)

// MigrateAll - performs AutoMigrate() for every dbo model in this project (needs to be updated for every new dbo model)
func (db *DatabaseConnection) MigrateAll() error {
	for _, value := range db.Tables {
		err := db.databaseHandle.AutoMigrate(value)
		if err != nil {
			fmt.Println("Failed to migrate the table: ", reflect.TypeOf(value).Name())
			return err
		}
	}

	return nil
}

// Respawn - drop the entire database and create everything anew
func (db *DatabaseConnection) Respawn() error {
	db.databaseHandle.Exec("DROP DATABASE IF EXISTS " + db.connectionInfo.DbName)
	db.databaseHandle.Exec("CREATE DATABASE " + db.connectionInfo.DbName)

	err := db.MigrateAll()
	if err != nil {
		fmt.Println("Failed to migrate the tables into the database with error: ", err)
		return err
	}

	return nil
}
