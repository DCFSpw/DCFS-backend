package mock

import (
	"dcfs/db"
	"dcfs/db/dbo"
	"github.com/DATA-DOG/go-sqlmock"
	_ "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DBMock sqlmock.Sqlmock

var DiskColumns []string = []string{"uuid", "user_uuid", "volume_uuid", "provider_uuid", "credentials", "name", "created_at", "used_space", "total_space"}

var VolumeColumns []string = []string{"uuid", "name", "user_uuid", "backup", "encryption", "file_partition", "created_at", "deleted_at"}

var BlockColumns []string = []string{"uuid", "user_uuid", "volume_uuid", "disk_uuid", "file_uuid", "size", "order", "checksum"}

var ProviderColumns []string = []string{"uuid", "type", "name", "logo"}

var UserColumns []string = []string{"uuid", "first_name", "last_name", "email", "password"}

func DiskRow(_dbos ...*dbo.Disk) *sqlmock.Rows {
	ret := sqlmock.NewRows(DiskColumns)

	for _, _dbo := range _dbos {
		if _dbo == nil {
			continue
		}

		ret.AddRow(
			_dbo.UUID,
			_dbo.UserUUID,
			_dbo.VolumeUUID,
			_dbo.ProviderUUID,
			_dbo.Credentials,
			_dbo.Name,
			_dbo.CreatedAt,
			_dbo.UsedSpace,
			_dbo.TotalSpace)
	}

	return ret
}

func VolumeRow(_dbos ...*dbo.Volume) *sqlmock.Rows {
	ret := sqlmock.NewRows(VolumeColumns)

	for _, _dbo := range _dbos {
		if _dbo == nil {
			continue
		}

		ret.AddRow(
			_dbo.UUID,
			_dbo.Name,
			_dbo.UserUUID,
			_dbo.VolumeSettings.Backup,
			_dbo.VolumeSettings.Encryption,
			_dbo.VolumeSettings.FilePartition,
			_dbo.CreatedAt,
			_dbo.DeletedAt)
	}

	return ret
}

func BlockRow(_dbos ...*dbo.Block) *sqlmock.Rows {
	ret := sqlmock.NewRows(BlockColumns)

	for _, _dbo := range _dbos {
		if _dbo == nil {
			continue
		}

		ret.AddRow(
			_dbo.UUID,
			_dbo.UserUUID,
			_dbo.VolumeUUID,
			_dbo.DiskUUID,
			_dbo.FileUUID,
			_dbo.Size,
			_dbo.Order,
			_dbo.Checksum)
	}

	return ret
}

func ProviderRow(_dbos ...*dbo.Provider) *sqlmock.Rows {
	ret := sqlmock.NewRows(ProviderColumns)

	for _, _dbo := range _dbos {
		if _dbo == nil {
			continue
		}

		ret.AddRow(
			_dbo.UUID,
			_dbo.Type,
			_dbo.Name,
			_dbo.Logo)
	}

	return ret
}

func UserRow(_dbos ...*dbo.User) *sqlmock.Rows {
	ret := sqlmock.NewRows(UserColumns)

	for _, _dbo := range _dbos {
		if _dbo == nil {
			continue
		}

		ret.AddRow(
			_dbo.UUID,
			_dbo.FirstName,
			_dbo.LastName,
			_dbo.Email,
			_dbo.Password)
	}

	return ret
}

func init() {
	_db, _mock, _ := sqlmock.New()
	_mock.MatchExpectationsInOrder(false)
	db.DB.DatabaseHandle, _ = gorm.Open(mysql.New(mysql.Config{
		Conn:                      _db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	DBMock = _mock
}
