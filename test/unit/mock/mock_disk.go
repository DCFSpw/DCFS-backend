package mock

import (
	"context"
	"dcfs/apicalls"
	"dcfs/constants"
	"dcfs/db/dbo"
	"dcfs/models"
	"dcfs/models/credentials"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"os"
	"time"
)

type MockDisk struct {
	UUID            uuid.UUID
	VirtualDiskUUID uuid.UUID
	Volume          *models.Volume
	Name            string
	SpeedFactor     int

	UsedSpace  uint64
	TotalSpace uint64

	CreationTime  time.Time
	DiskReadiness *MockDiskReadiness
}

/* Mandatory Disk interface implementations */

func (d *MockDisk) Upload(blockMetadata *apicalls.BlockMetadata) *apicalls.ErrorWrapper {
	time.Sleep(time.Duration(d.SpeedFactor) * time.Millisecond)

	*blockMetadata.Status = constants.BLOCK_STATUS_TRANSFERRED
	blockMetadata.CompleteCallback(blockMetadata.UUID, blockMetadata.Status)

	return nil
}

func (d *MockDisk) Download(blockMetadata *apicalls.BlockMetadata) *apicalls.ErrorWrapper {
	time.Sleep(time.Duration(d.SpeedFactor) * time.Millisecond)

	*blockMetadata.Status = constants.BLOCK_STATUS_TRANSFERRED
	blockMetadata.CompleteCallback(blockMetadata.UUID, blockMetadata.Status)

	return nil
}

func (d *MockDisk) Remove(blockMetadata *apicalls.BlockMetadata) *apicalls.ErrorWrapper {
	*blockMetadata.Status = constants.BLOCK_STATUS_TRANSFERRED
	blockMetadata.CompleteCallback(blockMetadata.UUID, blockMetadata.Status)

	return nil
}

func (d *MockDisk) SetVolume(volume *models.Volume) {
	d.Volume = volume
}

func (d *MockDisk) GetVolume() *models.Volume {
	return d.Volume
}

func (d *MockDisk) SetUUID(uuid uuid.UUID) {
	d.UUID = uuid
}

func (d *MockDisk) GetUUID() uuid.UUID {
	return d.UUID
}

func (d *MockDisk) SetName(name string) {
	d.Name = name
}

func (d *MockDisk) GetName() string {
	return d.Name
}

func (d *MockDisk) GetCredentials() credentials.Credentials {
	panic("Unimplemented")
}

func (d *MockDisk) SetCredentials(credentials credentials.Credentials) {
	return
}

func (d *MockDisk) CreateCredentials(c string) {
	return
}

func (d *MockDisk) GetProviderUUID() uuid.UUID {
	panic("Unimplemented")
}

func (d *MockDisk) UpdateUsedSpace(change int64) {
	d.UsedSpace = uint64(int64(d.UsedSpace) + change)
}

func (d *MockDisk) SetIsVirtualFlag(isVirtual bool) {
	return
}

func (d *MockDisk) GetIsVirtualFlag() bool {
	return false
}

func (d *MockDisk) SetVirtualDiskUUID(uuid uuid.UUID) {
	d.VirtualDiskUUID = uuid
}

func (d *MockDisk) GetVirtualDiskUUID() uuid.UUID {
	return d.VirtualDiskUUID
}

func (d *MockDisk) SetUsedSpace(usage uint64) {
	d.UsedSpace = usage
}

func (d *MockDisk) GetUsedSpace() uint64 {
	return d.UsedSpace
}

func (d *MockDisk) GetProviderSpace() (uint64, uint64, string) {
	return 0, 0, constants.OPERATION_NOT_SUPPORTED
}

func (d *MockDisk) SetTotalSpace(quota uint64) {
	d.TotalSpace = quota
}

func (d *MockDisk) GetTotalSpace() uint64 {
	return d.TotalSpace
}

func (d *MockDisk) SetCreationTime(creationTime time.Time) {
	d.CreationTime = creationTime
}

func (d *MockDisk) GetCreationTime() time.Time {
	return d.CreationTime
}

func (d *MockDisk) GetDiskDBO(userUUID uuid.UUID, providerUUID uuid.UUID, volumeUUID uuid.UUID) dbo.Disk {
	panic("Unimplemented")
}

func GetDiskDBOs(number int) []dbo.Disk {
	var ret []dbo.Disk = make([]dbo.Disk, 0)

	for i := 0; i < number; i++ {
		providerDbo, creds := GetRandomProviderDBO()

		ret = append(ret, dbo.Disk{
			AbstractDatabaseObject: dbo.AbstractDatabaseObject{
				UUID: uuid.New(),
			},
			UserUUID:     UserUUID,
			VolumeUUID:   VolumeUUID,
			ProviderUUID: providerDbo.UUID,
			Credentials:  creds,
			Name:         fmt.Sprintf("mock disk no. #%d", i),
			UsedSpace:    0,
			TotalSpace:   15 * 1024 * 1024 * 1024,
			FreeSpace:    15 * 1024 * 1024 * 1024,
			CreatedAt:    time.Time{},
			User:         *UserDBO,
			Volume:       *VolumeDBO,
			Provider:     *providerDbo,
		})
	}

	return ret
}

func (d *MockDisk) AssignDisk(disk models.Disk) {
	panic("Unimplemented")
}

func (d *MockDisk) GetReadiness() models.DiskReadiness {
	return d.DiskReadiness
}

func (d *MockDisk) GetResponse(_disk *dbo.Disk, ctx *gin.Context) *models.DiskResponse {
	return nil
}

type MockDiskReadiness struct {
	Readiness bool
}

func (mdr *MockDiskReadiness) IsReady(ctx context.Context) bool {
	return mdr.Readiness
}

func (mdr *MockDiskReadiness) IsReadyForce(ctx context.Context) bool {
	return mdr.Readiness
}

func (mdr *MockDiskReadiness) IsReadyForceNonBlocking(ctx context.Context) bool {
	return mdr.Readiness
}

func NewMockDisk() models.Disk {
	var d *MockDisk = new(MockDisk)

	d.CreationTime = time.Now()
	d.DiskReadiness = new(MockDiskReadiness)
	d.DiskReadiness.Readiness = true
	d.UUID = uuid.New()

	return d
}

func CreateMockDisk(_d dbo.Disk) models.Disk {
	d := NewMockDisk()
	d.SetCreationTime(_d.GetCreationTime())

	return d
}

func GetMockDisks(number int) []*MockDisk {
	var ret []*MockDisk = make([]*MockDisk, 0)

	for i := 0; i < number; i++ {
		ret = append(ret, &MockDisk{
			UUID:   uuid.New(),
			Volume: nil,
			Name:   fmt.Sprintf("mock dummy disk no. #%d", i),
		})
	}

	return ret
}

func GetSpecifiedDisksDBO(number int, provider int) []dbo.Disk {
	var ret []dbo.Disk = make([]dbo.Disk, 0)

	for i := 0; i < number; i++ {
		providerDbo, creds := GetProviderDBO(provider)

		ret = append(ret, dbo.Disk{
			AbstractDatabaseObject: dbo.AbstractDatabaseObject{
				UUID: uuid.New(),
			},
			UserUUID:     UserUUID,
			VolumeUUID:   VolumeUUID,
			ProviderUUID: providerDbo.UUID,
			Credentials:  creds,
			Name:         fmt.Sprintf("mock disk no. #%d", i),
			UsedSpace:    0,
			TotalSpace:   15 * 1024 * 1024 * 1024,
			FreeSpace:    15 * 1024 * 1024 * 1024,
			CreatedAt:    time.Time{},
			User:         *UserDBO,
			Volume:       *VolumeDBO,
			Provider:     *providerDbo,
		})
	}

	return ret
}

func init() {
	// change path so that there are no problems with the credential files
	_ = os.Chdir("../../")
}
