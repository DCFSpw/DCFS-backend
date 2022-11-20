package mock

import (
	"dcfs/constants"
	"dcfs/db/dbo"
	"github.com/google/uuid"
	"math/rand"
)

var SFTPProviderUUID uuid.UUID = uuid.New()
var FTPProviderUUID uuid.UUID = uuid.New()
var GDiskProviderUUI uuid.UUID = uuid.New()
var OneDriveProvider uuid.UUID = uuid.New()

var ProviderUUIDs []uuid.UUID = []uuid.UUID{
	SFTPProviderUUID,
	// FTPProviderUUID, /* to be implemented */
	GDiskProviderUUI,
	OneDriveProvider,
}

var Providers []int = []int{
	constants.PROVIDER_TYPE_SFTP,
	// constants.PROVIDER_TYPE_FTP, /* to be implemented */
	constants.PROVIDER_TYPE_GDRIVE,
	constants.PROVIDER_TYPE_ONEDRIVE,
}

var DummyCredentials []string = []string{
	/* SFTP */ "{\n        \"Login\": \"server720792_dcfs\",\n        \"Password\": \"UszatekM*00\",\n        \"Host\": \"ftp.server720792.nazwa.pl\",\n        \"Port\": \"22\",\n        \"Path\": \"\"\n    }",
	/* GDrive */ "{\"accessToken\":\"ya29.a0AeTM1idMIUcybNBRTIQ-v5F9ZQI30xrfW7iAxy1i72pOY60PzyGA2Pp7eqa5qXdjEjd-m20IVJo9LQXSJ4qEzAXaZaSAG6EmahVc6L2pIPioDpCkeSGactJOkiT_sJyOm3ss0YighVJWyhKu89MXbCz9FFhiaCgYKAcESARESFQHWtWOm37NYB3gIMOeF1LoyN9jcEQ0163\",\"refreshToken\":\"1//0cV-a2NWw9XaZCgYIARAAGAwSNwF-L9Ir8gcJloQ3ZeKARQM1X8OV0bqSof2CwpA198nQ0Ib8vNhv9Rviw9HUSBFk2hHUL5o0H6I\"}",
	/* OneDrive */ "{\"accessToken\":\"ya29.a0AeTM1idMIUcybNBRTIQ-v5F9ZQI30xrfW7iAxy1i72pOY60PzyGA2Pp7eqa5qXdjEjd-m20IVJo9LQXSJ4qEzAXaZaSAG6EmahVc6L2pIPioDpCkeSGactJOkiT_sJyOm3ss0YighVJWyhKu89MXbCz9FFhiaCgYKAcESARESFQHWtWOm37NYB3gIMOeF1LoyN9jcEQ0163\",\"refreshToken\":\"1//0cV-a2NWw9XaZCgYIARAAGAwSNwF-L9Ir8gcJloQ3ZeKARQM1X8OV0bqSof2CwpA198nQ0Ib8vNhv9Rviw9HUSBFk2hHUL5o0H6I\"}",
}

func GetRandomProviderIdx() int {
	return rand.Int() % len(Providers)
}

func GetRandomProviderDBO() (*dbo.Provider, string) {
	r := GetRandomProviderIdx()

	return &dbo.Provider{
		AbstractDatabaseObject: dbo.AbstractDatabaseObject{
			UUID: ProviderUUIDs[r],
		},
		Type: Providers[r],
		Name: "Random Provider",
		Logo: "Random Provider logo",
	}, DummyCredentials[r]
}
