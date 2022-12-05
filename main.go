package main

import (
	"dcfs/controllers"
	"dcfs/db"
	"dcfs/db/dbo"
	"dcfs/db/seeder"
	"dcfs/models"
	"dcfs/util/logger"
	"flag"
	"github.com/google/uuid"
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "dcfs/models/disk/FTPDisk"
	_ "dcfs/models/disk/GDriveDisk"
	_ "dcfs/models/disk/OneDriveDisk"
	_ "dcfs/models/disk/SFTPDisk"
)

// main - entry point for the server application
func main() {
	// Prepare sample account UUID for development purposes
	models.RootUUID, _ = uuid.Parse("91c32303-0856-43d6-8e18-1cc671e256e4")

	// Remove local downloaded files
	err := os.RemoveAll("./Download")
	if err != nil {
		log.Printf("Could not remove file the downloads directory")
	}

	// Parse settings and options
	path := flag.String("db-connection", "./connection.json", "file containing db connection info")
	rspw := flag.Bool("respawn", false, "set to true to drop and create the database anew")
	debugLevel := flag.Int("debug", 1, "debug level: 2 - debug, warnings and errors, 1 - warnings and errors, 0 - errors, -1 - none, default: 1")
	logScope := flag.String("log", "", "a comma separated list of modules to collect logs from, available are: middleware, api, db, disks, credentials, file, partitioner, transport, volume. The option: all enables logs from all modules")
	flag.Parse()

	logger.Logger.SetLogLevel(*debugLevel)
	logger.Logger.SetScopes(strings.Split(*logScope, ","))

	absolutePath, err := filepath.Abs(*path)
	if err != nil {
		log.Fatal(err)
	}

	// Connect to database
	err = db.DB.Connect(absolutePath)
	if err != nil {
		log.Fatal(err)
	}

	// Register all needed tables
	db.DB.RegisterTable(dbo.Volume{})
	db.DB.RegisterTable(dbo.Provider{})
	db.DB.RegisterTable(dbo.File{})
	db.DB.RegisterTable(dbo.VirtualDisk{})
	db.DB.RegisterTable(dbo.Disk{})
	db.DB.RegisterTable(dbo.Block{})
	db.DB.RegisterTable(dbo.User{})
	db.DB.RegisterTable(dbo.Provider{})

	if *rspw {
		err = db.DB.Respawn()
		if err != nil {
			log.Fatal(err)
		}
	} else {
		err = db.DB.MigrateAll()
		if err != nil {
			log.Fatal(err)
		}
	}

	// Seed required data
	seeder.Seed()

	// Serve API backend using Gin framework
	controllers.ServeBackend()
}
