package credentials

import (
	"dcfs/apicalls"
	"dcfs/requests"
	"encoding/json"
	"fmt"
	"github.com/jlaffaye/ftp"
	"log"
	"time"
)

type FTPCredentials struct {
	Login    string
	Password string
	Host     string
	Port     string
	Path     string
}

func (credentials *FTPCredentials) Authenticate(md *apicalls.CredentialsAuthenticateMetadata) interface{} {
	log.Printf("Connecting to %s ...\n", credentials.Host)

	// Prepare FTP server address
	addr := fmt.Sprintf("%s:%s", credentials.Host, credentials.Port)

	// Connect to FTP server
	conn, err := ftp.Dial(addr, ftp.DialWithTimeout(5*time.Second))
	if err != nil {
		log.Printf("Failed to connect to FTP server [%s]: %v", addr, err)
		return nil
	}

	// Login to FTP server
	err = conn.Login(credentials.Login, credentials.Password)
	if err != nil {
		log.Printf("Unable to login to FTP: %v", err)
		return nil
	}

	log.Printf("Connected to %s ...\n", credentials.Host)
	return conn
}

func (credentials *FTPCredentials) ToString() string {
	ret, _ := json.Marshal(credentials)
	return string(ret)
}

func (credentials *FTPCredentials) GetPath() string {
	return credentials.Path
}

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
