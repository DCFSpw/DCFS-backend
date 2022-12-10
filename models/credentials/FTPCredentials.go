package credentials

import (
	"dcfs/apicalls"
	"dcfs/requests"
	"dcfs/util/logger"
	"encoding/json"
	"fmt"
	"github.com/jlaffaye/ftp"
	"time"
)

type FTPCredentials struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     string `json:"port"`
	Path     string `json:"path"`
}

// Authenticate - authenticate to remote server using saved credentials
//
// params:
//   - md *apicalls.CredentialsAuthenticateMetadata: not used
//
// return type:
//   - *FTPCredentials: FTP client object
func (credentials *FTPCredentials) Authenticate(md *apicalls.CredentialsAuthenticateMetadata) interface{} {
	logger.Logger.Debug("credentials", "Connecting to ", credentials.Host, "...")

	// Prepare FTP server address
	addr := fmt.Sprintf("%s:%s", credentials.Host, credentials.Port)

	// Connect to FTP server
	conn, err := ftp.Dial(addr, ftp.DialWithTimeout(5*time.Second))
	if err != nil {
		logger.Logger.Error("credentials", "Failed to connect to the FTP server: ", addr, ". Got an error: ", err.Error(), ". ")
		return nil
	}

	// Login to FTP server
	err = conn.Login(credentials.Login, credentials.Password)
	if err != nil {
		logger.Logger.Error("credentials", "Unable to login into the FTP disk, got an error: ", err.Error())
		return nil
	}

	logger.Logger.Debug("credentials", "Connected to: ", credentials.Host)
	return conn
}

// ToString - convert credentials to JSON string
//
// return type:
//   - string: JSON credential string
func (credentials *FTPCredentials) ToString() string {
	ret, _ := json.Marshal(credentials)
	return string(ret)
}

// GetPath - get remote path from credentials
//
// return type:
//   - string: remote path
func (credentials *FTPCredentials) GetPath() string {
	return credentials.Path
}

// NewFTPCredentials - create new FTP credentials object based on JSON credential string
//
// params:
//   - cred string: JSON credential string
//
// return type:
//   - *FTPCredentials: created credentials object
func NewFTPCredentials(cred string) *FTPCredentials {
	var _credentials *requests.FTPCredentials = requests.StringToFTPCredentials(cred)

	credentials := FTPCredentials{
		Login:    _credentials.Login,
		Password: _credentials.Password,
		Host:     _credentials.Host,
		Port:     _credentials.Port,
		Path:     _credentials.Path,
	}

	return &credentials
}
