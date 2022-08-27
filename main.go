package main

import (
	"dcfs/db"
	"dcfs/db/dbo"
	"flag"
	"log"
	"path/filepath"
)

func main() {
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
}
