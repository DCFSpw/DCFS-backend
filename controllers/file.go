package controllers

import (
	"dcfs/apicalls"
	"dcfs/constants"
	"dcfs/db"
	"dcfs/db/dbo"
	"dcfs/middleware"
	"dcfs/models"
	"dcfs/models/disk/SFTPDisk"
	"dcfs/requests"
	"dcfs/responses"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"log"
)

func CreateDirectory(c *gin.Context) {
	var requestBody requests.DirectoryCreateRequest
	var userUUID uuid.UUID
	var volumeUUID uuid.UUID
	var rootUUID uuid.UUID
	var volume *models.Volume
	var directory *dbo.File

	// Retrieve and validate data from request
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(422, responses.NewValidationErrorResponse(err))
		return
	}

	// Retrieve volumeUUID from request
	volumeUUID, err := uuid.Parse(requestBody.VolumeUUID)
	if err != nil {
		c.JSON(422, responses.NewValidationErrorResponseSingle(constants.VAL_UUID_INVALID, "VolumeUUID", "Provided VolumeUUID is not a valid UUID"))
		return
	}

	// Retrieve rootUUID from request if provided
	if requestBody.RootUUID != "" {
		rootUUID, err = uuid.Parse(requestBody.RootUUID)
		if err != nil {
			c.JSON(422, responses.NewValidationErrorResponseSingle(constants.VAL_UUID_INVALID, "RootUUID", "Provided RootUUID is not a valid UUID"))
			return
		}
	} else {
		rootUUID = uuid.Nil
	}

	// Retrieve userUUID from context
	userUUID = c.MustGet("UserData").(middleware.UserData).UserUUID

	// Retrieve volume from transport
	volume = models.Transport.GetVolume(userUUID, volumeUUID)
	if volume == nil {
		c.JSON(404, responses.NewNotFoundErrorResponse(constants.TRANSPORT_VOLUME_NOT_FOUND, "Volume not found"))
		return
	}

	// Verify that the user is owner of the volume
	if userUUID != volume.UserUUID {
		c.JSON(404, responses.NewNotFoundErrorResponse(constants.OWNER_MISMATCH, "Volume not found"))
		return
	}

	// Verify that the rootUUID exists in the volume, and it's a directory
	errCode := db.ValidateRootDirectory(rootUUID, volumeUUID)
	if errCode != constants.SUCCESS {
		c.JSON(404, responses.NewNotFoundErrorResponse(errCode, "Root directory not found"))
		return
	}

	// Create a new directory
	directory = dbo.NewDirectoryFromRequest(&requestBody, userUUID, rootUUID)

	// Save directory to database
	result := db.DB.DatabaseHandle.Create(&directory)
	if result.Error != nil {
		c.JSON(500, responses.NewOperationFailureResponse(constants.DATABASE_ERROR, "Database operation failed: "+result.Error.Error()))
		return
	}

	c.JSON(200, responses.NewEmptySuccessResponse())
}

func GetFiles(c *gin.Context) {
	var files []dbo.File
	var userUUID uuid.UUID
	var volumeUUID uuid.UUID
	var rootUUID uuid.UUID
	var err error

	// Retrieve volumeUUID from query
	volumeUUIDString := c.Query("volumeUUID")
	if volumeUUIDString == "" {
		c.JSON(422, responses.NewValidationErrorResponseSingle(constants.VAL_VALIDATOR_ERROR, "volumeUUID", "Field VolumeUUID is required."))
		return
	} else {
		volumeUUID, err = uuid.Parse(volumeUUIDString)
		if err != nil {
			c.JSON(422, responses.NewValidationErrorResponseSingle(constants.VAL_UUID_INVALID, "volumeUUID", "Provided VolumeUUID is not a valid UUID"))
			return
		}
	}

	// Retrieve rootUUID from query
	rootUUIDString := c.Query("rootUUID")
	if rootUUIDString != "" {
		rootUUID, err = uuid.Parse(rootUUIDString)
		if err != nil {
			c.JSON(422, responses.NewValidationErrorResponseSingle(constants.VAL_UUID_INVALID, "RootUUID", "Provided RootUUID is not a valid UUID"))
			return
		}
	} else {
		rootUUID = uuid.Nil
	}

	// Retrieve userUUID from context
	userUUID = c.MustGet("UserData").(middleware.UserData).UserUUID

	// Retrieve list of volumes of current user from the database
	err = db.DB.DatabaseHandle.Where("user_uuid = ? AND volume_uuid = ? AND root_uuid = ?", userUUID, volumeUUID, rootUUID).Find(&files).Error
	if err != nil {
		c.JSON(500, responses.NewOperationFailureResponse(constants.DATABASE_ERROR, "Database operation failed: "+err.Error()))
		return
	}

	// Return list of volumes
	c.JSON(200, responses.NewGetFilesSuccessResponse(files))
}

func FileRequest(c *gin.Context) {
	var requestBody requests.GetFileRequest
	var fileUploadRequest apicalls.FileUploadRequest = apicalls.FileUploadRequest{}
	var file models.File
	var volume *models.Volume
	var err error

	userData, _ := c.Get("UserData")
	userUUID := userData.(middleware.UserData).UserUUID

	// Retrieve and validate data from request
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(422, responses.NewValidationErrorResponse(err))
		return
	}

	volumeUUID, err := uuid.Parse(requestBody.VolumeUUID)
	if err != nil {
		c.JSON(422, responses.NewValidationErrorResponseSingle(constants.VAL_UUID_INVALID, "VolumeUUID", "Provided VolumeUUID is not a valid UUID"))
		return
	}

	volume = models.Transport.GetVolume(userUUID, volumeUUID)
	if volume == nil {
		c.JSON(404, responses.NewNotFoundErrorResponse(constants.TRANSPORT_VOLUME_NOT_FOUND, "Cannot find volume with provided UUID"))
		return
	}

	log.Println("Got a request for a file named: ", requestBody.File.Name, "of size: ", requestBody.File.Size)

	fileUploadRequest.Size = requestBody.File.Size
	fileUploadRequest.Type = requestBody.File.Type
	fileUploadRequest.Name = requestBody.File.Name
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

	if file.GetType() != constants.FILE_TYPE_DIRECTORY {
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
		c.JSON(422, responses.NewValidationErrorResponseSingle(constants.VAL_UUID_INVALID, "block", "Missing block"))
		return
	}

	block, blockError := blockHeader.Open()
	if blockError != nil {
		c.JSON(500, responses.NewOperationFailureResponse(constants.FS_CANNOT_OPEN_BLOCK, "Block opening failed: "+blockError.Error()))
		return
	}

	// Validate data
	if _blockUUID == "" {
		c.JSON(422, responses.NewValidationErrorResponseSingle(constants.VAL_VALIDATOR_ERROR, "blockUUID", "Missing blockUUID"))
		return
	}

	blockUUID, err = uuid.Parse(_blockUUID)
	if err != nil {
		c.JSON(422, responses.NewValidationErrorResponseSingle(constants.VAL_UUID_INVALID, "blockUUID", "Provided blockUUID is not a valid UUID"))
		return
	}

	fileUUID, err = uuid.Parse(_fileUUID)
	if err != nil {
		c.JSON(422, responses.NewValidationErrorResponseSingle(constants.VAL_UUID_INVALID, "fileUUID", "Provided fileUUID is not a valid UUID"))
		return
	}

	err = models.Transport.MarkAsUsed(fileUUID)
	if err != nil {
		c.JSON(500, responses.NewOperationFailureResponse(constants.TRANSPORT_LOCK_FAILED, "Failed to lock file: "+err.Error()))
		return
	}

	var file *models.RegularFile = models.Transport.GetEnqueuedFileUpload(fileUUID).(*models.RegularFile)
	if file == nil {
		c.JSON(500, responses.NewOperationFailureResponse(constants.FS_BLOCK_MISMATCH, "Block belongs to an unknown file"))
		return
	}

	file.Blocks[blockUUID].Status = constants.BLOCK_STATUS_IN_PROGRESS

	// Prepare block for upload
	contents := make([]uint8, blockHeader.Size)
	readSize, err := block.Read(contents)

	if err != nil || readSize != int(blockHeader.Size) {
		c.JSON(500, responses.NewOperationFailureResponse(constants.FS_CANNOT_LOAD_BLOCK, "Block loading failed: "+err.Error()))
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
		*status = constants.BLOCK_STATUS_TRANSFERRED
		models.Transport.MarkAsCompleted(UUID)
	}

	errorWrapper := file.Blocks[blockUUID].Disk.Upload(blockMetadata)
	if errorWrapper != nil {
		c.JSON(500, responses.NewOperationFailureResponse(errorWrapper.Code, "Block loading failed: "+errorWrapper.Error.Error()))
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

	c.JSON(200, responses.NewEmptySuccessResponse())
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
	/*err := ftpDisk.Connect(nil)
	if err != nil {
		c.JSON(500, responses.OperationFailureResponse{Success: false, Message: "FTP connection failed: " + err.Error()})
		return
	}*/

	errorWrapper := ftpDisk.Download(&blockMetadata)
	if errorWrapper != nil {
		c.JSON(500, responses.OperationFailureResponse{Success: false, Message: "Block download failed: " + errorWrapper.Error.Error()})
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

func FileRequestComplete(c *gin.Context) {
	var requestBody requests.FileRequestCompleteRequest = requests.FileRequestCompleteRequest{}
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
		c.JSON(422, responses.NewValidationErrorResponseSingle(constants.VAL_UUID_INVALID, "fileUUID", "Provided fileUUID is not a valid UUID"))
		return
	}

	// Retrieve and validate data from request
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(422, responses.NewValidationErrorResponse(err))
		return
	}

	if requestBody.Direction {
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
			if _block.Status == constants.BLOCK_STATUS_TRANSFERRED {
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
				UserUUID: userUUID,
				Type:     file.GetType(),
				Name:     file.GetName(),
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
