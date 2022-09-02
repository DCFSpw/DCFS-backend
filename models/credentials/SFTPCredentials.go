package credentials

import "golang.org/x/net/context"

type SFTPCredentials struct {
	User     string
	Password string
	Host     string
	Port     string
}

func (credentials *SFTPCredentials) Authenticate(ctx context.Context) error {
	return nil
}
