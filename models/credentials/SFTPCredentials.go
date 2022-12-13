package credentials

import (
	"dcfs/apicalls"
	"dcfs/util/logger"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

type SFTPCredentials struct {
	FTPCredentials

	SSHConnection *ssh.Client
}

// Authenticate - authenticate to remote server using saved credentials
//
// params:
//   - md *apicalls.CredentialsAuthenticateMetadata: not used
//
// return type:
//   - *SFTPCredentials: SFTP client object
func (credentials *SFTPCredentials) Authenticate(md *apicalls.CredentialsAuthenticateMetadata) interface{} {
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
		logger.Logger.Error("credentials", "Failed to connect to: ", credentials.Host, " to authenticate an SSH operation.")
		return nil
	}

	// Create new SFTP client
	sftpClient, err := sftp.NewClient(conn)
	if err != nil {
		logger.Logger.Error("credentials", "Could not create a new SFTP server instance.")
		return nil
	}

	credentials.SSHConnection = conn
	return sftpClient
}

// ToString - convert credentials to JSON string
//
// return type:
//   - string: JSON credential string
func (credentials *SFTPCredentials) ToString() string {
	return credentials.FTPCredentials.ToString()
}

// GetPath - get remote path from credentials
//
// return type:
//   - string: remote path
func (credentials *SFTPCredentials) GetPath() string {
	return credentials.Path
}

// NewSFTPCredentials - create new SFTP credentials object based on JSON credential string
//
// params:
//   - cred string: JSON credential string
//
// return type:
//   - *SFTPCredentials: created credentials object
func NewSFTPCredentials(cred string) *SFTPCredentials {
	credentials := SFTPCredentials{
		FTPCredentials: *NewFTPCredentials(cred),
	}

	return &credentials
}
