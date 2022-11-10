package models

import (
	"dcfs/apicalls"
	"dcfs/constants"
	"dcfs/db"
	"dcfs/db/dbo"
	"dcfs/requests"
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"log"
	"math"
	"net/http"
	"os"
	"path"
	"sync"
	"time"
)

type File interface {
	GetUUID() uuid.UUID
	SetUUID(uuid.UUID)

	GetSize() int
	SetSize(newSize int)

	GetName() string
	SetName(newName string)

	GetType() int
	SetType(newType int)

	GetRoot() uuid.UUID
	SetRoot(rootUUID uuid.UUID)

	GetVolume() *Volume
	SetVolume(v *Volume)

	IsCompleted() bool
	GetBlocks() map[uuid.UUID]*Block

	GetFileDBO(userUUID uuid.UUID) dbo.File

	Download(blockMetadata *apicalls.BlockMetadata) *apicalls.ErrorWrapper // delegate the download process to avoid if statements

	Remove()
}

type AbstractFile struct {
	UUID uuid.UUID
	Name string
	Type int
	Size int

	RootUUID uuid.UUID
	Parent   File
	Volume   *Volume
}

func (file *AbstractFile) Remove() {
	panic("unimplemented")
}

func (file *AbstractFile) GetUUID() uuid.UUID {
	return file.UUID
}

func (file *AbstractFile) SetUUID(UUID uuid.UUID) {
	file.UUID = UUID
}

func (file *AbstractFile) GetSize() int {
	panic("Unimplemented abstract method")
}

func (file *AbstractFile) SetSize(newSize int) {
	file.Size = newSize
}

func (file *AbstractFile) GetName() string {
	return file.Name
}

func (file *AbstractFile) SetName(newName string) {
	file.Name = newName
}

func (file *AbstractFile) GetType() int {
	return file.Type
}

func (file *AbstractFile) SetType(newType int) {
	file.Type = newType
}

func (file *AbstractFile) GetRoot() uuid.UUID {
	return file.RootUUID
}

func (file *AbstractFile) SetRoot(rootUUID uuid.UUID) {
	file.RootUUID = rootUUID
}

func (file *AbstractFile) IsCompleted() bool {
	panic("Unimplemented abstract method")
}

func (file *AbstractFile) GetVolume() *Volume {
	return file.Volume
}

func (file *AbstractFile) SetVolume(v *Volume) {
	file.Volume = v
}

func (file *AbstractFile) GetBlocks() map[uuid.UUID]*Block {
	return nil
}

func (file *AbstractFile) GetFileDBO(userUUID uuid.UUID) dbo.File {
	var f = dbo.NewFile()

	f.UUID = file.UUID
	f.Name = file.Name
	f.Type = file.Type
	f.Size = file.Size
	f.RootUUID = file.RootUUID
	f.UserUUID = userUUID
	f.CreatedAt = time.Now()
	f.UpdatedAt = time.Now()
	f.Checksum = ""

	return *f
}

func (file *AbstractFile) Download(blockMetadata *apicalls.BlockMetadata) *apicalls.ErrorWrapper {
	panic("unimplemented abstract method")
}

type RegularFile struct {
	AbstractFile
	Blocks map[uuid.UUID]*Block
}

func (file *RegularFile) Remove() {
	panic("Unimplemented")
}

func (file *RegularFile) GetUUID() uuid.UUID {
	return file.AbstractFile.GetUUID()
}

func (file *RegularFile) SetUUID(UUID uuid.UUID) {
	file.AbstractFile.SetUUID(UUID)
}

func (file *RegularFile) GetSize() int {
	return file.Size
}

func (file *RegularFile) SetSize(newSize int) {
	file.AbstractFile.SetSize(newSize)
}

func (file *RegularFile) GetName() string {
	return file.AbstractFile.GetName()
}

func (file *RegularFile) SetName(newName string) {
	file.AbstractFile.SetName(newName)
}

func (file *RegularFile) GetType() int {
	return constants.FILE_TYPE_REGULAR
}

func (file *RegularFile) SetType(newType int) {
	file.AbstractFile.SetType(newType)
}

func (file *RegularFile) GetRoot() uuid.UUID {
	return file.AbstractFile.GetRoot()
}

func (file *RegularFile) SetRoot(rootUUID uuid.UUID) {
	file.AbstractFile.SetRoot(rootUUID)
}

func (file *RegularFile) GetFileDBO(userUUID uuid.UUID) dbo.File {
	return file.AbstractFile.GetFileDBO(userUUID)
}

func (file *RegularFile) IsCompleted() bool {
	for _, _block := range file.Blocks {
		if _block.Status != constants.BLOCK_STATUS_TRANSFERRED {
			return false
		}
	}

	return true
}

func (file *RegularFile) GetVolume() *Volume {
	return file.AbstractFile.GetVolume()
}

func (file *RegularFile) SetVolume(v *Volume) {
	file.AbstractFile.SetVolume(v)
}

func (file *RegularFile) GetBlocks() map[uuid.UUID]*Block {
	return file.Blocks
}

func (file *RegularFile) Download(blockMetadata *apicalls.BlockMetadata) *apicalls.ErrorWrapper {
	return file.AbstractFile.Download(blockMetadata)
}

type Directory struct {
	AbstractFile
	Files []File
}

func (d *Directory) Remove() {
	panic("Unimplemented")
}

func (d *Directory) GetUUID() uuid.UUID {
	return d.AbstractFile.GetUUID()
}

func (d *Directory) SetUUID(UUID uuid.UUID) {
	d.AbstractFile.SetUUID(UUID)
}

func (d *Directory) GetSize() int {
	var cumulativeSize int = 0
	for _, file := range d.Files {
		cumulativeSize += file.GetSize()
	}

	return cumulativeSize
}

func (d *Directory) SetSize(newSize int) {
	d.AbstractFile.SetSize(newSize)
}

func (d *Directory) GetName() string {
	return d.AbstractFile.GetName()
}

func (d *Directory) SetName(newName string) {
	d.AbstractFile.SetName(newName)
}

func (d *Directory) GetType() int {
	return d.AbstractFile.GetType()
}

func (d *Directory) SetType(newType int) {
	d.AbstractFile.SetType(newType)
}

func (d *Directory) GetRoot() uuid.UUID {
	return d.AbstractFile.GetRoot()
}

func (d *Directory) SetRoot(rootUUID uuid.UUID) {
	d.AbstractFile.SetRoot(rootUUID)
}

func (d *Directory) GetFileDBO(userUUID uuid.UUID) dbo.File {
	return d.AbstractFile.GetFileDBO(userUUID)
}

func (d *Directory) IsCompleted() bool {
	return true
}

func (d *Directory) GetVolume() *Volume {
	return d.AbstractFile.GetVolume()
}

func (d *Directory) SetVolume(v *Volume) {
	d.AbstractFile.SetVolume(v)
}

func (d *Directory) GetBlocks() map[uuid.UUID]*Block {
	return d.AbstractFile.GetBlocks()
}

func (d *Directory) Download(blockMetadata *apicalls.BlockMetadata) *apicalls.ErrorWrapper {
	return d.AbstractFile.Download(blockMetadata)
}

type SmallerFileWrapper struct {
	ActualFile File
}

func (f *SmallerFileWrapper) Remove() {
	panic("Unimplemented")
}

func (f *SmallerFileWrapper) GetUUID() uuid.UUID {
	return f.ActualFile.GetUUID()
}

func (f *SmallerFileWrapper) SetUUID(UUID uuid.UUID) {
	panic("unimplemented")
}

func (f *SmallerFileWrapper) GetSize() int {
	return f.ActualFile.GetSize()
}

func (f *SmallerFileWrapper) SetSize(newSize int) {
	panic("unimplemented")
}

func (f *SmallerFileWrapper) GetName() string {
	return f.ActualFile.GetName()
}

func (f *SmallerFileWrapper) SetName(newName string) {
	panic("unimplemented")
}

func (f *SmallerFileWrapper) GetType() int {
	return f.ActualFile.GetType()
}

func (f *SmallerFileWrapper) SetType(newType int) {
	panic("unimplemented")
}

func (f *SmallerFileWrapper) GetRoot() uuid.UUID {
	return f.ActualFile.GetRoot()
}

func (f *SmallerFileWrapper) SetRoot(rootUUID uuid.UUID) {
	panic("unimplemented")
}

func (f *SmallerFileWrapper) GetFileDBO(userUUID uuid.UUID) dbo.File {
	return f.ActualFile.GetFileDBO(userUUID)
}

func (f *SmallerFileWrapper) IsCompleted() bool {
	return f.ActualFile.IsCompleted()
}

func (f *SmallerFileWrapper) GetVolume() *Volume {
	return f.ActualFile.GetVolume()
}

func (f *SmallerFileWrapper) SetVolume(v *Volume) {
	panic("unimplemented")
}

func (f *SmallerFileWrapper) GetBlocks() map[uuid.UUID]*Block {
	return f.ActualFile.GetBlocks()
}

func (f *SmallerFileWrapper) Download(blockMetadata *apicalls.BlockMetadata) *apicalls.ErrorWrapper {
	block := f.GetBlocks()[blockMetadata.UUID]

	blockMetadata.Size = int64(block.Size)
	blockMetadata.Status = &block.Status
	blockMetadata.CompleteCallback = func(UUID uuid.UUID, status *int) {
		*status = constants.BLOCK_STATUS_TRANSFERRED

		// unblock the current file in the FileUploadQueue when this block is transferred
		Transport.FileUploadQueue.MarkAsCompleted(UUID)
	}
	blockMetadata.Content = new([]uint8)

	block.Status = constants.BLOCK_STATUS_QUEUED
	rsp := block.Disk.Download(blockMetadata)

	// the file should be deleted from the download queue after 6 minutes, or after the last block gets transferred
	if rsp != nil {
		// release the blocked file if download failed
		Transport.FileDownloadQueue.MarkAsCompleted(blockMetadata.UUID)
	}

	blockMetadata.Ctx.Data(http.StatusOK, "application/octet-stream", *blockMetadata.Content)
	return rsp
}

type FileWrapper struct {
	Files []File
	UUID  uuid.UUID
}

func (f *FileWrapper) Remove() {
	panic("Unimplemented")
}

func (f *FileWrapper) GetUUID() uuid.UUID {
	return f.UUID
}

func (f *FileWrapper) SetUUID(UUID uuid.UUID) {
	f.UUID = UUID
}

func (f *FileWrapper) GetSize() int {
	size := 0

	for _, f := range f.Files {
		size += f.GetSize()
	}

	return size
}

func (f *FileWrapper) SetSize(newSize int) {
	panic("unimplemented")
}

func (f *FileWrapper) GetName() string {
	if len(f.Files) == 1 {
		return f.Files[0].GetName()
	}

	return "files.zip"
}

func (f *FileWrapper) SetName(newName string) {
	panic("unimplemented")
}

func (f *FileWrapper) GetType() int {
	return constants.FILE_TYPE_WRAPPER
}

func (f *FileWrapper) SetType(newType int) {
	panic("unimplemented")
}

func (f *FileWrapper) GetRoot() uuid.UUID {
	// all files are in the same root
	return f.Files[0].GetRoot()
}

func (f *FileWrapper) SetRoot(rootUUID uuid.UUID) {
	panic("unimplemented")
}

func (f *FileWrapper) GetFileDBO(userUUID uuid.UUID) dbo.File {
	return dbo.File{
		AbstractDatabaseObject: dbo.AbstractDatabaseObject{
			UUID: f.UUID,
		},
		VolumeUUID: uuid.UUID{},
		RootUUID:   uuid.UUID{},
		UserUUID:   uuid.UUID{},
		Type:       constants.FILE_TYPE_WRAPPER,
		Name:       f.GetName(),
		Size:       f.GetSize(),
		Checksum:   "",
		CreatedAt:  time.Time{},
		UpdatedAt:  time.Time{},
		DeletedAt:  gorm.DeletedAt{},
		Volume:     dbo.Volume{},
		User:       dbo.User{},
	}
}

func (f *FileWrapper) IsCompleted() bool {
	panic("uimplemented")
}

func (f *FileWrapper) GetVolume() *Volume {
	// all files are on the same volume
	return f.Files[0].GetVolume()
}

func (f *FileWrapper) SetVolume(v *Volume) {
	panic("unimplemented")
}

func (f *FileWrapper) GetBlocks() map[uuid.UUID]*Block {
	ret := make(map[uuid.UUID]*Block)
	ret[f.UUID] = &Block{
		UUID:     f.UUID,
		UserUUID: uuid.UUID{},
		File:     f,
		Disk:     nil,
		Size:     f.GetSize(),
		Checksum: 0,
		Status:   0,
		Order:    0,
	}

	return ret
}

func (f *FileWrapper) downloadFile(_path string, file File, blockMetadata *apicalls.BlockMetadata) *apicalls.ErrorWrapper {
	downloadpath := path.Join(_path, file.GetName())
	fail := false
	blockSize := file.GetVolume().BlockSize

	_file, err := os.Create(downloadpath)
	if err != nil {
		return &apicalls.ErrorWrapper{
			Error: err,
			Code:  constants.REAL_FS_CREATE_FILE_ERROR,
		}
	}

	err = _file.Close()
	if err != nil {
		return &apicalls.ErrorWrapper{
			Error: err,
			Code:  constants.REAL_FS_CLOSE_FILE_ERROR,
		}
	}

	var wg sync.WaitGroup
	for _, block := range file.GetBlocks() {
		wg.Add(1)

		go func(_f *os.File, _b *Block) {
			defer wg.Done()

			bm := &apicalls.BlockMetadata{
				Ctx:      blockMetadata.Ctx,
				FileUUID: blockMetadata.FileUUID,
				UUID:     _b.UUID,
				Size:     int64(_b.Size),
				Status:   &_b.Status,
				Content:  new([]uint8),
				CompleteCallback: func(UUID uuid.UUID, status *int) {
					*status = constants.BLOCK_STATUS_TRANSFERRED
				},
			}

			errWrapper := _b.Disk.Download(bm)
			if errWrapper != nil {
				// one retry
				errWrapper = _b.Disk.Download(bm)
				if errWrapper != nil {
					fail = true
					return
				}
			}

			dest, err := os.OpenFile(downloadpath, os.O_RDWR, 777)
			if err != nil {
				fail = true
				return
			}

			defer func() {
				err := dest.Close()
				if err != nil {
					fail = true
				}
			}()

			// we may assume, every block is equal size, only the last one may be bigger / smaller
			_, err = dest.Seek(int64(_b.Order*blockSize), 0)
			if err != nil {
				fail = true
				return
			}

			var n int
			n, err = dest.Write(*bm.Content)

			if err != nil {
				fail = true
				return
			}

			if n != _b.Size {
				fail = true
				return
			}
		}(_file, block)
	}

	wg.Wait()

	if fail {
		return &apicalls.ErrorWrapper{
			Error: fmt.Errorf("Block downloading failed"),
			Code:  constants.REMOTE_FAILED_JOB,
		}
	}

	return nil
}

func (f *FileWrapper) downloadDirectory(_path string, dir *Directory, blockMetadata *apicalls.BlockMetadata) *apicalls.ErrorWrapper {
	downloadPath := path.Join(_path, dir.GetName())
	err := os.MkdirAll(downloadPath, 0777)
	if err != nil {
		return &apicalls.ErrorWrapper{
			Error: err,
			Code:  constants.REAL_FS_CREATE_DIR_ERROR,
		}
	}

	for _, file := range dir.Files {
		errWrapper := f.downloadFile(downloadPath, file, blockMetadata)
		if errWrapper != nil {
			return errWrapper
		}
	}

	return nil
}

func (f *FileWrapper) Download(blockMetadata *apicalls.BlockMetadata) *apicalls.ErrorWrapper {
	downloadPath := path.Join("./Download", f.GetName())
	err := os.MkdirAll(downloadPath, 0777)
	if err != nil {
		return &apicalls.ErrorWrapper{
			Error: err,
			Code:  constants.REAL_FS_CREATE_DIR_ERROR,
		}
	}

	for _, file := range f.Files {
		if file.GetType() == constants.FILE_TYPE_DIRECTORY {
			errWrapper := f.downloadDirectory(downloadPath, file.(*Directory), blockMetadata)
			if errWrapper != nil {
				return errWrapper
			}
		} else if file.GetType() == constants.FILE_TYPE_REGULAR {
			errWrapper := f.downloadFile(downloadPath, file, blockMetadata)
			if errWrapper != nil {
				return errWrapper
			}
		} else {
			return &apicalls.ErrorWrapper{
				Error: nil,
				Code:  constants.FS_BAD_FILE,
			}
		}
	}

	filename := f.Files[0].GetName()
	if len(f.Files) > 1 || len(f.Files) == 1 && f.Files[0].GetType() == constants.FILE_TYPE_DIRECTORY {
		// zip files and send the zip
		filename = "files.zip"
	}

	blockMetadata.Ctx.File(path.Join(downloadPath, filename))

	// the files should reside on the server for 1hr x 1GiB
	t := int(math.Max(1, math.Ceil(float64(f.GetSize())/float64(1024*1024*1024))))
	time.AfterFunc(time.Duration(t)*60*time.Minute, func() {
		err := os.RemoveAll(downloadPath)
		if err != nil {
			log.Printf("Could not remove file: %s", filename)
		}
	})

	return nil
}

func NewFileWrapper(filetype int, actualFiles []File) File {
	var f File

	if filetype == constants.FILE_TYPE_SMALLER_WRAPPER {
		f = new(SmallerFileWrapper)
		f.(*SmallerFileWrapper).ActualFile = actualFiles[0]
	} else if filetype == constants.FILE_TYPE_WRAPPER {
		f = new(FileWrapper)
		f.(*FileWrapper).Files = actualFiles
		f.SetUUID(uuid.New())
	} else {
		f = nil
	}

	return f
}

func NewFile(filetype int) File {
	var f File
	if filetype == constants.FILE_TYPE_REGULAR {
		f = new(RegularFile)
	} else {
		f = new(Directory)
	}

	f.SetType(filetype)
	f.SetUUID(uuid.New())

	return f
}

func NewFileFromDBO(fileDBO *dbo.File) File {
	if fileDBO.Type == constants.FILE_TYPE_DIRECTORY {
		return NewDirectoryFromDBO(fileDBO)
	}
	var file File

	if fileDBO.Type == constants.FILE_TYPE_REGULAR {
		_blocks, _ := db.BlocksFromDatabase(fileDBO.UUID.String())
		if _blocks == nil {
			return nil
		}
		var blocks map[uuid.UUID]*Block = make(map[uuid.UUID]*Block)

		for _, _b := range _blocks {
			d := CreateDiskFromUUID(_b.DiskUUID)

			b := NewBlockFromDBO(_b)
			b.File = file
			b.Disk = d

			blocks[b.UUID] = b
		}

		file = &RegularFile{
			AbstractFile: AbstractFile{
				UUID:     fileDBO.UUID,
				Name:     fileDBO.Name,
				Type:     fileDBO.Type,
				Size:     fileDBO.Size,
				RootUUID: fileDBO.RootUUID,
				Parent:   nil, // don't want to walk all the way up to '/'
				Volume:   Transport.GetVolume(fileDBO.VolumeUUID),
			},
			Blocks: blocks,
		}
	} else {
		return nil
	}

	return file
}

func NewDirectoryFromDBO(directoryDBO *dbo.File) File {
	if directoryDBO.Type != constants.FILE_TYPE_DIRECTORY {
		return nil
	}

	var _files []dbo.File
	var files []File

	err := db.DB.DatabaseHandle.Where("parent_uuid = ?", directoryDBO.UUID.String()).Find(&_files).Error
	if err != nil {
		return nil
	}

	for _, _f := range _files {
		f := NewFileFromDBO(&_f)
		files = append(files, f)
	}

	return &Directory{
		AbstractFile: AbstractFile{
			UUID:     directoryDBO.UUID,
			Name:     directoryDBO.Name,
			Type:     directoryDBO.Type,
			Size:     directoryDBO.Size,
			RootUUID: directoryDBO.RootUUID,
			Parent:   nil, // don't want to walk all the way up to '/'
			Volume:   Transport.GetVolume(directoryDBO.VolumeUUID),
		},
		Files: files,
	}
}

func NewFileFromRequest(request *requests.InitFileUploadRequest, rootUUID uuid.UUID) File {
	var f File = NewFile(request.File.Type)
	f.SetName(request.File.Name)
	f.SetSize(request.File.Size)
	f.SetRoot(rootUUID)

	return f
}
