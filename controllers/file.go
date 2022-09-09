package controllers

import (
	"dcfs/apicalls"
	"dcfs/db"
	"dcfs/db/dbo"
	"dcfs/models/disk/GDriveDisk"
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

	// get disk from transport - to be implemented
	var _disk dbo.Disk = dbo.Disk{}
	var disk *GDriveDisk.GDriveDisk = GDriveDisk.NewGDriveDisk()
	var blockMetadata *apicalls.GDriveBlockMetadata = new(apicalls.GDriveBlockMetadata)

	db.DB.DatabaseHandle.Where("uuid = ?", "c91515a7-6c3c-4fb2-a82c-d3960d667ea3").Last(&_disk)
	disk.CreateCredentials(_disk.Credentials)
	disk.SetUUID(_disk.UUID)

	// Prepare block for upload
	contents := make([]uint8, blockHeader.Size)
	readSize, err := block.Read(contents)

	if err != nil || readSize != int(blockHeader.Size) {
		c.JSON(400, responses.SuccessResponse{Success: false, Msg: "Block loading failed"})
		return
	}

	blockMetadata.Ctx = c
	blockMetadata.Content = &contents
	blockMetadata.UUID = blockUUID
	blockMetadata.Size = blockHeader.Size

	disk.Upload(blockMetadata)

	/* SFTP demo
	var blockMetadata = apicalls.SFTPBlockMetadata{
		AbstractBlockMetadata: apicalls.AbstractBlockMetadata{
			UUID:    blockUUID,
			Size:    blockHeader.Size,
			Content: &contents}}

	// [TEMP] Upload block using SFTP
	var sftpDisk = FTPDisk.NewFTPDisk()
	sftpDisk.CreateCredentials("tester:password:192.168.1.176:21")
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
	*/

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
