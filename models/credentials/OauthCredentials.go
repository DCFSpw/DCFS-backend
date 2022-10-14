package credentials

import (
	"dcfs/apicalls"
	"dcfs/db"
	"dcfs/db/dbo"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"net/http"
	"strings"
	"time"
)

type OauthCredentials struct {
	Token *oauth2.Token
}

func (credentials *OauthCredentials) Authenticate(md *apicalls.CredentialsAuthenticateMetadata) interface{} {
	var config *oauth2.Config = md.Config
	var ret *http.Client
	credentials.PerformOp(
		func(token *oauth2.Token) {
			ret = config.Client(md.Ctx, credentials.Token)
		}, md.DiskUUID)
	//return config.Client(md.Ctx, credentials.Token)
	return ret
	//if credentials.Token.Valid() {
	//	return nil
	//}

	/*
		json, err := os.ReadFile("credentials.json")
		if err != nil {
			fmt.Println("Reading json file failed with error: ", err)
			return err
		}

		var config *oauth2.Config
		config, err = google.ConfigFromJSON(json, drive.DriveMetadataReadonlyScope)
		if err != nil {
			fmt.Println("Parsing configuration json failed with err: ", err)
			return err
		}
	*/

	//return nil
}

// PerformOp
//
// Performs an operation on the token handle and updates the token in the DB if needed
func (credentials *OauthCredentials) PerformOp(operator func(token *oauth2.Token), diskUUID uuid.UUID) {
	var _disk dbo.Disk = dbo.Disk{}
	operator(credentials.Token)

	db.DB.DatabaseHandle.Where("uuid = ?", diskUUID.String()).First(&_disk)
	if _disk.Credentials != credentials.ToString() {
		_disk.Credentials = credentials.ToString()
		db.DB.DatabaseHandle.Save(&_disk)
	}
}

func (credentials *OauthCredentials) ToString() string {
	return credentials.Token.AccessToken + ":" + credentials.Token.RefreshToken
}

func NewOauthCredentials(str string) *OauthCredentials {
	var credentials *OauthCredentials = new(OauthCredentials)
	tokens := strings.Split(str, ":")
	if len(tokens) < 2 {
		return nil
	}

	credentials.Token = &oauth2.Token{AccessToken: tokens[0], RefreshToken: tokens[1]}

	// invalidate token
	credentials.Token.Expiry = time.Now()
	return credentials
}
