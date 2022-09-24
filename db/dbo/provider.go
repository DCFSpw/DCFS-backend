package dbo

type Provider struct {
	AbstractDatabaseObject
	ProviderType int
}

func NewProvider() *Provider {
	var p *Provider = new(Provider)
	p.AbstractDatabaseObject.DatabaseObject = p
	return p
}
