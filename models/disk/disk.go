package disk

import "golang.org/x/net/context"

type Disk interface {
	Connect() error
	Upload() error
	Download() error
	Rename() error
	Remove() error
}

type AbstractDisk struct {
	Disk
}

func (d *AbstractDisk) Connect(ctx context.Context) error {

	return nil
}
