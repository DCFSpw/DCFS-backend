package controllers

import (
	"dcfs/responses"
	"github.com/gin-gonic/gin"
)

func DiskCreate(c *gin.Context) {
	c.JSON(200, responses.SuccessResponse{Success: true, Msg: "Disk Create Endpoint"})
}

func DiskGet(c *gin.Context) {
	c.JSON(200, responses.SuccessResponse{Success: true, Msg: "Disk Get Endpoint"})
}

func DiskUpdate(c *gin.Context) {
	c.JSON(200, responses.SuccessResponse{Success: true, Msg: "Disk Update Endpoint"})
}

func DiskDelete(c *gin.Context) {
	c.JSON(200, responses.SuccessResponse{Success: true, Msg: "Disk Delete Endpoint"})
}

func GetDisks(c *gin.Context) {
	c.JSON(200, responses.SuccessResponse{Success: true, Msg: "Get Disks Endpoint"})
}

func DiskAssociate(c *gin.Context) {
	c.JSON(200, responses.SuccessResponse{Success: true, Msg: "Disk Associate Endpoint"})
}

func DiskDissociate(c *gin.Context) {
	c.JSON(200, responses.SuccessResponse{Success: true, Msg: "Disk Dissociate Endpoint"})
}
