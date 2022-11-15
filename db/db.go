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

// Connect - connect to a databased, specified in the provided JSON file
//
// params:
//   - filepath string: path to the JSON file containing the needed connection string
//
// return type:
//   - error: nil when no error occurred
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

// RegisterTable - register model as a table in database
//
// params:
//   - obj dbo.DatabaseObject: abstract database object representing one of the models
func (db *DatabaseConnection) RegisterTable(obj dbo.DatabaseObject) {
	db.Tables = append(db.Tables, obj)
}

// DB - a global database object
var DB *DatabaseConnection = new(DatabaseConnection)
