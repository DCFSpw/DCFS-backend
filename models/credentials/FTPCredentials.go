package credentials

import (
	"dcfs/apicalls"
	"fmt"
	"github.com/jlaffaye/ftp"
	"log"
	"strings"
	"time"
)

type FTPCredentials struct {
	User     string
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
	err = conn.Login(credentials.User, credentials.Password)
	if err != nil {
		log.Printf("Unable to login to FTP: %v", err)
		return nil
	}

	log.Printf("Connected to %s ...\n", credentials.Host)
	return conn
}

func (credentials *FTPCredentials) ToString() string {
	return credentials.User + ":" + credentials.Password + ":" + credentials.Host + ":" + credentials.Port + ":" + credentials.Path
}

func (credentials *FTPCredentials) GetPath() string {
	return credentials.Path
}

func NewFTPCredentials(cred string) *FTPCredentials {
	// string format: user:password:host:port:path
	parsed := strings.Split(cred, ":")
	credentials := FTPCredentials{
		User:     parsed[0],
		Password: parsed[1],
		Host:     parsed[2],
		Port:     parsed[3],
		Path:     parsed[4],
	}

	return &credentials
}
