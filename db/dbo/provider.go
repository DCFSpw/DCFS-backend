package dbo

type Provider struct {
	AbstractDatabaseObject
	Type int    `json:"type"`
	Name string `json:"name"`
	Logo string `json:"logo"`
}

// NewProvider - create new provider object
//
// return type:
//   - *dbo.Provider: created provider DBO
func NewProvider() *Provider {
	var p *Provider = new(Provider)
	p.AbstractDatabaseObject.DatabaseObject = p
	return p
}
