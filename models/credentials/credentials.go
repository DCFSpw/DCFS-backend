package credentials

import (
	"dcfs/apicalls"
)

type Credentials interface {
	Authenticate(md *apicalls.CredentialsAuthenticateMetadata) interface{}
	ToString() string
	GetPath() string
}
