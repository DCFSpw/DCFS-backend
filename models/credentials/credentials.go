package credentials

import (
	"dcfs/apicalls"
)

const (
	CREDENTIALS_SFTP  int = 0
	CREDENTIALS_OAUTH int = 1
)

type Credentials interface {
	Authenticate(md *apicalls.CredentialsAuthenticateMetadata) interface{}
	ToString() string
	GetPath() string
}
