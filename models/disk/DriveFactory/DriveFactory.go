package DriveFactory

import (
	"dcfs/constants"
	"dcfs/models/disk"
	"dcfs/models/disk/GDriveDisk"
	"dcfs/models/disk/OneDriveDisk"
	"dcfs/models/disk/SFTPDisk"
)

func NewDisk(providerType int) disk.Disk {
	switch providerType {
	case constants.PROVIDER_TYPE_SFTP:
		return SFTPDisk.NewSFTPDisk()
	case constants.PROVIDER_TYPE_GDRIVE:
		return GDriveDisk.NewGDriveDisk()
	case constants.PROVIDER_TYPE_ONEDRIVE:
		return OneDriveDisk.NewOneDriveDisk()
	default:
		return nil
	}
}
