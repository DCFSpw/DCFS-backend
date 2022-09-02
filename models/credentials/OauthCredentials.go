package credentials

import (
	"golang.org/x/net/context"
	"time"
)

type OauthCredentials struct {
	Token      string
	ExpiryDate time.Time
}

func (credentials *OauthCredentials) Authenticate(ctx context.Context) error {
	return nil
}
