package controllers

import (
	"dcfs/responses"
	"fmt"
	"github.com/gin-gonic/gin"
)

func FileRequest(c *gin.Context) {
	c.JSON(200, responses.SuccessResponse{Success: true, Msg: "File Request Endpoint"})
}

func FileUpload(c *gin.Context) {
	uuid := c.PostForm("blockUUID")
	file, _ := c.FormFile("block")

	fmt.Print(uuid)
	fmt.Print(file.Filename)

	c.JSON(200, responses.SuccessResponse{Success: true, Msg: "File Upload Endpoint"})
}

func FileRename(c *gin.Context) {
	c.JSON(200, responses.SuccessResponse{Success: true, Msg: "File Rename Endpoint"})
}

func FileRemove(c *gin.Context) {
	c.JSON(200, responses.SuccessResponse{Success: true, Msg: "File Remove Endpoint"})
}

func FileGet(c *gin.Context) {
	c.JSON(200, responses.SuccessResponse{Success: true, Msg: "File Get Endpoint"})
}

func FileDownload(c *gin.Context) {
	c.JSON(200, responses.SuccessResponse{Success: true, Msg: "File Download Endpoint"})
}

func FileShare(c *gin.Context) {
	c.JSON(200, responses.SuccessResponse{Success: true, Msg: "File Share Endpoint"})
}

func FileShareRemove(c *gin.Context) {
	c.JSON(200, responses.SuccessResponse{Success: true, Msg: "File Share Remove Endpoint"})
}

func FileRequestComplete(c *gin.Context) {
	c.JSON(200, responses.SuccessResponse{Success: true, Msg: "File Request Complete Endpoint"})
}

func GetFiles(c *gin.Context) {
	c.JSON(200, responses.SuccessResponse{Success: true, Msg: "Get Files Endpoint"})
}
