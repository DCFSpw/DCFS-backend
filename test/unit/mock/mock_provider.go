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
	/* SFTP */ "{\n        \"Login\": \"dcfs\",\n        \"Password\": \"UszatekM*01\",\n        \"Host\": \"34.116.204.182\",\n        \"Port\": \"2022\",\n        \"Path\": \"sftp\"\n }",
	/* GDrive */ "{\"accessToken\":\"ya29.a0AeTM1icyEKlMfnLIr1yUhopNFRFY5im7_IOCBgVl5sVfrRytmXOwO9nRW6IFzSyO1eQrwNcXibYhaI1r4sRmQe_695VM2vnRK1VFeAPOrX46v2iAlBnO9SbiftcUP9JGlXAIejPHNTF1um1Bo1KIHC89VRFMaCgYKARcSARESFQHWtWOmmOqpwZVxMGsFBckDEG9yKw0163\",\"refreshToken\":\"1//0ckpMy_6Z7b3KCgYIARAAGAwSNwF-L9IrhNPrRl1cDW3hTSKF5W31u9vPt_JHcE5XLMe85qPmHaNA-R8FGwu5hQB9Ka1auue-HVU\"}",
	/* OneDrive */ "{\"accessToken\":\"EwBoA8l6BAAUkj1NuJYtTVha+Mogk+HEiPbQo04AAchFRC5PDv5fTwCbjPv/WuQI09Q4Nw4n4OkJsc7NaYnfiC3dT6RkUCUOFdTeegkvpom4UjjIR+SXlkSintNxhBW2giJyTyuXWYrLzip1nz56XRc06i3oKfUMFkY/b7HkZa7KQoiItGV7OqznTv6lUm50qOyhzw7RuHU1sXQ5QSAtVtQlqOYI4O3+vOglcWK+AU6UEytcSbeIpHYbHY+WhEOodClRiTdeqe/IRPcLZeHCe6hkomFNoqJheFtwTpQirzCakNRLPE8uqNz4j4T8YLEeFhlwnQDFaNStVxWXw/V3lQjaWVZ+szJP+NhukBnwEVTiAwEHByKRYW37L+nyB0MDZgAACBWi7LHIguO0OAKCB85qK72YF/9oPRcPAs7hDNx8huTv9cBvwqFSbK9ohYQN/WqMyrwcKpGjtXYxmO9EaTcIFKM0ZWITf62zyQXEp/8p5qvqRL850D3i6f0C94dgGFpKYJaDvpQnU1cbt9d3KDWh3JnQ3P9dDmR2T92ZLXWcKIAa4GnfVknQfULlO4YFvc6MwGZtm57jBpAfZLJ6IlUhuDAHb6xdidcGKJQFh+GdWADQj5/yigBKOCjhUDzoYj0b+Mi+c8hTWOSw+4oBWW0kbmyEQebHF0G9OrIuW9Egm6NAFWvgQQOpalXwyDzUl82xRoNrPz4QKb4gWJrmLrF7PJoVm+V+Vh4MaWkSLtg0hnKZ/Sf4iGlHmD5J2FIyoKq0Q0WB5Khy43USKHsme63VdIzvB61LQ+N7Qqk9Q2RWl1DWYkbBPQLnb/UCf+DZhX2K8fdY6dwiNXya3zB54D+zqv2p0GB+c4jYC0jwhc+J4bYJA0Jt0EBWkoP/sikfD47faDA1C141WYEyT/DC0hx0ws/BZVW4BCxAZTXKR1CVKY+23R4b4tngE0LDcE1YYMViWyllGHSM+YsWr6Kglkqx+QFJaP9WetEp3BI9ocXohsEgKhCpdmE2MHQs/MNCfK7sHbEn54Araq+EZNSXrX8DMdgiaa4veDFEh/lSVike7G3/W/hryRREVLHxW6DatPDbhJBXQYdjYr2kDi8oM7s5/W/CBXRgVfH/0XClKML8bYXRDteJ3hRnBNrWvIc31FUhecvrcAI=\",\"refreshToken\":\"M.R3_BAY.-CQ7xzIkBvl*6*v0Vrzqq89mqsiHkVGnvJpVUUPoz3rjl76x1JDceujKm6Sef*FNJw47VrBCuj10bxzg7WNfX2hmDSjAqEYRzFxjKa2bH54JejTpP3CPrESkAQg79DMJMDyrxyyCNXbkGB8g6hdtBb7BoTN9tpYL7J2IVh9kwla7D2GxPv36hbxG208!5VK9mNR4N!qbpziol!nVbZBEoPM3xdYFt6ZP02NmCYjmstyoOmTS5VkKjOu9V2!J78z2v9bu0U!*r!JKXL3woLkHpegx8OnMmGJh*9XVk8QOom2Z8\"}",
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

func GetProviderDBO(p int) (*dbo.Provider, string) {
	r := 0

	if p == constants.PROVIDER_TYPE_GDRIVE {
		r = 1
	}

	if p == constants.PROVIDER_TYPE_ONEDRIVE {
		r = 2
	}

	return &dbo.Provider{
		AbstractDatabaseObject: dbo.AbstractDatabaseObject{
			UUID: ProviderUUIDs[r],
		},
		Type: Providers[r],
		Name: "Random Provider",
		Logo: "Random Provider logo",
	}, DummyCredentials[r]
}
