package db

import (
	"dcfs/db/dbo"
	"encoding/json"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"io"
	"os"
)

type connection struct {
	Username       string `json:"username"`
	Password       string `json:"password"`
	ConnectionType string `json:"connectionType"`
	Address        string `json:"address"`
	DbName         string `json:"dbName"`
}

// DatabaseConnection - object handling connecting to and querying the database
//
// fields:
//   - DatabaseHandle
//   - connectionInfo: contains connection string required to connect to a db
type DatabaseConnection struct {
	DatabaseHandle *gorm.DB
	connectionInfo connection
	Tables         []dbo.DatabaseObject
}

/* private methods */

func (db *DatabaseConnection) parseConnection(filepath string) error {
	jsonFile, err := os.Open(filepath)
	if err != nil {
		fmt.Println("Failed to open the file: ", filepath, " with err: ", err)
		return err
	}

	defer jsonFile.Close()

	var byteValue []byte
	byteValue, err = io.ReadAll(jsonFile)
	if err != nil {
		fmt.Println("Failed to read the ", filepath, " file with err: ", err)
		return err
	}

	err = json.Unmarshal(byteValue, &db.connectionInfo)
	if err != nil {
		fmt.Println("Could not unmarshal json file: ", filepath, " with err: ", err)
		return err
	}

	return nil
}

/* public methods */

// Connect - connect to a db, specified in the provided json file
//
// params:
//   - filepath string: path to the json file containing the needed connection string
//
// return type:
//   - error (nil when no error occurred)
func (db *DatabaseConnection) Connect(filepath string) error {
	err := db.parseConnection(filepath)
	if err != nil {
		fmt.Println("Could not parse file: ", filepath, " with err: ", err)
		return err
	}

	dsn :=
		db.connectionInfo.Username +
			":" +
			db.connectionInfo.Password +
			"@" +
			db.connectionInfo.ConnectionType +
			"(" +
			db.connectionInfo.Address +
			")/" +
			db.connectionInfo.DbName +
			"?charset=utf8mb4&parseTime=True&loc=Local"

	db.DatabaseHandle, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Println("Could not connect to the database with the error: ", err)
		return err
	}

	return nil
}

// PerformOperation - perform a db operation on the database handle
// DEPRECATED - TO BE DELETED - db handle is made public instead
//
// params:
//   - operation func(handle *gorm.DB) error: a function that performs the desired operation on the db handle
//
// return type:
//   - error (nil on success)
func (db *DatabaseConnection) PerformOperation(operation func(handle *gorm.DB) error) error {
	err := operation(db.DatabaseHandle)
	if err != nil {
		fmt.Println("Could not perform a database operation with error: ", err)
		return err
	}

	return nil
}

func (db *DatabaseConnection) RegisterTable(obj dbo.DatabaseObject) {
	db.Tables = append(db.Tables, obj)
}

// DB - a global db object
var DB *DatabaseConnection = new(DatabaseConnection)
