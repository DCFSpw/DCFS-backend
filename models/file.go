package models

import (
	"dcfs/apicalls"
	"dcfs/constants"
	"github.com/google/uuid"
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

	GetVolume() *Volume
	SetVolume(v *Volume)

	IsCompleted() bool
	GetBlocks() *[]*Block

	Remove()
}

type AbstractFile struct {
	UUID   uuid.UUID
	Name   string
	Type   int
	Size   int
	Parent *File
	Volume *Volume
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

func (file *AbstractFile) IsCompleted() bool {
	panic("Unimplemented abstract method")
}

func (file *AbstractFile) GetVolume() *Volume {
	return file.Volume
}

func (file *AbstractFile) SetVolume(v *Volume) {
	file.Volume = v
}

func (file *AbstractFile) GetBlocks() *[]*Block {
	return nil
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

func (file *RegularFile) GetBlocks() *[]*Block {
	var blocks []*Block = nil

	for _, block := range file.Blocks {
		blocks = append(blocks, block)
	}

	return &blocks
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

func (d *Directory) IsCompleted() bool {
	return true
}

func (d *Directory) GetVolume() *Volume {
	return d.AbstractFile.GetVolume()
}

func (d *Directory) SetVolume(v *Volume) {
	d.AbstractFile.SetVolume(v)
}

func (d *Directory) GetBlocks() *[]*Block {
	return d.AbstractFile.GetBlocks()
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

func NewFileFromReq(req *apicalls.FileUploadRequest) File {
	var f File = NewFile(req.Type)
	f.SetName(req.Name)
	f.SetSize(req.Size)

	return f
}
