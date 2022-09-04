package credentials

import "golang.org/x/net/context"

const (
	CREDENTIALS_SFTP  int = 0
	CREDENTIALS_OAUTH int = 1
)

type Credentials interface {
	Authenticate(ctx context.Context) error
	ToString() string
}
