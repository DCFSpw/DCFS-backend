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
	"dcfs/util/logger"
	"dcfs/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"strconv"
)

// CreateDirectory - handler for Create directory request
//
// Create directory (POST /files/manage) - creating a new directory
// in the file system.
//
// params:
//   - c *gin.Context: context of the request
//
// return type:
//   - API response with appropriate HTTP code
func CreateDirectory(c *gin.Context) {
	var requestBody requests.DirectoryCreateRequest
	var userUUID uuid.UUID
	var volumeUUID uuid.UUID
	var rootUUID uuid.UUID
	var volume *models.Volume
	var directory *dbo.File

	// Retrieve and validate data from request
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		logger.Logger.Error("api", " Wrong request body.")
		c.JSON(422, responses.NewValidationErrorResponse(err))
		return
	}

	// Retrieve volumeUUID from request
	volumeUUID, err := uuid.Parse(requestBody.VolumeUUID)
	if err != nil {
		logger.Logger.Error("api", "Wrong volume UUID.")
		c.JSON(422, responses.NewValidationErrorResponseSingle(constants.VAL_UUID_INVALID, "VolumeUUID", "Provided VolumeUUID is not a valid UUID"))
		return
	}

	// Retrieve rootUUID from request if provided
	if requestBody.RootUUID != "" {
		rootUUID, err = uuid.Parse(requestBody.RootUUID)
		if err != nil {
			logger.Logger.Error("api", "Wrong root uuid.")
			c.JSON(422, responses.NewValidationErrorResponseSingle(constants.VAL_UUID_INVALID, "RootUUID", "Provided RootUUID is not a valid UUID"))
			return
		}

		logger.Logger.Debug("api", "The root uuid is: ", requestBody.RootUUID, ".")
	} else {
		rootUUID = uuid.Nil
		logger.Logger.Debug("api", "The root uuid is empty.")
	}

	// Retrieve userUUID from context
	userUUID = c.MustGet("UserData").(middleware.UserData).UserUUID

	// Retrieve volume from transport
	volume = models.Transport.GetVolume(volumeUUID)
	if volume == nil {
		logger.Logger.Error("api", "The volume with the provided uuid: ", volumeUUID.String(), " is not found.")
		c.JSON(404, responses.NewNotFoundErrorResponse(constants.TRANSPORT_VOLUME_NOT_FOUND, "Volume not found"))
		return
	}

	// Verify that the user is owner of the volume
	if userUUID != volume.UserUUID {
		logger.Logger.Error("api", "The provided user: ", userUUID.String(), " is not an owner of the provided volume: ", volume.UUID.String(), ".")
		c.JSON(404, responses.NewNotFoundErrorResponse(constants.OWNER_MISMATCH, "Volume not found"))
		return
	}

	// Verify that the rootUUID exists in the volume, and it's a directory
	errCode := db.ValidateRootDirectory(rootUUID, volumeUUID)
	if errCode != constants.SUCCESS {
		logger.Logger.Error("api", "The root directory could not be found on the provided volume.")
		c.JSON(404, responses.NewNotFoundErrorResponse(errCode, "Root directory not found"))
		return
	}

	// Create a new directory
	directory = dbo.NewDirectoryFromRequest(&requestBody, userUUID, rootUUID)
	logger.Logger.Debug("api", "Created a new directory ", directory.Name, " (", directory.UUID.String(), ")", ".")

	// Save directory to database
	result := db.DB.DatabaseHandle.Create(&directory)
	if result.Error != nil {
		logger.Logger.Error("api", "Could not save the newly created directory in the db.")
		c.JSON(500, responses.NewOperationFailureResponse(constants.DATABASE_ERROR, "Database operation failed: "+result.Error.Error()))
		return
	}
	logger.Logger.Debug("api", "The new directory: ", directory.Name, " has been saved in the db.")

	logger.Logger.Debug("api", "CreateDirectory endpoint successful exit.")
	c.JSON(200, responses.NewEmptySuccessResponse())
}

// GetFile - handler for Get file details request
//
// Get file details (GET /files/manage/{fileUUID}) - retrieving metadata
// of the specified file.
//
// params:
//   - c *gin.Context: context of the request
//
// return type:
//   - API response with appropriate HTTP code
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
		logger.Logger.Error("api", "A file with the provided uuid: ", fileUUID, " does not exist.")
		c.JSON(404, responses.NewNotFoundErrorResponse(dbErr, "File not found"))
		return
	}

	// Verify that the user is owner of the volume
	if userUUID != file.UserUUID {
		logger.Logger.Error("api", "The provided user: ", userUUID.String(), " is not the owner of the file with the provided uuid: ", file.UUID.String(), ".")
		c.JSON(404, responses.NewNotFoundErrorResponse(constants.OWNER_MISMATCH, "File not found"))
		return
	}

	// Retrieve file full path
	path, dbErr = db.GenerateFileFullPath(file.RootUUID)
	if dbErr != constants.SUCCESS {
		logger.Logger.Error("api", "Could not generate the complete path for the file with the uuid: ", file.UUID.String(), ".")
		c.JSON(404, responses.NewNotFoundErrorResponse(dbErr, "File not found"))
		return
	}

	// Return volume data
	logger.Logger.Debug("api", "GetFile endpoint successful exit.")
	c.JSON(200, responses.NewFileDataWithPathSuccessResponse(file, path))
}

// GetFiles - handler for Get list of files request
//
// Get list of files (GET /files/manage) - retrieving list of files in
// the specified directory of the file system.
//
// params:
//   - c *gin.Context: context of the request
//
// return type:
//   - API response with appropriate HTTP code
func GetFiles(c *gin.Context) {
	var files []dbo.File
	var userUUID uuid.UUID
	var volumeUUID uuid.UUID
	var rootUUID uuid.UUID
	var err error

	// Retrieve volumeUUID from query
	volumeUUIDString := c.Query("volumeUUID")
	if volumeUUIDString == "" {
		logger.Logger.Error("api", "Missing the field: VolumeUUID from the request url.")
		c.JSON(422, responses.NewValidationErrorResponseSingle(constants.VAL_VALIDATOR_ERROR, "volumeUUID", "Field VolumeUUID is required."))
		return
	} else {
		volumeUUID, err = uuid.Parse(volumeUUIDString)
		if err != nil {
			logger.Logger.Error("api", "Wrong volume uuid.")
			c.JSON(422, responses.NewValidationErrorResponseSingle(constants.VAL_UUID_INVALID, "volumeUUID", "Provided VolumeUUID is not a valid UUID"))
			return
		}
	}

	// Retrieve rootUUID from query
	rootUUIDString := c.Query("rootUUID")
	if rootUUIDString != "" {
		rootUUID, err = uuid.Parse(rootUUIDString)
		if err != nil {
			logger.Logger.Error("api", "Wrong root uuid.")
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
		logger.Logger.Error("api", "Could not retrieve the specified files from the db.")
		c.JSON(500, responses.NewOperationFailureResponse(constants.DATABASE_ERROR, "Database operation failed: "+err.Error()))
		return
	}

	// Return list of volumes
	logger.Logger.Debug("api", "GetFiles endpoint successful exit.")
	c.JSON(200, responses.NewGetFilesSuccessResponse(files))
}

// InitFileUploadRequest - handler for Init file upload request
//
// Init file upload request (POST /files/upload - initiating the process of
// uploading a file and retrieving a list of blocks to partition the file
// for upload.
//
// params:
//   - c *gin.Context: context of the request
//
// return type:
//   - API response with appropriate HTTP code
func InitFileUploadRequest(c *gin.Context) {
	var requestBody requests.InitFileUploadRequest
	var userUUID uuid.UUID
	var volumeUUID uuid.UUID
	var rootUUID uuid.UUID

	var file models.RegularFile
	var volume *models.Volume

	// Retrieve and validate data from request
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		logger.Logger.Error("api", "Wrong request body.")
		c.JSON(422, responses.NewValidationErrorResponse(err))
		return
	}

	// Retrieve userUUID from context
	userUUID = c.MustGet("UserData").(middleware.UserData).UserUUID

	// Retrieve volumeUUID from request
	volumeUUID, err := uuid.Parse(requestBody.VolumeUUID)
	if err != nil {
		logger.Logger.Error("api", "Wrong volume uuid.")
		c.JSON(422, responses.NewValidationErrorResponseSingle(constants.VAL_UUID_INVALID, "VolumeUUID", "Provided VolumeUUID is not a valid UUID"))
		return
	}

	// Retrieve rootUUID from request if provided
	if requestBody.RootUUID != "" {
		rootUUID, err = uuid.Parse(requestBody.RootUUID)
		if err != nil {
			logger.Logger.Error("api", "Wrong root uuid.")
			c.JSON(422, responses.NewValidationErrorResponseSingle(constants.VAL_UUID_INVALID, "RootUUID", "Provided RootUUID is not a valid UUID"))
			return
		}
	} else {
		rootUUID = uuid.Nil
	}

	// Retrieve volume from transport
	volume = models.Transport.GetVolume(volumeUUID)
	if volume == nil {
		logger.Logger.Error("api", "Could not find a volume with the provided uuid: ", volumeUUID.String())
		c.JSON(404, responses.NewNotFoundErrorResponse(constants.TRANSPORT_VOLUME_NOT_FOUND, "Volume not found"))
		return
	}

	// Verify that the user is owner of the volume
	if userUUID != volume.UserUUID {
		logger.Logger.Error("api", "The provided user: ", userUUID.String(), " is not the owner of the volume: ", volume.UUID.String(), ".")
		c.JSON(404, responses.NewNotFoundErrorResponse(constants.OWNER_MISMATCH, "Volume not found"))
		return
	}

	// Verify that the rootUUID exists in the volume, and it's a directory
	errCode := db.ValidateRootDirectory(rootUUID, volumeUUID)
	if errCode != constants.SUCCESS {
		logger.Logger.Error("api", "The provided root directory: ", rootUUID.String(), " does not exist on the provided volume: ", volumeUUID.String(), ".")
		c.JSON(404, responses.NewNotFoundErrorResponse(errCode, "Root directory not found"))
		return
	}

	logger.Logger.Debug("api", "Got a request for a file named: ", requestBody.File.Name, "of size: ", strconv.FormatUint(uint64(requestBody.File.Size), 10), ".")

	// Enqueue file for upload
	file = volume.FileUploadRequest(&requestBody, userUUID, rootUUID)
	models.Transport.FileUploadQueue.EnqueueInstance(file.GetUUID(), &file)

	logger.Logger.Debug("api", "Prepared a request with ", strconv.FormatUint(uint64(len(file.Blocks)), 10), " blocks")

	logger.Logger.Debug("api", "InitFileUploadRequest endpoint successful exit.")
	c.JSON(200, responses.NewInitFileUploadRequestResponse(userUUID, &file))
}

// UploadBlock - handler for Upload block details request
//
// Upload block (POST /files/upload/{fileUUID}) - uploading a single block of
// a file (according to the partitioning scheme returned by
// the Init file upload request).
//
// params:
//   - c *gin.Context: context of the request
//
// return type:
//   - API response with appropriate HTTP code
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
		logger.Logger.Error("api", "Wrong block uuid.")
		c.JSON(422, responses.NewValidationErrorResponseSingle(constants.VAL_UUID_INVALID, "BlockUUID", "Provided BlockUUID is not a valid UUID"))
		return
	}

	fileUUID, err = uuid.Parse(_fileUUID)
	if err != nil {
		logger.Logger.Error("api", "Wrong file uuid.")
		c.JSON(422, responses.NewValidationErrorResponseSingle(constants.VAL_UUID_INVALID, "FileUUID", "Provided fileUUID is not a valid UUID"))
		return
	}

	// Retrieve block binary data from request
	blockHeader, err := c.FormFile("block")
	if err != nil {
		logger.Logger.Error("api", "Missing block binary data in the request.")
		c.JSON(422, responses.NewValidationErrorResponseSingle(constants.VAL_UUID_INVALID, "block", "Missing block"))
		return
	}

	// Open block data
	block, blockError := blockHeader.Open()
	if blockError != nil {
		logger.Logger.Error("api", "Could not open the block binary data for reading.")
		c.JSON(500, responses.NewOperationFailureResponse(constants.FS_CANNOT_OPEN_BLOCK, "Block opening failed: "+blockError.Error()))
		return
	}

	// Lock file for upload
	err = models.Transport.FileUploadQueue.MarkAsUsed(fileUUID)
	if err != nil {
		logger.Logger.Error("api", "Failed to lock the file: ", fileUUID.String())
		c.JSON(500, responses.NewOperationFailureResponse(constants.TRANSPORT_LOCK_FAILED, "Failed to lock file: "+err.Error()))
		return
	}

	// Retrieve file from transport and set block status to uploading
	file = models.Transport.FileUploadQueue.GetEnqueuedInstance(fileUUID).(*models.RegularFile)
	if file == nil {
		logger.Logger.Error("api", "The block: ", blockUUID.String(), " belongs to an unknown file.")
		c.JSON(500, responses.NewOperationFailureResponse(constants.FS_BLOCK_MISMATCH, "Block belongs to an unknown file"))
		return
	}
	file.Blocks[blockUUID].Status = constants.BLOCK_STATUS_IN_PROGRESS
	logger.Logger.Debug("api", "Changed the status of the block: ", blockUUID.String(), " to BLOCK_STATUS_IN_PROGRESS.")

	// Read block binary data
	contents := make([]uint8, blockHeader.Size)
	readSize, err := block.Read(contents)

	if err != nil || readSize != int(blockHeader.Size) {
		logger.Logger.Error("api", "Could not read the contents of the block: ", blockUUID.String(), ".")
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

		// Unblock the current file in the FileUploadQueue when this block is transferred
		models.Transport.FileUploadQueue.MarkAsCompleted(UUID)
	}

	// Block the current file in the FileUploadQueue
	err = models.Transport.FileUploadQueue.MarkAsUsed(fileUUID)
	if err != nil {
		logger.Logger.Error("api", "Failed to lock file: ", file.UUID.String())
		c.JSON(500, responses.NewOperationFailureResponse(constants.TRANSPORT_LOCK_FAILED, "Failed to lock file: "+err.Error()))
		return
	}

	// Calculate block checksum
	file.Blocks[blockUUID].Checksum = utils.CalculateChecksum(contents)

	// Upload file to target disk
	errorWrapper := file.Blocks[blockUUID].Disk.Upload(blockMetadata)
	if errorWrapper != nil {
		logger.Logger.Error("api", "Failed to upload the block: ", _blockUUID)
		c.JSON(500, responses.NewOperationFailureResponse(errorWrapper.Code, "Block loading failed: "+errorWrapper.Error.Error()))

		// Unblock the current file in the FileUploadQueue in case of failure
		models.Transport.FileUploadQueue.MarkAsCompleted(fileUUID)
		return
	}

	// Update target disk usage
	file.Blocks[blockUUID].Disk.UpdateUsedSpace(int64(file.Blocks[blockUUID].Size))

	logger.Logger.Debug("api", "UploadBlock endpoint successful exit.")
	c.JSON(200, responses.NewEmptySuccessResponse())
}

// UpdateFile - handler for Update file request
//
// Update file (PUT /files/manage/{fileUUID}) - updating the name or location
// of the specified file.
//
// params:
//   - c *gin.Context: context of the request
//
// return type:
//   - API response with appropriate HTTP code
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
		logger.Logger.Error("api", "Wrong file uuid.")
		c.JSON(422, responses.NewValidationErrorResponseSingle(constants.VAL_UUID_INVALID, "FileUUID", "Provided FileUUID is not a valid UUID"))
		return
	}

	// Retrieve and validate data from request
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		logger.Logger.Error("api", "Wrong request body.")
		c.JSON(422, responses.NewValidationErrorResponse(err))
		return
	}

	// Retrieve rootUUID from request if provided
	if requestBody.RootUUID != "" {
		rootUUID, err = uuid.Parse(requestBody.RootUUID)
		if err != nil {
			logger.Logger.Error("api", "Wrong root uuid.")
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
		logger.Logger.Error("api", "A file with the provided uuid: ", fileUUID, " was not found in the db.")
		c.JSON(404, responses.NewNotFoundErrorResponse(dbErr, "File not found"))
		return
	}

	// Verify that the user is owner of the file
	if userUUID != file.UserUUID {
		logger.Logger.Error("api", "The user: ", userUUID.String(), " is not the owner of the file: ", file.UUID.String(), ".")
		c.JSON(404, responses.NewNotFoundErrorResponse(constants.OWNER_MISMATCH, "Volume not found"))
		return
	}

	// Verify that the rootUUID exists in the volume, and it's a directory
	errCode := db.ValidateRootDirectory(rootUUID, file.VolumeUUID)
	if errCode != constants.SUCCESS {
		logger.Logger.Error("api", "The provided root directory: ", rootUUID.String(), " was not found on the volume: ", file.VolumeUUID.String(), ".")
		c.JSON(404, responses.NewNotFoundErrorResponse(errCode, "Root directory not found"))
		return
	}

	// Update file name and root directory
	file.Name = requestBody.Name
	file.RootUUID = rootUUID

	// Save changes to database
	result := db.DB.DatabaseHandle.Save(&file)
	if result.Error != nil {
		logger.Logger.Error("api", "Could not update the file: ", file.UUID.String(), " in the database.")
		c.JSON(500, responses.NewOperationFailureResponse(constants.DATABASE_ERROR, "Database operation failed: "+result.Error.Error()))
		return
	}

	logger.Logger.Debug("api", "UpdateFile endpoint successful exit.")
	c.JSON(200, responses.NewFileDataSuccessResponse(file))
}

// DeleteFile - handler for Delete file request
//
// Delete file (DELETE /files/manage/fileUUID) - deleting the specified file.
//
// params:
//   - c *gin.Context: context of the request
//
// return type:
//   - API response with appropriate HTTP code
func DeleteFile(c *gin.Context) {
	// TODO: Implement deletion of the file
	logger.Logger.Debug("api", "DeleteFile endpoint successful exit.")
	c.JSON(200, responses.NewEmptySuccessResponse())
}

// InitFileDownloadRequest - handler for Init file download request
//
// Init file download request (POST /files/download/{fileUUID}) - initiating
// the process of downloading a file and retrieving a list of blocks.
//
// params:
//   - c *gin.Context: context of the request
//
// return type:
//   - API response with appropriate HTTP code
func InitFileDownloadRequest(c *gin.Context) {
	var fileUUID uuid.UUID
	var files []uuid.UUID = make([]uuid.UUID, 0)
	var file *dbo.File
	var blocks []*dbo.Block
	var err error
	var code string
	var response *responses.SuccessResponse = nil

	// Retrieve and validate fileUUID from params
	fileUUID, err = uuid.Parse(c.Param("FileUUID"))
	if err != nil {
		logger.Logger.Error("api", "Wrong file uuid.")
		c.JSON(422, responses.NewValidationErrorResponseSingle(constants.VAL_UUID_INVALID, "FileUUID", "Provided FileUUID is not a valid UUID"))
		return
	}

	// Potentially backend could receive an array of fileUUIDs to download
	files = append(files, fileUUID)

	if len(files) == 1 {
		file, code = db.FileFromDatabase(fileUUID.String())
		if file == nil {
			logger.Logger.Error("api", "A file with the provided uuid: ", fileUUID.String(), " was not found in the db.")
			c.JSON(404, responses.NewNotFoundErrorResponse(code, "File not found"))
			return
		}

		if file.Size <= constants.FRONT_RAM_CAPACITY {
			blocks, code = db.BlocksFromDatabase(fileUUID.String())
			if blocks == nil {
				logger.Logger.Warning("api", "Could not find file blocks in the db.")
				c.JSON(405, responses.NewOperationFailureResponse(code, "File corrupted"))
				return
			}

			f := models.NewFileFromDBO(file)
			f = models.NewFileWrapper(constants.FILE_TYPE_SMALLER_WRAPPER, []models.File{f})
			models.Transport.FileDownloadQueue.EnqueueInstance(f.GetUUID(), f)
			logger.Logger.Debug("api", "Successfully enqueued the file: ", file.UUID.String(), " for download")

			response = responses.NewInitFileUploadRequestResponse(file.UserUUID, f)
		}
	}

	if response == nil {
		var _files []models.File
		for _, UUID := range files {
			_f, code := db.FileFromDatabase(UUID.String())
			if file == nil {
				logger.Logger.Error("api", "A File with the uuid: ", UUID.String(), " was not found in the db.")
				c.JSON(404, responses.NewNotFoundErrorResponse(code, "File not found"))
				return
			}

			f := models.NewFileFromDBO(_f)
			_files = append(_files, f)
		}

		wrapper := models.NewFileWrapper(constants.FILE_TYPE_WRAPPER, _files)
		models.Transport.FileDownloadQueue.EnqueueInstance(wrapper.GetUUID(), wrapper)
		logger.Logger.Debug("api", "Successfully enqueued the file: ", wrapper.GetUUID().String(), " for download")

		response = responses.NewInitFileUploadRequestResponse(file.UserUUID, wrapper)
	}

	logger.Logger.Debug("api", "InitDownloadRequest endpoint successful exit.")
	c.JSON(200, response)
}

// DownloadBlock - handler for Download block request
//
// Download block (POST /files/download/{fileUUID}) - downloading a single
// block of a file (according to the partitioning scheme returned by
// the Init file download request).
//
// params:
//   - c *gin.Context: context of the request
//
// return type:
//   - API response with appropriate HTTP code
func DownloadBlock(c *gin.Context) {
	// Retrieve and validate fileUUID from query
	fileUUID, err := uuid.Parse(c.Query("fileUUID"))
	if err != nil {
		logger.Logger.Error("api", "Wrong file uuid.")
		c.JSON(422, responses.NewValidationErrorResponseSingle(constants.VAL_UUID_INVALID, "FileUUID", "Provided FileUUID is not a valid UUID"))
		return
	}

	// Retrieve and validate blockUUID from param
	blockUUID, err := uuid.Parse(c.Param("BlockUUID"))
	if err != nil {
		logger.Logger.Error("api", "Wrong block uuid.")
		c.JSON(422, responses.NewValidationErrorResponseSingle(constants.VAL_UUID_INVALID, "BlockUUID", "Provided BlockUUID is not a valid UUID"))
		return
	}

	// Retrieve file from transport queue
	file := models.Transport.FileDownloadQueue.GetEnqueuedInstance(fileUUID).(models.File)
	if file == nil {
		logger.Logger.Error("A file with the uuid: ", fileUUID.String(), " is not enqueued for download.")
		c.JSON(422, responses.NewValidationErrorResponseSingle(constants.VAL_UUID_INVALID, "FileUUID", "A file with the given UUID is not enqueued for download"))
		return
	}

	// Prepare internal request data
	bm := apicalls.BlockMetadata{
		Ctx:              c,
		FileUUID:         fileUUID,
		UUID:             blockUUID,
		Size:             0,
		Status:           nil,
		Content:          nil,
		CompleteCallback: nil,
		Checksum:         file.GetBlocks()[blockUUID].Checksum,
	}

	// Block the current file in the FileUploadQueue
	err = models.Transport.FileDownloadQueue.MarkAsUsed(fileUUID)
	if err != nil {
		logger.Logger.Error("api", "Failed to lock the file: ", fileUUID.String(), " with an error: ", err.Error(), ".")
		c.JSON(500, responses.NewOperationFailureResponse(constants.TRANSPORT_LOCK_FAILED, "Failed to lock file: "+err.Error()))
		return
	}

	// Download block and return it via callback
	errorWrapper := file.Download(&bm)
	if errorWrapper != nil {
		logger.Logger.Error("api", "Failed to download the block with the code: ", errorWrapper.Code, ".")
		c.JSON(500, responses.NewOperationFailureResponse(constants.REMOTE_FAILED_JOB, errorWrapper.Code))
	}

	logger.Logger.Debug("api", "DownloadBlock endpoint successful exit.")
}

// CompleteFileUploadRequest - handler for Complete file upload request
//
// Complete file upload request (POST /files/upload/{fileUUID}) - notifying
// that all blocks should have been uploaded which results in an integrity
// check on the backend side. In case of failure, it will return the list
// of blocks that need to be reuploaded.
//
// params:
//   - c *gin.Context: context of the request
//
// return type:
//   - API response with appropriate HTTP code
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
		logger.Logger.Error("api", "Wrong file uuid.")
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
		logger.Logger.Debug("api", "Found: ", strconv.FormatInt(int64(len(failedBlocks)), 10), " failed blocks.")
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
		logger.Logger.Error("api", "Could not save the file: ", file.GetUUID().String(), " in the db.")
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
			Checksum:   _block.Checksum,
		})

		if result.Error != nil {
			logger.Logger.Error("api", "Could not save the block: ", _block.UUID.String(), " in the db.")
			c.JSON(500, responses.NewOperationFailureResponse(constants.DATABASE_ERROR, "Database operation failed: "+result.Error.Error()))
			return
		}
	}

	// Remove file from transport
	models.Transport.FileUploadQueue.RemoveEnqueuedInstance(fileUUID)

	logger.Logger.Debug("api", "CompleteFileUploadRequest endpoint exit.")
	c.JSON(200, responses.NewEmptySuccessResponse())
}
