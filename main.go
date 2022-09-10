package main

import (
	"dcfs/controllers"
	"dcfs/db"
	"dcfs/db/dbo"
	"dcfs/models/disk"
	"flag"
	"github.com/google/uuid"
	"log"
	"path/filepath"
)

func main() {
	disk.RootUUID, _ = uuid.Parse("91c32303-0856-43d6-8e18-1cc671e256e4")
	// ignore error

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
	db.DB.RegisterTable(dbo.File{})
	db.DB.RegisterTable(dbo.Disk{})
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

	//db.Seed()

	controllers.ServeBackend()
}
