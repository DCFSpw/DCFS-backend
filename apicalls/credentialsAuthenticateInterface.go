package apicalls

import (
	"context"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

type CredentialsAuthenticateMetadata struct {
	Ctx      context.Context
	Config   *oauth2.Config
	DiskUUID uuid.UUID
}
