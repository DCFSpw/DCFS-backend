package seeder

import (
	"dcfs/db"
	"dcfs/db/dbo"
	"dcfs/models"
)

// Seed - add necessary entries to database
//
// This function adds necessary entries to empty database or database which
// doesn't contain these entries. It creates providers for disks and sample
// user account with empty volume.
func Seed() {
	user := dbo.User{}

	// Add a root user
	db.DB.DatabaseHandle.Where("uuid = ?", models.RootUUID).First(&user)
	if user.UUID != models.RootUUID {
		user.UUID = models.RootUUID
		user.FirstName = "Root"
		user.LastName = "Root"
		user.Email = "root@root.com"
		user.Password = dbo.HashPassword("password")
		db.DB.DatabaseHandle.Create(&user)
	}

	// Add initialized providers
	for _, providerInit := range models.ProviderTypesRegistry {
		providerInit()
	}
}
