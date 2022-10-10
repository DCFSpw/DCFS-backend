package dbo

type Provider struct {
	AbstractDatabaseObject
	Type int
	Name string
	Logo string
}

func NewProvider() *Provider {
	var p *Provider = new(Provider)
	p.AbstractDatabaseObject.DatabaseObject = p
	return p
}
