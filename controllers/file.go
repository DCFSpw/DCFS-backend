package controllers

import (
	"dcfs/apicalls"
	"dcfs/db"
	"dcfs/db/dbo"
	"dcfs/middleware"
	"dcfs/models"
	"dcfs/models/disk/SFTPDisk"
	"dcfs/responses"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"log"
	"time"
)

type RequestedFile struct {
	Name string `json:"name"`
	Type int    `json:"type"`
	Size int    `json:"size"`
}

type FileRequestRequest struct {
	VolumeUUID string        `json:"volumeUUID"`
	File       RequestedFile `json:"file"`
}

func FileRequest(c *gin.Context) {
	var req FileRequestRequest
	var fileUploadRequest apicalls.FileUploadRequest = apicalls.FileUploadRequest{}
	var file models.File
	var volume *models.Volume
	var err error

	userData, _ := c.Get("UserData")
	userUUID := userData.(middleware.UserData).UserUUID

	err = c.BindJSON(&req)
	if err != nil {
		panic("unimplemented")
	}

	volumeUUID, err := uuid.Parse(req.VolumeUUID)
	if err != nil {
		// TODO
		panic("unimplemented")
	}

	volume = models.Transport.GetVolume(userUUID, volumeUUID)

	err = c.Bind(&req)
	if err != nil {
		// TODO
		panic("unimplemented")
	}
	log.Println("Got a request for a file named: ", req.File.Name, "of size: ", req.File.Size)

	fileUploadRequest.Size = req.File.Size
	fileUploadRequest.Type = req.File.Type
	fileUploadRequest.Name = req.File.Name
	fileUploadRequest.UserUUID = userUUID

	file = volume.FileUploadRequest(&fileUploadRequest)
	models.Transport.EnqueueFileUpload(file.GetUUID(), file)

	var blocks []responses.FileRequestBlockResponse
	var rsp responses.FileRequestResponse = responses.FileRequestResponse{
		SuccessResponse: responses.SuccessResponse{Success: true, Message: "File Request Successful"},
		UUID:            file.GetUUID().String(),
		Name:            file.GetName(),
		Type:            file.GetType(),
		Size:            file.GetSize(),
	}

	if file.GetType() != dbo.FILE_TYPE_DIRECTORY {
		var _file *models.RegularFile = file.(*models.RegularFile)
		for _, block := range _file.Blocks {
			blocks = append(blocks, responses.FileRequestBlockResponse{
				UUID:  block.UUID.String(),
				Order: block.Order,
				Size:  block.Size,
			})
		}
	}

	rsp.Blocks = blocks

	log.Println("Prepared a request with ", len(blocks), " blocks")
	c.JSON(200, rsp)
}

type FileUploadBody struct {
}

func FileUpload(c *gin.Context) {
	var fileUUID uuid.UUID
	var blockUUID uuid.UUID
	var _fileUUID string
	var _blockUUID string
	var err error

	// Get data from request
	_fileUUID = c.Param("FileUUID")
	_blockUUID = c.PostForm("blockUUID")
	blockHeader, err := c.FormFile("block")
	if err != nil {
		c.JSON(422, responses.ValidationErrorResponse{Success: false, Message: "Missing block"})
		return
	}
	block, blockError := blockHeader.Open()
	if blockError != nil {
		c.JSON(500, responses.OperationFailureResponse{Success: false, Message: "Block opening failed: " + blockError.Error()})
		return
	}

	// Validate data
	if _blockUUID == "" {
		c.JSON(422, responses.ValidationErrorResponse{Success: false, Message: "Missing blockUUID"})
		return
	}

	blockUUID, err = uuid.Parse(_blockUUID)
	if err != nil {
		c.JSON(422, responses.ValidationErrorResponse{Success: false, Message: "Invalid blockUUID"})
		return
	}

	fileUUID, err = uuid.Parse(_fileUUID)
	if err != nil {
		c.JSON(422, responses.ValidationErrorResponse{Success: false, Message: "Invalid fileUUID"})
		return
	}

	err = models.Transport.MarkAsUsed(fileUUID)
	if err != nil {
		panic("unimplemented")
	}

	var file *models.RegularFile = models.Transport.GetEnqueuedFileUpload(fileUUID).(*models.RegularFile)
	if file == nil {
		c.JSON(422, responses.ValidationErrorResponse{Success: false, Message: "Block belongs to an unknown file"})
		return
	}

	file.Blocks[blockUUID].Status = models.BLOCK_STATUS_IN_PROGRESS

	// Prepare block for upload
	contents := make([]uint8, blockHeader.Size)
	readSize, err := block.Read(contents)

	if err != nil || readSize != int(blockHeader.Size) {
		c.JSON(500, responses.OperationFailureResponse{Success: false, Message: "Block loading failed: " + err.Error()})
		return
	}

	var blockMetadata *apicalls.BlockMetadata = new(apicalls.BlockMetadata)
	blockMetadata.Ctx = c
	blockMetadata.FileUUID = fileUUID
	blockMetadata.Content = &contents
	blockMetadata.UUID = blockUUID
	blockMetadata.Size = blockHeader.Size
	blockMetadata.Status = &file.Blocks[blockUUID].Status
	blockMetadata.CompleteCallback = func(UUID uuid.UUID, status *int) {
		*status = models.BLOCK_STATUS_TRANSFERRED
		models.Transport.MarkAsCompleted(UUID)
	}

	err = file.Blocks[blockUUID].Disk.Upload(blockMetadata)
	if err != nil {
		c.JSON(500, responses.OperationFailureResponse{Success: false, Message: "Block uploading failed: " + err.Error()})
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
		c.JSON(404, responses.SuccessResponse{Success: false, Message: "SFTP connection failed"})
		return
	}
	err = sftpDisk.Upload(&blockMetadata)
	if err != nil {
		c.JSON(404, responses.SuccessResponse{Success: false, Message: "SFTP upload failed"})
		return
	}
	*/

	c.JSON(200, responses.SuccessResponse{Success: true, Message: "File Upload Endpoint"})
}

func FileRename(c *gin.Context) {
	c.JSON(200, responses.SuccessResponse{Success: true, Message: "File Rename Endpoint"})
}

func FileRemove(c *gin.Context) {
	c.JSON(200, responses.SuccessResponse{Success: true, Message: "File Remove Endpoint"})
}

func FileGet(c *gin.Context) {
	c.JSON(200, responses.SuccessResponse{Success: true, Message: "File Get Endpoint"})
}

func FileDownload(c *gin.Context) {
	// Get data from request
	//fileUUIDString := c.Param("FileUUID")
	blockUUIDString := c.PostForm("blockUUID")

	// Validate data
	if blockUUIDString == "" {
		c.JSON(422, responses.ValidationErrorResponse{Success: false, Message: "Missing blockUUID"})
		return
	}

	blockUUID, uuidError := uuid.Parse(blockUUIDString)
	if uuidError != nil {
		c.JSON(422, responses.ValidationErrorResponse{Success: false, Message: "Invalid blockUUID"})
		return
	}

	// FTP demo
	var blockMetadata = apicalls.BlockMetadata{
		UUID: blockUUID}

	var ftpDisk = SFTPDisk.NewSFTPDisk()
	ftpDisk.CreateCredentials("...")
	err := ftpDisk.Connect(nil)
	if err != nil {
		c.JSON(500, responses.OperationFailureResponse{Success: false, Message: "FTP connection failed: " + err.Error()})
		return
	}

	err = ftpDisk.Download(&blockMetadata)
	if err != nil {
		c.JSON(500, responses.OperationFailureResponse{Success: false, Message: "Block download failed: " + err.Error()})
		return
	}

	c.JSON(200, responses.BlockDownloadResponse{Success: true, Message: "File Download Endpoint", Block: *blockMetadata.Content})
}

func FileShare(c *gin.Context) {
	c.JSON(200, responses.SuccessResponse{Success: true, Message: "File Share Endpoint"})
}

func FileShareRemove(c *gin.Context) {
	c.JSON(200, responses.SuccessResponse{Success: true, Message: "File Share Remove Endpoint"})
}

type FileRequestCompleteBody struct {
	direction bool // true - download, false - upload
}

func FileRequestComplete(c *gin.Context) {
	var body FileRequestCompleteBody = FileRequestCompleteBody{}
	var _fileUUID string
	var fileUUID uuid.UUID
	var userUUID uuid.UUID
	var err error
	var file models.File
	var rsp responses.FileRequestResponse = responses.FileRequestResponse{}
	var code int

	userData, _ := c.Get("UserData")
	userUUID = userData.(middleware.UserData).UserUUID

	_fileUUID = c.Param("FileUUID")
	fileUUID, err = uuid.Parse(_fileUUID)
	if err != nil {
		// TODO
		panic("Unimplemented")
	}

	err = c.Bind(&body)
	if err != nil {
		// TODO
		panic("Unimplemented")
	}

	if body.direction {
		// TODO
		panic("Unimplemented")
	} else {
		file = models.Transport.GetEnqueuedFileUpload(fileUUID)
		rsp.Type = file.GetType()
		rsp.UUID = file.GetUUID().String()
		rsp.Name = file.GetName()
		rsp.Size = file.GetSize()

		var _file *models.RegularFile = file.(*models.RegularFile)
		for _, _block := range _file.Blocks {
			if _block.Status == models.BLOCK_STATUS_TRANSFERRED {
				continue
			}

			var b responses.FileRequestBlockResponse = responses.FileRequestBlockResponse{
				UUID:  _block.UUID.String(),
				Order: _block.Order,
				Size:  _block.Size,
			}

			rsp.Blocks = append(rsp.Blocks, b)
		}

		if len(rsp.Blocks) == 0 {
			rsp.Success = true
			rsp.Message = "Successfully transferred the file"

			db.DB.DatabaseHandle.Create(&dbo.File{
				AbstractDatabaseObject: dbo.AbstractDatabaseObject{
					UUID: file.GetUUID(),
				},
				UserUUID:         userUUID,
				Type:             file.GetType(),
				Name:             file.GetName(),
				CreationDate:     time.Now(),
				ModificationDate: time.Now(),
			})

			for _, _block := range _file.Blocks {
				db.DB.DatabaseHandle.Create(&dbo.Block{
					AbstractDatabaseObject: dbo.AbstractDatabaseObject{
						UUID: _block.UUID,
					},
					UserUUID:   userUUID,
					VolumeUUID: _file.Volume.UUID,
					DiskUUID:   _block.Disk.GetUUID(),
				})
			}

			models.Transport.RemoveEnqueuedFileUpload(fileUUID)

			code = 200
		} else {
			rsp.Success = false
			rsp.Message = "Failed to transfer the file"

			code = 449
		}
	}

	c.JSON(code, rsp)
}

func GetFiles(c *gin.Context) {
	c.JSON(200, responses.SuccessResponse{Success: true, Message: "Get Files Endpoint"})
}
