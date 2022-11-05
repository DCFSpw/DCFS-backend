package controllers

import (
	"dcfs/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"log"
	"time"
)

func ServeBackend() {
	r := gin.New()

	// Cors
	corsConfig := cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}
	r.Use(cors.New(corsConfig))

	// TODO: rethink logger here
	r.Use(gin.Logger())

	// Recovery middleware recovers from any panics and writes a 500 if there was one.
	r.Use(gin.Recovery())

	// Unauthorized requests
	// 	Authorization
	unauthorized := r.Group("/")
	unauthorized.POST("/auth/register", RegisterUser)
	unauthorized.POST("/auth/login", LoginUser)

	// volume

	// disk

	// file

	// Authorized requests
	authorized := r.Group("/")
	authorized.Use(middleware.Authenticate())
	{
		// Account settings
		authorized.GET("/user/profile", GetUserProfile)
		authorized.PUT("/user/profile", UpdateUserProfile)
		authorized.PUT("/user/password", ChangeUserPassword)

		// Volume
		authorized.POST("/volumes/manage", CreateVolume)
		authorized.GET("/volumes/manage", GetVolumes)
		authorized.GET("/volumes/manage/:VolumeUUID", GetVolume)
		authorized.PUT("/volumes/manage/:VolumeUUID", UpdateVolume)
		authorized.DELETE("/volumes/manage/:VolumeUUID", DeleteVolume)

		// Disk
		authorized.POST("/disks/manage", DiskCreate)
		authorized.GET("/disks/manage", GetDisks)

		authorized.GET("/disks/manage/:DiskUUID", DiskGet)
		authorized.PUT("/disks/manage/:DiskUUID", DiskUpdate)
		authorized.DELETE("/disks/manage/:DiskUUID", DiskDelete)

		authorized.POST("/disks/oauth/:DiskUUID", DiskOAuth)

		// File
		authorized.POST("/files/manage", CreateDirectory)

		//authorized.POST("/file/io/:FileUUID", FileUpload)
		//authorized.GET("/file/io/:FileUUID", FileDownload)

		authorized.POST("/file/request", FileRequest)
		authorized.POST("/file/request/complete/:FileUUID", FileRequestComplete)

		//authorized.GET("/file/request", FileRequest)
		//authorized.POST("/file/request/complete/:FileUUID", FileRequestComplete)

		//authorized.POST("/file/share/:FileUUID", FileShare)
		//authorized.DELETE("/file/share/FileUUID", FileShareRemove)

		authorized.GET("/file/files", GetFiles)

		// Providers
		authorized.GET("/providers", GetProviders)
	}

	// Listen and serve on 0.0.0.0:8080
	err := r.Run(":8080")
	if err != nil {
		log.Fatal(err)
	}
}
