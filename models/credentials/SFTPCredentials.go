package credentials

import (
	"fmt"
	"golang.org/x/net/context"
	"log"
	"net"
	"os"
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

func (credentials *SFTPCredentials) Authenticate(ctx context.Context) error {
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
		return err
	}

	// Create new SFTP client
	sftpClient, err := sftp.NewClient(conn)
	if err != nil {
		log.Fatalf("Unable to create SFTP client: %v", err)
		return err
	}
	credentials.Client = sftpClient
	defer credentials.Client.Close()

	log.Printf("Connected to %s ...\n", credentials.Host)
	return nil
}
