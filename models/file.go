package models

import (
	"dcfs/apicalls"
	"dcfs/constants"
	"dcfs/db"
	"dcfs/db/dbo"
	"dcfs/requests"
	"github.com/google/uuid"
	"net/http"
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

func NewFileWrapper(filetype int, actualFile File) File {
	var f File

	if filetype == constants.FILE_TYPE_DOWNLOAD_SMALLER {
		f = new(SmallerFileWrapper)
		f.(*SmallerFileWrapper).ActualFile = actualFile
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

	err := db.DB.DatabaseHandle.Where("parent_uuid = ?", directoryDBO.UUID.String()).Find(&_files).Error()
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
