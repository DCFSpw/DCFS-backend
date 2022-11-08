package controllers

import (
	"dcfs/apicalls"
	"dcfs/constants"
	"dcfs/db"
	"dcfs/db/dbo"
	"dcfs/middleware"
	"dcfs/models"
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
	volume = models.Transport.GetVolume(volumeUUID)
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

func GetFile(c *gin.Context) {
	var file *dbo.File
	var fileUUID string
	var userUUID uuid.UUID
	var path []dbo.PathEntry

	// Retrieve fileUUID from path parameters
	fileUUID = c.Param("FileUUID")

	// Retrieve userUUID from context
	userUUID = c.MustGet("UserData").(middleware.UserData).UserUUID

	// Retrieve file from database
	file, dbErr := db.FileFromDatabase(fileUUID)
	if dbErr != constants.SUCCESS {
		c.JSON(404, responses.NewNotFoundErrorResponse(dbErr, "File not found"))
		return
	}

	// Verify that the user is owner of the volume
	if userUUID != file.UserUUID {
		c.JSON(404, responses.NewNotFoundErrorResponse(constants.OWNER_MISMATCH, "File not found"))
		return
	}

	// Retrieve file full path
	path, dbErr = db.GenerateFileFullPath(file.RootUUID)
	if dbErr != constants.SUCCESS {
		c.JSON(404, responses.NewNotFoundErrorResponse(dbErr, "File not found"))
		return
	}

	// Return volume data
	c.JSON(200, responses.NewFileDataWithPathSuccessResponse(file, path))
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

	// Retrieve list of files of current user from the database
	err = db.DB.DatabaseHandle.Where("user_uuid = ? AND volume_uuid = ? AND root_uuid = ?", userUUID, volumeUUID, rootUUID).Find(&files).Error
	if err != nil {
		c.JSON(500, responses.NewOperationFailureResponse(constants.DATABASE_ERROR, "Database operation failed: "+err.Error()))
		return
	}

	// Return list of volumes
	c.JSON(200, responses.NewGetFilesSuccessResponse(files))
}

func InitFileUploadRequest(c *gin.Context) {
	var requestBody requests.InitFileUploadRequest
	var userUUID uuid.UUID
	var volumeUUID uuid.UUID
	var rootUUID uuid.UUID

	var file models.RegularFile
	var volume *models.Volume

	// Retrieve and validate data from request
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(422, responses.NewValidationErrorResponse(err))
		return
	}

	// Retrieve userUUID from context
	userUUID = c.MustGet("UserData").(middleware.UserData).UserUUID

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

	// Retrieve volume from transport
	volume = models.Transport.GetVolume(volumeUUID)
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

	log.Println("[Init file upload] Got a request for a file named: ", requestBody.File.Name, "of size: ", requestBody.File.Size)

	// Enqueue file for upload
	file = volume.FileUploadRequest(&requestBody, userUUID, rootUUID)
	models.Transport.FileUploadQueue.EnqueueInstance(file.GetUUID(), &file)

	log.Println("[Init file upload] Prepared a request with ", len(file.Blocks), " blocks")

	c.JSON(200, responses.NewInitFileUploadRequestResponse(userUUID, &file))
}

func UploadBlock(c *gin.Context) {
	var fileUUID uuid.UUID
	var blockUUID uuid.UUID
	var _fileUUID string
	var _blockUUID string
	var file *models.RegularFile

	// Retrieve and validate data from request
	_blockUUID = c.Param("BlockUUID")
	_fileUUID = c.PostForm("fileUUID")

	// Validate data
	blockUUID, err := uuid.Parse(_blockUUID)
	if err != nil {
		c.JSON(422, responses.NewValidationErrorResponseSingle(constants.VAL_UUID_INVALID, "BlockUUID", "Provided BlockUUID is not a valid UUID"))
		return
	}

	fileUUID, err = uuid.Parse(_fileUUID)
	if err != nil {
		c.JSON(422, responses.NewValidationErrorResponseSingle(constants.VAL_UUID_INVALID, "FileUUID", "Provided fileUUID is not a valid UUID"))
		return
	}

	// Retrieve block binary data from request
	blockHeader, err := c.FormFile("block")
	if err != nil {
		c.JSON(422, responses.NewValidationErrorResponseSingle(constants.VAL_UUID_INVALID, "block", "Missing block"))
		return
	}

	// Open block data
	block, blockError := blockHeader.Open()
	if blockError != nil {
		c.JSON(500, responses.NewOperationFailureResponse(constants.FS_CANNOT_OPEN_BLOCK, "Block opening failed: "+blockError.Error()))
		return
	}

	// Lock file for upload
	err = models.Transport.FileUploadQueue.MarkAsUsed(fileUUID)
	if err != nil {
		c.JSON(500, responses.NewOperationFailureResponse(constants.TRANSPORT_LOCK_FAILED, "Failed to lock file: "+err.Error()))
		return
	}

	// Retrieve file from transport and set block status to uploading
	file = models.Transport.FileUploadQueue.GetEnqueuedInstance(fileUUID).(*models.RegularFile)
	if file == nil {
		c.JSON(500, responses.NewOperationFailureResponse(constants.FS_BLOCK_MISMATCH, "Block belongs to an unknown file"))
		return
	}
	file.Blocks[blockUUID].Status = constants.BLOCK_STATUS_IN_PROGRESS

	// Read block binary data
	contents := make([]uint8, blockHeader.Size)
	readSize, err := block.Read(contents)

	if err != nil || readSize != int(blockHeader.Size) {
		c.JSON(500, responses.NewOperationFailureResponse(constants.FS_CANNOT_LOAD_BLOCK, "Block loading failed: "+err.Error()))
		return
	}

	// Save real size of the block
	file.Blocks[blockUUID].Size = readSize

	// Prepare internal block metadata
	var blockMetadata *apicalls.BlockMetadata = new(apicalls.BlockMetadata)
	blockMetadata.Ctx = c
	blockMetadata.FileUUID = fileUUID
	blockMetadata.Content = &contents
	blockMetadata.UUID = blockUUID
	blockMetadata.Size = blockHeader.Size
	blockMetadata.Status = &file.Blocks[blockUUID].Status
	blockMetadata.CompleteCallback = func(UUID uuid.UUID, status *int) {
		*status = constants.BLOCK_STATUS_TRANSFERRED

		// unblock the current file in the FileUploadQueue when this block is transferred
		models.Transport.FileUploadQueue.MarkAsCompleted(UUID)
	}

	// block the current file in the FileUploadQueue
	err = models.Transport.FileUploadQueue.MarkAsUsed(fileUUID)
	if err != nil {
		c.JSON(500, responses.NewOperationFailureResponse(constants.TRANSPORT_LOCK_FAILED, "Failed to lock file: "+err.Error()))
		return
	}

	// Upload file to target disk
	errorWrapper := file.Blocks[blockUUID].Disk.Upload(blockMetadata)
	if errorWrapper != nil {
		c.JSON(500, responses.NewOperationFailureResponse(errorWrapper.Code, "Block loading failed: "+errorWrapper.Error.Error()))

		// unblock the current file in the FileUploadQueue in case of failure
		models.Transport.FileUploadQueue.MarkAsCompleted(fileUUID)
		return
	}

	c.JSON(200, responses.NewEmptySuccessResponse())
}

func UpdateFile(c *gin.Context) {
	var requestBody requests.UpdateFileRequest
	var fileUUID string
	var userUUID uuid.UUID
	var rootUUID uuid.UUID
	var file *dbo.File

	// Retrieve and validate fileUUID
	fileUUID = c.Param("FileUUID")
	_, err := uuid.Parse(fileUUID)
	if err != nil {
		c.JSON(422, responses.NewValidationErrorResponseSingle(constants.VAL_UUID_INVALID, "FileUUID", "Provided FileUUID is not a valid UUID"))
		return
	}

	// Retrieve and validate data from request
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(422, responses.NewValidationErrorResponse(err))
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

	// Retrieve file from database
	file, dbErr := db.FileFromDatabase(fileUUID)
	if dbErr != constants.SUCCESS {
		c.JSON(404, responses.NewNotFoundErrorResponse(dbErr, "File not found"))
		return
	}

	// Verify that the user is owner of the file
	if userUUID != file.UserUUID {
		c.JSON(404, responses.NewNotFoundErrorResponse(constants.OWNER_MISMATCH, "Volume not found"))
		return
	}

	// Verify that the rootUUID exists in the volume, and it's a directory
	errCode := db.ValidateRootDirectory(rootUUID, file.VolumeUUID)
	if errCode != constants.SUCCESS {
		c.JSON(404, responses.NewNotFoundErrorResponse(errCode, "Root directory not found"))
		return
	}

	// Update file name and root directory
	file.Name = requestBody.Name
	file.RootUUID = rootUUID

	// Save changes to database
	result := db.DB.DatabaseHandle.Save(&file)
	if result.Error != nil {
		c.JSON(500, responses.NewOperationFailureResponse(constants.DATABASE_ERROR, "Database operation failed: "+result.Error.Error()))
		return
	}

	c.JSON(200, responses.NewFileDataSuccessResponse(file))
}

func FileRemove(c *gin.Context) {
	c.JSON(200, responses.SuccessResponse{Success: true, Message: "File Remove Endpoint"})
}

func InitFileDownloadRequest(c *gin.Context) {
	var fileUUID uuid.UUID
	var files []uuid.UUID = make([]uuid.UUID, 1)
	var file *dbo.File
	var blocks []*dbo.Block
	var err error
	var code string
	var response *responses.EmptySuccessResponse = nil

	fileUUID, err = uuid.Parse(c.Param("FileUUID"))
	if err != nil {
		c.JSON(422, responses.NewValidationErrorResponseSingle(constants.VAL_UUID_INVALID, "FileUUID", "Provided FileUUID is not a valid UUID"))
		return
	}

	// potentially backend could receive an array of fileUUIDs to download
	files = append(files, fileUUID)

	if len(files) == 1 {
		file, code = db.FileFromDatabase(fileUUID.String())
		if file == nil {
			c.JSON(404, responses.NewNotFoundErrorResponse(code, "File not found"))
			return
		}

		if file.Size <= constants.FRONT_RAM_CAPACITY {
			blocks, code = db.BlocksFromDatabase(fileUUID.String())
			if blocks == nil {
				c.JSON(405, responses.NewOperationFailureResponse(code, "File corrupted"))
				return
			}

			f := models.NewFileFromDBO(file)
			f = models.NewFileWrapper(constants.FILE_TYPE_SMALLER_WRAPPER, []models.File{f})
			models.Transport.FileDownloadQueue.EnqueueInstance(f.GetUUID(), f)

			response = responses.NewInitFileUploadRequestResponse(file.UserUUID, f)
		}
	}

	if response == nil {
		var _files []models.File
		for _, UUID := range files {
			_f, code := db.FileFromDatabase(UUID.String())
			if file == nil {
				c.JSON(404, responses.NewNotFoundErrorResponse(code, "File not found"))
				return
			}

			f := models.NewFileFromDBO(_f)
			_files = append(_files, f)
		}

		wrapper := models.NewFileWrapper(constants.FILE_TYPE_WRAPPER, _files)
		models.Transport.FileDownloadQueue.EnqueueInstance(wrapper.GetUUID(), wrapper)

		response = responses.NewInitFileUploadRequestResponse(file.UserUUID, wrapper)
	}

	c.JSON(200, response)
}

func DownloadBlock(c *gin.Context) {
	fileUUID, err := uuid.Parse(c.PostForm("fileUUID"))
	if err != nil {
		c.JSON(422, responses.NewValidationErrorResponseSingle(constants.VAL_UUID_INVALID, "FileUUID", "Provided FileUUID is not a valid UUID"))
		return
	}

	blockUUID, err := uuid.Parse(c.Param("BlockUUID"))
	if err != nil {
		c.JSON(422, responses.NewValidationErrorResponseSingle(constants.VAL_UUID_INVALID, "BlockUUID", "Provided BlockUUID is not a valid UUID"))
		return
	}

	file := models.Transport.FileDownloadQueue.GetEnqueuedInstance(fileUUID).(models.File)
	bm := apicalls.BlockMetadata{
		Ctx:              c,
		FileUUID:         fileUUID,
		UUID:             blockUUID,
		Size:             0,
		Status:           nil,
		Content:          nil,
		CompleteCallback: nil,
	}

	// block the current file in the FileUploadQueue
	err = models.Transport.FileDownloadQueue.MarkAsUsed(fileUUID)
	if err != nil {
		c.JSON(500, responses.NewOperationFailureResponse(constants.TRANSPORT_LOCK_FAILED, "Failed to lock file: "+err.Error()))
		return
	}

	errorWrapper := file.Download(&bm)
	if errorWrapper != nil {
		c.JSON(500, responses.NewOperationFailureResponse(constants.REMOTE_FAILED_JOB, errorWrapper.Code))
	}
}

func CompleteFileUploadRequest(c *gin.Context) {
	var fileUUID uuid.UUID
	var _fileUUID string
	var userUUID uuid.UUID
	var file models.RegularFile
	var failedBlocks []responses.FileRequestBlockResponse

	// Retrieve userUUID from context
	userUUID = c.MustGet("UserData").(middleware.UserData).UserUUID

	// Retrieve and validate data from request
	_fileUUID = c.Param("FileUUID")
	fileUUID, err := uuid.Parse(_fileUUID)
	if err != nil {
		c.JSON(422, responses.NewValidationErrorResponseSingle(constants.VAL_UUID_INVALID, "FileUUID", "Provided fileUUID is not a valid UUID"))
		return
	}

	// Retrieve file from transport
	file = *models.Transport.FileUploadQueue.GetEnqueuedInstance(fileUUID).(*models.RegularFile)

	// Verify whether blocks were successfully uploaded
	for _, _block := range file.Blocks {
		if _block.Status == constants.BLOCK_STATUS_TRANSFERRED {
			continue
		}

		// Add untransferred blocks to failed response
		var b = responses.FileRequestBlockResponse{
			UUID:  _block.UUID,
			Order: _block.Order,
			Size:  _block.Size,
		}

		failedBlocks = append(failedBlocks, b)
	}

	// If there are failed blocks, return them
	if len(failedBlocks) > 0 {
		c.JSON(449, responses.NewBlockTransferFailureResponse(failedBlocks))
		return
	}

	// Save file to database
	result := db.DB.DatabaseHandle.Create(&dbo.File{
		AbstractDatabaseObject: dbo.AbstractDatabaseObject{
			UUID: file.GetUUID(),
		},
		VolumeUUID: file.GetVolume().UUID,
		RootUUID:   file.GetRoot(),
		UserUUID:   userUUID,
		Type:       file.GetType(),
		Name:       file.GetName(),
		Size:       file.GetSize(),
		Checksum:   "",
	})

	if result.Error != nil {
		c.JSON(500, responses.NewOperationFailureResponse(constants.DATABASE_ERROR, "Database operation failed: "+result.Error.Error()))
		return
	}

	// Save blocks to database
	for _, _block := range file.Blocks {
		result := db.DB.DatabaseHandle.Create(&dbo.Block{
			AbstractDatabaseObject: dbo.AbstractDatabaseObject{
				UUID: _block.UUID,
			},
			FileUUID:   fileUUID,
			UserUUID:   userUUID,
			VolumeUUID: file.Volume.UUID,
			DiskUUID:   _block.Disk.GetUUID(),
			Size:       _block.Size,
			Order:      _block.Order,
		})

		if result.Error != nil {
			c.JSON(500, responses.NewOperationFailureResponse(constants.DATABASE_ERROR, "Database operation failed: "+result.Error.Error()))
			return
		}
	}

	// Remove file from transport
	models.Transport.FileUploadQueue.RemoveEnqueuedInstance(fileUUID)

	c.JSON(200, responses.NewEmptySuccessResponse())
}
