package controllers

import (
	"dcfs/apicalls"
	"dcfs/models/disk/SFTPDisk"
	"dcfs/responses"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func FileRequest(c *gin.Context) {
	c.JSON(200, responses.SuccessResponse{Success: true, Msg: "File Request Endpoint"})
}

func FileUpload(c *gin.Context) {
	// Get data from request
	uuidString := c.PostForm("blockUUID")
	blockHeader, blockHeaderError := c.FormFile("block")
	if blockHeaderError != nil {
		c.JSON(400, responses.SuccessResponse{Success: false, Msg: "Missing block"})
		return
	}
	block, blockError := blockHeader.Open()
	if blockError != nil {
		c.JSON(400, responses.SuccessResponse{Success: false, Msg: "Block opening failed"})
		return
	}

	// Validate data
	if uuidString == "" {
		c.JSON(400, responses.SuccessResponse{Success: false, Msg: "Missing blockUUID"})
		return
	}
	blockUUID, uuidError := uuid.Parse(uuidString)
	if uuidError != nil {
		c.JSON(400, responses.SuccessResponse{Success: false, Msg: "Invalid blockUUID"})
		return
	}

	// Prepeare block for upload
	contents := make([]uint8, blockHeader.Size)
	readSize, err := block.Read(contents)

	if err != nil || readSize != int(blockHeader.Size) {
		c.JSON(400, responses.SuccessResponse{Success: false, Msg: "Block loading failed"})
		return
	}

	var blockMetadata = apicalls.BlockMetadata{}
	blockMetadata.UUID = blockUUID
	blockMetadata.Size = blockHeader.Size
	blockMetadata.Content = &contents

	// [TEMP] Upload block using SFTP
	var sftpDisk = SFTPDisk.NewSFTPDisk()
	sftpDisk.CreateCredentials("tester:password:192.168.1.176:2222")
	err = sftpDisk.Connect(nil)
	if err != nil {
		c.JSON(404, responses.SuccessResponse{Success: false, Msg: "SFTP connection failed"})
		return
	}
	err = sftpDisk.Upload(&blockMetadata)
	if err != nil {
		c.JSON(404, responses.SuccessResponse{Success: false, Msg: "SFTP upload failed"})
		return
	}

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
