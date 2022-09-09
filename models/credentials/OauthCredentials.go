package credentials

import (
	"dcfs/apicalls"
	"golang.org/x/oauth2"
	"net/http"
	"strings"
)

type OauthCredentials struct {
	Token oauth2.Token
}

func (credentials *OauthCredentials) Authenticate(md *apicalls.CredentialsAuthenticateMetadata) *http.Client {
	var config *oauth2.Config = md.Config
	return config.Client(md.Ctx, &credentials.Token)

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

func (credentials *OauthCredentials) ToString() string {
	return credentials.Token.AccessToken + ":" + credentials.Token.RefreshToken
}

func NewOauthCredentials(str string) *OauthCredentials {
	var credentials *OauthCredentials = new(OauthCredentials)
	tokens := strings.Split(str, ":")
	credentials.Token = oauth2.Token{AccessToken: tokens[0], RefreshToken: tokens[1]}
	return credentials
}
