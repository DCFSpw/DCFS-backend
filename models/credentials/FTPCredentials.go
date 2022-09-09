package credentials

import (
	"fmt"
	"github.com/jlaffaye/ftp"
	"golang.org/x/net/context"
	"log"
	"strings"
	"time"
)

type FTPCredentials struct {
	User     string
	Password string
	Host     string
	Port     string
	Client   *ftp.ServerConn
}

func (credentials *FTPCredentials) Authenticate(ctx context.Context) error {
	log.Printf("Connecting to %s ...\n", credentials.Host)

	// Prepare FTP server address
	addr := fmt.Sprintf("%s:%s", credentials.Host, credentials.Port)

	// Connect to FTP server
	conn, err := ftp.Dial(addr, ftp.DialWithTimeout(5*time.Second))
	if err != nil {
		log.Fatalf("Failed to connect to FTP server [%s]: %v", addr, err)
		return err
	}

	// Login to FTP server
	err = conn.Login(credentials.User, credentials.Password)
	if err != nil {
		log.Fatalf("Unable to login to FTP: %v", err)
		return err
	}
	credentials.Client = conn
	//defer credentials.Client.Close()

	log.Printf("Connected to %s ...\n", credentials.Host)
	return nil
}

func (credentials *FTPCredentials) ToString() string {
	return credentials.User + ":" + credentials.Password + ":" + credentials.Host + ":" + credentials.Port
}

func NewFTPCredentials(cred string) *FTPCredentials {
	// string format: user:password:host:port
	parsed := strings.Split(cred, ":")
	credentials := FTPCredentials{
		User:     parsed[0],
		Password: parsed[1],
		Host:     parsed[2],
		Port:     parsed[3],
	}

	return &credentials
}
