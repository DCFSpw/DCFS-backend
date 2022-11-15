package credentials

import (
	"dcfs/apicalls"
	"dcfs/db"
	"dcfs/db/dbo"
	"dcfs/requests"
	"encoding/json"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"net/http"
	"time"
)

type OauthCredentials struct {
	Token *oauth2.Token
}

// Authenticate - authenticate to remote server using saved credentials
//
// params:
//   - md *apicalls.CredentialsAuthenticateMetadata: internal metadata (with current context)
//
// return type:
//   - *http.Client: HTTP client object
func (credentials *OauthCredentials) Authenticate(md *apicalls.CredentialsAuthenticateMetadata) interface{} {
	var config *oauth2.Config = md.Config
	var ret *http.Client

	credentials.performOperation(
		func(token *oauth2.Token) {
			ret = config.Client(md.Ctx, credentials.Token)
		}, md.DiskUUID)

	return ret
}

// performOperation - perform an operation on the token handle and updates the token in the DB if needed
func (credentials *OauthCredentials) performOperation(operator func(token *oauth2.Token), diskUUID uuid.UUID) {
	var _disk dbo.Disk = dbo.Disk{}
	operator(credentials.Token)

	db.DB.DatabaseHandle.Where("uuid = ?", diskUUID.String()).First(&_disk)
	if _disk.Credentials != credentials.ToString() {
		_disk.Credentials = credentials.ToString()
		db.DB.DatabaseHandle.Save(&_disk)
	}
}

// ToString - convert credentials to JSON string
//
// return type:
//   - string: JSON credential string
func (credentials *OauthCredentials) ToString() string {
	var _cred *requests.OAuthCredentials = &requests.OAuthCredentials{AccessToken: credentials.Token.AccessToken, RefreshToken: credentials.Token.RefreshToken}
	str, _ := json.Marshal(_cred)

	return string(str)
}

// GetPath - not supported for OAuth credentials
func (credentials *OauthCredentials) GetPath() string {
	panic("OAuth credentials do not return a path!")
}

// NewOauthCredentials - create new OAuth credentials object based on JSON credential string
//
// params:
//   - cred string: JSON credential string
//
// return type:
//   - *OauthCredentials: created credentials object
func NewOauthCredentials(cred string) *OauthCredentials {
	var _credentials *requests.OAuthCredentials = requests.StringToOAuthCredentials(cred)
	var credentials *OauthCredentials = &OauthCredentials{}

	credentials.Token = &oauth2.Token{AccessToken: _credentials.AccessToken, RefreshToken: _credentials.RefreshToken}

	// Invalidate token
	credentials.Token.Expiry = time.Now()
	return credentials
}
