package controllers

import (
	"dcfs/apicalls"
	"dcfs/db"
	"dcfs/db/dbo"
	"dcfs/models/disk/GDriveDisk"
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
	//fileUUIDString := c.Param("FileUUID")
	blockUUIDString := c.PostForm("blockUUID")
	blockHeader, blockHeaderError := c.FormFile("block")
	if blockHeaderError != nil {
		c.JSON(422, responses.ValidationErrorResponse{Success: false, Msg: "Missing block"})
		return
	}
	block, blockError := blockHeader.Open()
	if blockError != nil {
		c.JSON(500, responses.OperationFailureResponse{Success: false, Msg: "Block opening failed: " + blockError.Error()})
		return
	}

	// Validate data
	if blockUUIDString == "" {
		c.JSON(422, responses.ValidationErrorResponse{Success: false, Msg: "Missing blockUUID"})
		return
	}
	blockUUID, uuidError := uuid.Parse(blockUUIDString)
	if uuidError != nil {
		c.JSON(422, responses.ValidationErrorResponse{Success: false, Msg: "Invalid blockUUID"})
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
		c.JSON(500, responses.OperationFailureResponse{Success: false, Msg: "Block loading failed: " + err.Error()})
		return
	}

	blockMetadata.Ctx = c
	blockMetadata.Content = &contents
	blockMetadata.UUID = blockUUID
	blockMetadata.Size = blockHeader.Size

	err = disk.Upload(blockMetadata)
	if err != nil {
		c.JSON(500, responses.OperationFailureResponse{Success: false, Msg: "Block uploading failed: " + err.Error()})
		return
	}

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
	// Get data from request
	//fileUUIDString := c.Param("FileUUID")
	blockUUIDString := c.PostForm("blockUUID")

	// Validate data
	if blockUUIDString == "" {
		c.JSON(422, responses.ValidationErrorResponse{Success: false, Msg: "Missing blockUUID"})
		return
	}

	blockUUID, uuidError := uuid.Parse(blockUUIDString)
	if uuidError != nil {
		c.JSON(422, responses.ValidationErrorResponse{Success: false, Msg: "Invalid blockUUID"})
		return
	}

	// FTP demo
	var blockMetadata = apicalls.SFTPBlockMetadata{
		AbstractBlockMetadata: apicalls.AbstractBlockMetadata{
			UUID: blockUUID}}

	var ftpDisk = SFTPDisk.NewSFTPDisk()
	ftpDisk.CreateCredentials("...")
	err := ftpDisk.Connect(nil)
	if err != nil {
		c.JSON(500, responses.OperationFailureResponse{Success: false, Msg: "FTP connection failed: " + err.Error()})
		return
	}

	err = ftpDisk.Download(&blockMetadata)
	if err != nil {
		c.JSON(500, responses.OperationFailureResponse{Success: false, Msg: "Block download failed: " + err.Error()})
		return
	}

	c.JSON(200, responses.BlockDownloadResponse{Success: true, Msg: "File Download Endpoint", Block: *blockMetadata.Content})
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
