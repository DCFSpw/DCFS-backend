package OneDriveDisk

import (
	"bytes"
	"dcfs/apicalls"
	"dcfs/constants"
	"dcfs/db/dbo"
	"dcfs/models"
	"dcfs/models/credentials"
	"dcfs/models/disk/AbstractDisk"
	"dcfs/responses"
	"fmt"
	"github.com/goh-chunlin/go-onedrive/onedrive"
	"github.com/google/uuid"
	"github.com/h2non/filetype"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

type OneDriveDisk struct {
	abstractDisk AbstractDisk.AbstractDisk
}

/* Mandatory Disk interface implementations */

func (d *OneDriveDisk) Upload(blockMetadata *apicalls.BlockMetadata) error {
	var _client interface{} = d.GetCredentials().Authenticate(&apicalls.CredentialsAuthenticateMetadata{Ctx: blockMetadata.Ctx, Config: d.GetConfig(), DiskUUID: d.GetUUID()})
	if _client == nil {
		return fmt.Errorf("could not connect to the remote server")
	}

	var client *http.Client = _client.(*http.Client)
	oneDriveClient := onedrive.NewClient(client)

	// size in bytes
	var size int = len(*blockMetadata.Content)
	var apiURL string = "me/drive/root:/" + url.PathEscape(blockMetadata.UUID.String())

	ft, err := filetype.Match(*blockMetadata.Content)
	if err != nil {
		return fmt.Errorf("file %s is corrupted", blockMetadata.FileUUID.String())
	}
	err = nil

	if size <= constants.ONEDRIVE_SIZE_LIMIT {
		// fast upload
		req, err := oneDriveClient.NewFileUploadRequest(apiURL+":/content?@microsoft.graph.conflictBehavior=rename", ft.MIME.Value, bytes.NewReader(*blockMetadata.Content))
		if err != nil {
			return err
		}
		err = nil

		var response *onedrive.DriveItem
		err = oneDriveClient.Do(blockMetadata.Ctx, req, false, &response)
		if err != nil {
			return err
		}
		err = nil
	} else {
		// upload session
		url, err := oneDriveClient.BaseURL.Parse(apiURL + ":/createUploadSession")
		err = nil

		req, err := http.NewRequest("POST", url.String(), nil)
		req.Header.Set("Content-Type", "application/json")
		//req, err := oneDriveClient.NewFileUploadRequest(apiURL+":/createUploadSession", "application/json", nil)
		if err != nil {
			return err
		}
		err = nil

		var response *responses.CreateUploadSessionResponse
		err = oneDriveClient.Do(blockMetadata.Ctx, req, false, &response)
		if err != nil {
			return err
		}
		err = nil

		// fill the end of file with 0, so the byte number is divisible by 320 KiB
		remainder := len(*blockMetadata.Content) % (320 * 1024)
		if remainder > 0 {
			complement := make([]uint8, 320*1024-remainder)
			*blockMetadata.Content = append(*blockMetadata.Content, complement...)
		}
		size = len(*blockMetadata.Content)

		for i := 0; i < size; i = i + constants.ONEDRIVE_UPLOAD_LIMIT {
			upperBound := int(math.Min(float64(i+constants.ONEDRIVE_UPLOAD_LIMIT), float64(size)))

			request, err := http.NewRequest("PUT", response.UploadUrl, bytes.NewReader((*blockMetadata.Content)[i:upperBound]))
			request.Header.Set("Content-Length", strconv.Itoa(upperBound-i))
			request.Header.Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", i, upperBound-1, len(*blockMetadata.Content)))
			if err != nil {
				return err
			}
			err = nil

			if upperBound < size {
				rsp := responses.UploadSessionResponse{}
				err = oneDriveClient.Do(blockMetadata.Ctx, request, false, &rsp)
				if err != nil {
					return err
				}
			} else {
				rsp := responses.UploadSessionFinalResponse{}
				err = oneDriveClient.Do(blockMetadata.Ctx, request, false, &rsp)
				if err != nil {
					return err
				}
			}
			err = nil
		}
	}

	blockMetadata.CompleteCallback(blockMetadata.FileUUID, blockMetadata.Status)
	return nil
}

func (d *OneDriveDisk) Download(bm *apicalls.BlockMetadata) error {
	panic("unimplemented")
}

func (d *OneDriveDisk) Rename(blockMetadata *apicalls.BlockMetadata) error {
	panic("unimplemented")
}

func (d *OneDriveDisk) Remove(blockMetadata *apicalls.BlockMetadata) error {
	panic("unimplemented")
}

func (d *OneDriveDisk) SetUUID(uuid uuid.UUID) {
	d.abstractDisk.SetUUID(uuid)
}

func (d *OneDriveDisk) GetUUID() uuid.UUID {
	return d.abstractDisk.GetUUID()
}

func (d *OneDriveDisk) SetVolume(volume *models.Volume) {
	d.abstractDisk.SetVolume(volume)
}

func (d *OneDriveDisk) GetVolume() *models.Volume {
	return d.abstractDisk.GetVolume()
}

func (d *OneDriveDisk) GetCredentials() credentials.Credentials {
	return d.abstractDisk.GetCredentials()
}

func (d *OneDriveDisk) SetCredentials(credentials credentials.Credentials) {
	d.abstractDisk.SetCredentials(credentials)
}

func (d *OneDriveDisk) CreateCredentials(c string) {
	d.abstractDisk.Credentials = credentials.NewOauthCredentials(c)
}

func (d *OneDriveDisk) GetProviderUUID() uuid.UUID {
	return d.abstractDisk.GetProvider(constants.PROVIDER_TYPE_ONEDRIVE)
}

func (d *OneDriveDisk) GetDiskDBO(userUUID uuid.UUID, providerUUID uuid.UUID, volumeUUID uuid.UUID) dbo.Disk {
	return d.abstractDisk.GetDiskDBO(userUUID, providerUUID, volumeUUID)
}

func (d *OneDriveDisk) Delete() (string, error) {
	return d.abstractDisk.Delete()
}

/* Mandatory OAuthDisk interface implementations */
func NewOneDriveDisk() *OneDriveDisk {
	var d *OneDriveDisk = new(OneDriveDisk)
	d.abstractDisk.Disk = d
	return d
}

func (d *OneDriveDisk) GetConfig() *oauth2.Config {
	b, err := os.ReadFile("./models/disk/OneDriveDisk/credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, "files.readwrite", "wl.offline_access")
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	return config
}

/* Factory methods */

func NewOneDriveDisk() *OneDriveDisk {
	var d *OneDriveDisk = new(OneDriveDisk)
	d.abstractDisk.Disk = d
	return d
}

func init() {
	models.DiskTypesRegistry[constants.PROVIDER_TYPE_ONEDRIVE] = func() models.Disk { return NewOneDriveDisk() }
}
