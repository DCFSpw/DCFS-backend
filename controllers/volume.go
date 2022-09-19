package controllers

import (
	"dcfs/responses"
	"github.com/gin-gonic/gin"
)

func CreateVolume(c *gin.Context) {
	c.JSON(200, responses.SuccessResponse{Success: true, Message: "Create Volume Endpoint"})
}

func UpdateVolume(c *gin.Context) {
	c.JSON(200, responses.SuccessResponse{Success: true, Message: "Update Volume Endpoint"})
}

func DeleteVolume(c *gin.Context) {
	c.JSON(200, responses.SuccessResponse{Success: true, Message: "Delete Volume Endpoint"})
}

func GetVolume(c *gin.Context) {
	c.JSON(200, responses.SuccessResponse{Success: true, Message: "Get Volume Endpoint"})
}

func ShareVolume(c *gin.Context) {
	c.JSON(200, responses.SuccessResponse{Success: true, Message: "Share Volume Endpoint"})
}

func GetVolumes(c *gin.Context) {
	c.JSON(200, responses.SuccessResponse{Success: true, Message: "Get Volumes Endpoint"})
}
