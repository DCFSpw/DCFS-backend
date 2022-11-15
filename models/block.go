package models

import (
	"dcfs/db/dbo"
	"github.com/google/uuid"
)

type Block struct {
	UUID     uuid.UUID
	UserUUID uuid.UUID
	File     File
	Disk     Disk

	Size     int
	Checksum string

	Status int
	Order  int
}

func NewBlock(_UUID uuid.UUID, _userUUID uuid.UUID, _file File, _disk Disk, _size int, _checksum string, _status int, _order int) *Block {
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

func NewBlockFromDBO(_block *dbo.Block) *Block {
	return &Block{
		UUID:     _block.UUID,
		UserUUID: _block.UserUUID,
		File:     nil,
		Disk:     nil,
		Size:     _block.Size,
		Checksum: "",
		Status:   0,
		Order:    _block.Order,
	}
}
