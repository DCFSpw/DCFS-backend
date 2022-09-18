package controllers

import (
	"dcfs/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"log"
)

func ServeBackend() {
	r := gin.New()

	// Cors
	r.Use(cors.Default())

	// TODO: rethink logger here
	r.Use(gin.Logger())

	// Recovery middleware recovers from any panics and writes a 500 if there was one.
	r.Use(gin.Recovery())

	// unauthorized requests

	// volume

	// disk

	// file

	// authorized requests
	authorized := r.Group("/")
	authorized.Use(middleware.Authenticate())
	{
		// volume
		authorized.POST("/volume/manage", CreateVolume)
		authorized.PUT("/volume/manage/:VolumeUUID", UpdateVolume)
		authorized.DELETE("/volume/manage/:VolumeUUID", DeleteVolume)
		authorized.GET("/volume/manage/:VolumeUUID", GetVolume)

		authorized.POST("/volume/share/:VolumeUUID", ShareVolume)

		authorized.GET("/volume/volumes", GetVolumes)

		// disk
		authorized.POST("/disks/manage", DiskCreate)
		authorized.GET("/disks/manage", GetDisks)

		authorized.GET("/disks/manage/:DiskUUID", DiskGet)
		authorized.PUT("/disks/manage/:DiskUUID", DiskUpdate)
		authorized.DELETE("/disks/manage/:DiskUUID", DiskDelete)

		authorized.POST("/disks/association/:DiskUUID", DiskAssociate)
		authorized.DELETE("/disks/association/:DiskUUID", DiskDissociate)

		authorized.POST("/disks/oauth/:DiskUUID", DiskOAuth)

		// file
		authorized.POST("/file/io/:FileUUID", FileUpload)
		authorized.GET("/file/io/:FileUUID", FileDownload)

		authorized.PUT("/file/manage/:FileUUID", FileRename)
		authorized.DELETE("/file/manage/:FileUUID", FileRemove)
		authorized.GET("/file/manage/:FileUUID", FileGet)

		authorized.GET("/file/request", FileRequest)
		authorized.POST("/file/request/complete/:FileUUID", FileRequestComplete)

		authorized.POST("/file/share/:FileUUID", FileShare)
		authorized.DELETE("/file/share/FileUUID", FileShareRemove)

		authorized.GET("/file/files", GetFiles)
	}

	// Listen and serve on 0.0.0.0:8080
	err := r.Run(":8080")
	if err != nil {
		log.Fatal(err)
	}
}
