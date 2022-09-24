package models

import (
	"dcfs/models/disk"
	"github.com/google/uuid"
)

const (
	BLOCK_STATUS_QUEUED      int = 0
	BLOCK_STATUS_IN_PROGRESS int = 1
	BLOCK_STATUS_TRANSFERRED int = 2
	BLOCK_STATUS_FAILED      int = 3
)

type Block struct {
	UUID     uuid.UUID
	UserUUID uuid.UUID
	File     *File
	Disk     disk.Disk

	Size     int
	Checksum int

	Status int
	Order  int
}

func NewBlock(_UUID uuid.UUID, _userUUID uuid.UUID, _file *File, _disk disk.Disk, _size int, _checksum int, _status int, _order int) *Block {
	var block *Block = new(Block)
	block.UUID = _UUID
	block.UserUUID = _userUUID
	block.File = _file
	block.Disk = _disk
	block.Size = _size
	block.Checksum = _checksum
	block.Status = _status
	block.Order = _order

	return block
}
