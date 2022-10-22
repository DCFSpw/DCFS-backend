package credentials

import (
	"dcfs/apicalls"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

type SFTPCredentials struct {
	FTPCredentials
}

func (credentials *SFTPCredentials) Authenticate(md *apicalls.CredentialsAuthenticateMetadata) interface{} {
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
		User:            credentials.Login,
		Auth:            auths,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         30 * time.Second,
	}

	// Prepare SFTP server address
	addr := fmt.Sprintf("%s:%s", credentials.Host, credentials.Port)

	// Connect to SFTP server
	conn, err := ssh.Dial("tcp", addr, &config)
	if err != nil {
		log.Printf("Failed to connect to SFTP server [%s]: %v", addr, err)
		return nil
	}

	// Create new SFTP client
	sftpClient, err := sftp.NewClient(conn)
	if err != nil {
		log.Printf("Unable to create SFTP client: %v", err)
		return nil
	}
	//defer credentials.Client.Close()

	log.Printf("Connected to %s ...\n", credentials.Host)

	return sftpClient
}

func (credentials *SFTPCredentials) ToString() string {
	return credentials.FTPCredentials.ToString()
}

func (credentials *SFTPCredentials) GetPath() string {
	return credentials.Path
}

func NewSFTPCredentials(cred string) *SFTPCredentials {
	credentials := SFTPCredentials{
		FTPCredentials: *NewFTPCredentials(cred),
	}

	return &credentials
}
