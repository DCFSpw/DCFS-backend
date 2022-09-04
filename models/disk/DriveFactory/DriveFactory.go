package DriveFactory

import (
	"dcfs/db/dbo"
	"dcfs/models/disk"
	"dcfs/models/disk/GDriveDisk"
	"dcfs/models/disk/OneDriveDisk"
	"dcfs/models/disk/SFTPDisk"
)

func NewDisk(providerType int) disk.Disk {
	switch providerType {
	case dbo.SFTP:
		return SFTPDisk.NewSFTPDisk()
	case dbo.GDRIVE:
		return GDriveDisk.NewGDriveDisk()
	case dbo.ONEDRIVE:
		return OneDriveDisk.NewOneDriveDisk()
	default:
		return nil
	}
}
