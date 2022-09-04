package dbo

const (
	SFTP     int = 0
	GDRIVE   int = 1
	ONEDRIVE int = 2
)

type Provider struct {
	AbstractDatabaseObject
	ProviderType int
}

func NewProvider() *Provider {
	var p *Provider = new(Provider)
	p.AbstractDatabaseObject.DatabaseObject = p
	return p
}
