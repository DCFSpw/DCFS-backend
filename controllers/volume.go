package controllers

import (
	"dcfs/responses"
	"github.com/gin-gonic/gin"
)

func CreateVolume(c *gin.Context) {
	c.JSON(200, responses.SuccessResponse{Success: true, Msg: "Create Volume Endpoint"})
}

func UpdateVolume(c *gin.Context) {
	c.JSON(200, responses.SuccessResponse{Success: true, Msg: "Update Volume Endpoint"})
}

func DeleteVolume(c *gin.Context) {
	c.JSON(200, responses.SuccessResponse{Success: true, Msg: "Delete Volume Endpoint"})
}

func GetVolume(c *gin.Context) {
	c.JSON(200, responses.SuccessResponse{Success: true, Msg: "Get Volume Endpoint"})
}

func ShareVolume(c *gin.Context) {
	c.JSON(200, responses.SuccessResponse{Success: true, Msg: "Share Volume Endpoint"})
}

func GetVolumes(c *gin.Context) {
	c.JSON(200, responses.SuccessResponse{Success: true, Msg: "Get Volumes Endpoint"})
}
