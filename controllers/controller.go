package controllers

import (
	"dcfs/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"log"
	"time"
)

// ServeBackend - serve API backend using Gin framework
func ServeBackend() {
	r := gin.New()

	// Cors configuration
	corsConfig := cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}
	r.Use(cors.New(corsConfig))

	// Logger middleware for printing logs
	r.Use(gin.Logger())
	r.Use(middleware.LogApi())

	// Recovery middleware recovers from any panics and writes a 500 if there was one.
	r.Use(gin.Recovery())

	// Unauthorized requests
	unauthorized := r.Group("/")
	unauthorized.POST("/auth/register", RegisterUser)
	unauthorized.POST("/auth/login", LoginUser)

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
		authorized.POST("/disks/manage", CreateDisk)
		authorized.GET("/disks/manage", GetDisks)

		authorized.GET("/disks/manage/:DiskUUID", GetDisk)
		authorized.PUT("/disks/manage/:DiskUUID", UpdateDisk)
		authorized.DELETE("/disks/manage/:DiskUUID", DeleteDisk)
		authorized.DELETE("/disks/backup/:DiskUUID", ReplaceBackupDisk)

		authorized.POST("/disks/oauth/:DiskUUID", DiskOAuth)

		// File
		authorized.POST("/files/manage", CreateDirectory)

		authorized.GET("/files/manage/:FileUUID", GetFile)
		authorized.GET("/files/manage", GetFiles)

		authorized.POST("/files/upload", InitFileUploadRequest)
		authorized.POST("/files/download/:FileUUID", InitFileDownloadRequest)
		authorized.POST("/files/upload/:FileUUID", CompleteFileUploadRequest)

		authorized.POST("/files/block/:BlockUUID", UploadBlock)
		authorized.GET("/files/block/:BlockUUID", DownloadBlock)

		authorized.PUT("/files/manage/:FileUUID", UpdateFile)
		authorized.DELETE("/files/manage/:FileUUID", DeleteFile)

		// Providers
		authorized.GET("/providers", GetProviders)
	}

	// Listen and serve on localhost:8080
	err := r.Run(":8080")
	if err != nil {
		log.Fatal(err)
	}
}
