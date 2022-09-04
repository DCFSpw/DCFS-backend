package credentials

import (
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
)

type OauthCredentials struct {
	Token oauth2.Token
}

func (credentials *OauthCredentials) Authenticate(ctx context.Context) error {
	panic("Unimplemented")

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

	return nil
}

func (credentials *OauthCredentials) ToString() string {
	return credentials.Token.AccessToken
}

func NewOauthCredentials(token string) *OauthCredentials {
	var credentials *OauthCredentials = new(OauthCredentials)
	credentials.Token = oauth2.Token{AccessToken: token}
	return credentials
}
