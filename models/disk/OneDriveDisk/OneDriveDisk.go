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
	"github.com/gin-gonic/gin"
	"github.com/goh-chunlin/go-onedrive/onedrive"
	"github.com/google/uuid"
	"github.com/h2non/filetype"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"io"
	"log"
	"math"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"time"
)

type OneDriveDisk struct {
	abstractDisk AbstractDisk.AbstractDisk
}

/* Mandatory Disk interface implementations */

func (d *OneDriveDisk) Upload(blockMetadata *apicalls.BlockMetadata) *apicalls.ErrorWrapper {
	var _client interface{} = d.GetCredentials().Authenticate(&apicalls.CredentialsAuthenticateMetadata{Ctx: blockMetadata.Ctx, Config: d.GetConfig(), DiskUUID: d.GetUUID()})
	if _client == nil {
		return apicalls.CreateErrorWrapper(constants.REMOTE_CANNOT_AUTHENTICATE, "could not connect to the remote server")
	}

	var client *http.Client = _client.(*http.Client)
	oneDriveClient := onedrive.NewClient(client)

	// size in bytes
	var size int = len(*blockMetadata.Content)
	var apiURL string = "me/drive/root:/" + url.PathEscape(blockMetadata.UUID.String())

	ft, err := filetype.Match(*blockMetadata.Content)
	if err != nil {
		log.Printf("[OneDrive upload] file %s is corrupted", blockMetadata.FileUUID.String())
	}

	if size <= constants.ONEDRIVE_SIZE_LIMIT {
		// fast upload
		req, err := oneDriveClient.NewFileUploadRequest(apiURL+":/content?@microsoft.graph.conflictBehavior=rename", ft.MIME.Value, bytes.NewReader(*blockMetadata.Content))
		if err != nil {
			return apicalls.CreateErrorWrapper(constants.REMOTE_BAD_REQUEST, "Could not create a file upload request:", err.Error())
		}

		var response *onedrive.DriveItem
		err = oneDriveClient.Do(blockMetadata.Ctx, req, false, &response)
		if err != nil {
			return apicalls.CreateErrorWrapper(constants.REMOTE_FAILED_JOB, "Could not send file:", err.Error())
		}
	} else {
		// upload session
		url, err := oneDriveClient.BaseURL.Parse(apiURL + ":/createUploadSession")
		err = nil

		req, err := http.NewRequest("POST", url.String(), nil)
		req.Header.Set("Content-Type", "application/json")
		if err != nil {
			return apicalls.CreateErrorWrapper(constants.REMOTE_BAD_REQUEST, "Could not create a file upload session:", err.Error())
		}

		var response *responses.CreateUploadSessionResponse
		err = oneDriveClient.Do(blockMetadata.Ctx, req, false, &response)
		if err != nil {
			return apicalls.CreateErrorWrapper(constants.REMOTE_FAILED_JOB, "Could not send file:", err.Error())
		}

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
				return apicalls.CreateErrorWrapper(constants.REMOTE_BAD_REQUEST, "Could not create a file upload request:", err.Error())
			}

			if upperBound < size {
				rsp := responses.UploadSessionResponse{}
				err = oneDriveClient.Do(blockMetadata.Ctx, request, false, &rsp)
				if err != nil {
					return apicalls.CreateErrorWrapper(constants.REMOTE_FAILED_JOB, "Could not send file:", err.Error())
				}
			} else {
				rsp := responses.UploadSessionFinalResponse{}
				err = oneDriveClient.Do(blockMetadata.Ctx, request, false, &rsp)
				if err != nil {
					return apicalls.CreateErrorWrapper(constants.REMOTE_FAILED_JOB, "Could not send file:", err.Error())
				}
			}
		}
	}

	blockMetadata.CompleteCallback(blockMetadata.FileUUID, blockMetadata.Status)
	return nil
}

type oneDriveSearchDriveItem struct {
	DataType string `json:"@odata.type"`
	Id       string `json:"id"`
}

type oneDriveSearchResponse struct {
	Context string                    `json:"@odata.context"`
	Value   []oneDriveSearchDriveItem `json:"value"`
}

func (d *OneDriveDisk) Download(blockMetadata *apicalls.BlockMetadata) *apicalls.ErrorWrapper {
	var _client interface{} = d.GetCredentials().Authenticate(&apicalls.CredentialsAuthenticateMetadata{Ctx: blockMetadata.Ctx, Config: d.GetConfig(), DiskUUID: d.GetUUID()})
	if _client == nil {
		return apicalls.CreateErrorWrapper(constants.REMOTE_CANNOT_AUTHENTICATE, "could not connect to the remote server")
	}

	var client *http.Client = _client.(*http.Client)
	oneDriveClient := onedrive.NewClient(client)
	var searchReqUrl string = "me/drive/root/search(q='" + url.PathEscape(blockMetadata.UUID.String()) + "')?select=id"

	req, err := oneDriveClient.NewRequest("GET", searchReqUrl, nil)
	if err != nil {
		return apicalls.CreateErrorWrapper(constants.REMOTE_BAD_REQUEST, "Could not create a file search request:", err.Error())
	}

	var response oneDriveSearchResponse = oneDriveSearchResponse{}
	err = oneDriveClient.Do(blockMetadata.Ctx, req, false, &response)
	if err != nil {
		return apicalls.CreateErrorWrapper(constants.REMOTE_FAILED_JOB, "Could not find file:", err.Error())
	}

	if len(response.Value) > 1 {
		return apicalls.CreateErrorWrapper(constants.REMOTE_FAILED_JOB, "Block hierarchy corrupted")
	}

	if len(response.Value) == 0 {
		return apicalls.CreateErrorWrapper(constants.REMOTE_FAILED_JOB, "Could not find file")
	}

	item, err := oneDriveClient.DriveItems.Get(blockMetadata.Ctx, response.Value[0].Id)
	if err != nil {
		return apicalls.CreateErrorWrapper(constants.REMOTE_FAILED_JOB, "Could not download file:", err.Error())
	}

	downloadReq, err := oneDriveClient.NewRequest("GET", item.DownloadURL, nil)
	if err != nil {
		return apicalls.CreateErrorWrapper(constants.REMOTE_BAD_REQUEST, "Could not create a file download request:", err.Error())
	}

	var downloadRsp *http.Response
	downloadRsp, err = client.Do(downloadReq.WithContext(blockMetadata.Ctx))
	if err != nil {
		return apicalls.CreateErrorWrapper(constants.REMOTE_FAILED_JOB, "Could not download file:", err.Error())
	}
	defer func() { _ = downloadRsp.Body.Close() }()
	buf := bytes.NewBuffer(nil)

	n, err := io.Copy(buf, downloadRsp.Body)
	if err != nil {
		log.Printf("download failed: %s", err.Error())
		return apicalls.CreateErrorWrapper(constants.REMOTE_FAILED_JOB, "download failed:", err.Error())
	}

	if n < blockMetadata.Size {
		log.Printf("downloaded not enough bytes: %d", n)
		return apicalls.CreateErrorWrapper(constants.REMOTE_FAILED_JOB, "downloaded not enough bytes:", fmt.Sprint(n), "out of:", strconv.FormatInt(blockMetadata.Size, 10))
	}

	block := buf.Bytes()[0:blockMetadata.Size]
	blockMetadata.Content = &block
	blockMetadata.CompleteCallback(blockMetadata.FileUUID, blockMetadata.Status)
	return nil
}

func (d *OneDriveDisk) Remove(blockMetadata *apicalls.BlockMetadata) *apicalls.ErrorWrapper {
	var _client interface{} = d.GetCredentials().Authenticate(&apicalls.CredentialsAuthenticateMetadata{Ctx: blockMetadata.Ctx, Config: d.GetConfig(), DiskUUID: d.GetUUID()})
	if _client == nil {
		return apicalls.CreateErrorWrapper(constants.REMOTE_CANNOT_AUTHENTICATE, "could not connect to the remote server")
	}

	var client *http.Client = _client.(*http.Client)
	oneDriveClient := onedrive.NewClient(client)
	var searchReqUrl string = "me/drive/root/search(q='" + url.PathEscape(blockMetadata.UUID.String()) + "')?select=id"

	req, err := oneDriveClient.NewRequest("GET", searchReqUrl, nil)
	if err != nil {
		return apicalls.CreateErrorWrapper(constants.REMOTE_BAD_REQUEST, "Could not create a file search request:", err.Error())
	}

	var response oneDriveSearchResponse = oneDriveSearchResponse{}
	err = oneDriveClient.Do(blockMetadata.Ctx, req, false, &response)
	if err != nil {
		return apicalls.CreateErrorWrapper(constants.REMOTE_FAILED_JOB, "Could not find file:", err.Error())
	}

	if len(response.Value) > 1 {
		return apicalls.CreateErrorWrapper(constants.REMOTE_FAILED_JOB, "Block hierarchy corrupted")
	}

	if len(response.Value) == 0 {
		return apicalls.CreateErrorWrapper(constants.REMOTE_FAILED_JOB, "Could not find file")
	}

	err = oneDriveClient.DriveItems.Delete(blockMetadata.Ctx, "", response.Value[0].Id)
	if err != nil {
		return apicalls.CreateErrorWrapper(constants.REMOTE_FAILED_JOB, "Could not remove file:", err.Error())
	}

	blockMetadata.CompleteCallback(blockMetadata.FileUUID, blockMetadata.Status)
	return nil
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

func (d *OneDriveDisk) SetName(name string) {
	d.abstractDisk.SetName(name)
}

func (d *OneDriveDisk) GetName() string {
	return d.abstractDisk.GetName()
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

func (d *OneDriveDisk) SetCreationTime(creationTime time.Time) {
	d.abstractDisk.SetCreationTime(creationTime)
}

func (d *OneDriveDisk) GetCreationTime() time.Time {
	return d.abstractDisk.GetCreationTime()
}

func (d *OneDriveDisk) GetProviderUUID() uuid.UUID {
	return d.abstractDisk.GetProvider(constants.PROVIDER_TYPE_ONEDRIVE)
}

func (d *OneDriveDisk) GetProviderSpace() (uint64, uint64, string) {
	var err error

	// Prepare test context
	writer := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(writer)

	// Authenticate to the remote server
	var _client interface{} = d.GetCredentials().Authenticate(&apicalls.CredentialsAuthenticateMetadata{Ctx: ctx, Config: d.GetConfig(), DiskUUID: d.GetUUID()})
	if _client == nil {
		return 0, 0, constants.REMOTE_CANNOT_AUTHENTICATE
	}

	// Connect to the remote server
	var client *http.Client = _client.(*http.Client)
	oneDriveClient := onedrive.NewClient(client)

	// Get the disk stats from the remote server
	data, err := oneDriveClient.Drives.List(ctx)

	if err != nil || len(data.Drives) == 0 {
		return 0, 0, constants.REMOTE_CANNOT_GET_STATS
	}

	return uint64(data.Drives[0].Quota.Used), uint64(data.Drives[0].Quota.Total), constants.SUCCESS
}

func (d *OneDriveDisk) SetTotalSpace(quota uint64) {
	d.abstractDisk.SetTotalSpace(quota)
}

func (d *OneDriveDisk) GetTotalSpace() uint64 {
	return d.abstractDisk.GetTotalSpace()
}

func (d *OneDriveDisk) SetUsedSpace(usage uint64) {
	d.abstractDisk.SetUsedSpace(usage)
}

func (d *OneDriveDisk) GetUsedSpace() uint64 {
	return d.abstractDisk.GetUsedSpace()
}

func (d *OneDriveDisk) UpdateUsedSpace(change int64) {
	d.abstractDisk.UpdateUsedSpace(change)
}

func (d *OneDriveDisk) GetDiskDBO(userUUID uuid.UUID, providerUUID uuid.UUID, volumeUUID uuid.UUID) dbo.Disk {
	return d.abstractDisk.GetDiskDBO(userUUID, providerUUID, volumeUUID)
}

/* Mandatory OAuthDisk interface implementations */
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
