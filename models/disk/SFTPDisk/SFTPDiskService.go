package SFTPDisk

import (
	"bytes"
	"dcfs/models/credentials"
	"fmt"
	"github.com/google/uuid"
	"io"
	"os"
)

func (d *SFTPDisk) connect(c credentials.Credentials) error {
	// Authenticate and connect to SFTP server
	err := c.Authenticate(nil)
	if err != nil {
		return err
	}

	// Save credentials
	d.credentials = c.(*credentials.SFTPCredentials)

	return nil
}

func (d *SFTPDisk) upload(fileName uuid.UUID, fileContents *[]byte) error {
	// Create remote file
	remoteFile, err := d.credentials.Client.OpenFile(fileName.String(), os.O_WRONLY|os.O_CREATE|os.O_TRUNC)
	if err != nil {
		return fmt.Errorf("Cannot o open remote file: %v", err)
	}
	defer remoteFile.Close()

	_, err = io.Copy(remoteFile, bytes.NewReader(*fileContents))
	if err != nil {
		return fmt.Errorf("Cannot upload local file: %v", err)
	}

	return nil
}

func (d *SFTPDisk) download(fileName uuid.UUID, fileContents *[]byte) error {
	// Open remote file
	remoteFile, err := d.credentials.Client.OpenFile(fileName.String(), os.O_RDONLY)
	if err != nil {
		return fmt.Errorf("Cannot open remote file: %v", err)
	}
	defer remoteFile.Close()

	// Download remote file
	buff := bytes.NewBuffer(*fileContents)
	_, err = io.Copy(buff, remoteFile)
	if err != nil {
		return fmt.Errorf("Cannot download remote file: %v", err)
	}

	return nil
}

func (d *SFTPDisk) rename(oldName uuid.UUID, newName uuid.UUID) error {
	err := d.credentials.Client.Rename(oldName.String(), newName.String())
	if err != nil {
		return err
	}

	return nil
}

func (d *SFTPDisk) remove(fileName uuid.UUID) error {
	err := d.credentials.Client.Remove(fileName.String())
	if err != nil {
		return err
	}

	return nil
}
