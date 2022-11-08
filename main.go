package main

import (
	"dcfs/controllers"
	"dcfs/db"
	"dcfs/db/dbo"
	"dcfs/db/seeder"
	"dcfs/models"
	_ "dcfs/models/disk/FTPDisk"
	_ "dcfs/models/disk/GDriveDisk"
	_ "dcfs/models/disk/OneDriveDisk"
	_ "dcfs/models/disk/SFTPDisk"
	"flag"
	"github.com/google/uuid"
	"log"
	"os"
	"path/filepath"
)

func main() {
	models.RootUUID, _ = uuid.Parse("91c32303-0856-43d6-8e18-1cc671e256e4")
	// ignore error

	// remove downloads
	err := os.RemoveAll("./Download")
	if err != nil {
		log.Printf("Could not remove file the downloads dir")
	}

	path := flag.String("db-connection", "./connection.json", "file containing db connection info")
	rspw := flag.Bool("respawn", false, "set to true to drop and create the database anew")
	flag.Parse()

	absolutePath, err := filepath.Abs(*path)
	if err != nil {
		log.Fatal(err)
	}

	err = db.DB.Connect(absolutePath)
	if err != nil {
		log.Fatal(err)
	}

	// Register all needed tables
	db.DB.RegisterTable(dbo.Volume{})
	db.DB.RegisterTable(dbo.Provider{})
	db.DB.RegisterTable(dbo.File{})
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

	seeder.Seed()

	controllers.ServeBackend()
}
