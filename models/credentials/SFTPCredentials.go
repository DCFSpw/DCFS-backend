package credentials

import (
	"dcfs/apicalls"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

type SFTPCredentials struct {
	User     string
	Password string
	Host     string
	Port     string
	Client   *sftp.Client
}

func (credentials *SFTPCredentials) Authenticate(md *apicalls.CredentialsAuthenticateMetadata) *http.Client {
	log.Printf("Connecting to %s ...\n", credentials.Host)

	// Try to use $SSH_AUTH_SOCK which contains the path of the unix file socket that the sshd agent uses
	// for communication with other processes.
	var auths []ssh.AuthMethod
	if aconn, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
		auths = append(auths, ssh.PublicKeysCallback(agent.NewClient(aconn).Signers))
	}

	// Use password authentication if provided
	if credentials.Password != "" {
		auths = append(auths, ssh.Password(credentials.Password))
	}

	// Prepare client configuration
	config := ssh.ClientConfig{
		User:            credentials.User,
		Auth:            auths,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         30 * time.Second,
	}

	// Prepare SFTP server address
	addr := fmt.Sprintf("%s:%s", credentials.Host, credentials.Port)

	// Connect to SFTP server
	conn, err := ssh.Dial("tcp", addr, &config)
	if err != nil {
		log.Fatalf("Failed to connect to SFTP server [%s]: %v", addr, err)
		return nil
	}

	// Create new SFTP client
	sftpClient, err := sftp.NewClient(conn)
	if err != nil {
		log.Fatalf("Unable to create SFTP client: %v", err)
		return nil
	}
	credentials.Client = sftpClient
	//defer credentials.Client.Close()

	log.Printf("Connected to %s ...\n", credentials.Host)
	return nil
}

func (credentials *SFTPCredentials) ToString() string {
	return credentials.User + ":" + credentials.Password + ":" + credentials.Host + ":" + credentials.Port
}

func NewSFTPCredentials(cred string) *SFTPCredentials {
	// string format: user:password:host:port
	parsed := strings.Split(cred, ":")
	credentials := SFTPCredentials{
		User:     parsed[0],
		Password: parsed[1],
		Host:     parsed[2],
		Port:     parsed[3],
	}

	return &credentials
}
