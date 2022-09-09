package credentials

import (
	"dcfs/apicalls"
	"net/http"
)

const (
	CREDENTIALS_SFTP  int = 0
	CREDENTIALS_OAUTH int = 1
)

type Credentials interface {
	Authenticate(md *apicalls.CredentialsAuthenticateMetadata) *http.Client
	ToString() string
}
