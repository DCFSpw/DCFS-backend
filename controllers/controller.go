package controllers

import (
	"dcfs/middleware"
	"github.com/gin-gonic/gin"
	"log"
)

func ServeBackend() {
	r := gin.New()

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
		authorized.POST("/volume", CreateVolume)
		authorized.PUT("/volume/:VolumeUUID", UpdateVolume)
		authorized.DELETE("/volume/:VolumeUUID", DeleteVolume)
		authorized.GET("/volume/:VolumeUUID", GetVolume)

		authorized.POST("/volume/share/:VolumeUUID", ShareVolume)

		authorized.GET("/volume/volumes", GetVolumes)

		// disk
		authorized.POST("/disk", DiskCreate)
		authorized.PUT("/disk/:DiskUUID", DiskUpdate)
		authorized.DELETE("/disk/:DiskUUID", DiskDelete)
		authorized.GET("/disk/:DiskUUID", DiskGet)

		authorized.GET("/disk/disks", GetDisks)

		authorized.PUT("/disk/associate/:DiskUUID", DiskAssociate)
		authorized.DELETE("/disk/associate/:DiskUUID", DiskDissociate)

		// file
		authorized.POST("/file/:BlockUUID", FileUpload)
		authorized.GET("/file/download/:BlockUUID", FileDownload)

		authorized.PUT("/file/:FileUUID", FileRename)
		authorized.DELETE("/file/:FileUUID", FileRemove)
		authorized.GET("/file/:FileUUID", FileGet)

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
