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

// NewBlock - create new block model based on provided data
//
// This function creates block model used internally by backend based on
// block data obtained from database.
//
// params:
//   - _UUID uuid.UUID: block UUID
//   - _userUUID uuid.UUID: UUID of the user who owns the block
//   - _file File: UUID of the file to which the block belongs
//   - _disk Disk: UUID of the disk on which the block is stored
//   - _size int: size of the block in bytes
//   - _checksum string: checksum of the block
//   - _status int: transport status of the block
//   - _order int: order of the block in the file
//
// return type:
//   - *models.Block: created block model
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

// NewBlockFromDBO - create new block model based on block DBO
//
// This function creates block model used internally by backend based on
// block data obtained from database.
//
// params:
//   - _block *dbo.Block: block DBO data (from database)
//
// return type:
//   - *models.Block: created block model
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
