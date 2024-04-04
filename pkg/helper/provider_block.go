package helper

// Provider structur
type Provider struct {
	Provider     string
	Organisation string
}

// New function type that accepts pointer to Provider
// (~= Signature of option functions)
type ProviderOption func(*Provider)

// Provider constructor:
//   - desc: Build a new Provider and apply specifics ProviderOption functions
//   - args: provider name, ProviderOption function
//   - return: pointer to Provider
func NewProvider(provider string, opts ...ProviderOption) *Provider {
	// default values
	const (
		defaultOrganisation = ""
	)

	p := &Provider{
		Provider:     provider,
		Organisation: defaultOrganisation,
	}

	// ProviderOption functions
	for _, opt := range opts {
		opt(p)
	}

	return p
}

// Organisation name:
//   - desc: chained function that set Provider.Organisation then return Provider
//   - args: new organisation name
//   - return: pointer to Provider
func (p *Provider) SetOrganisation(orgName string) *Provider {
	p.Organisation = orgName
	return p
}

// Provider block
//   - desc: chained function that stringify Provider into a terraform block
//   - args: none
//   - return: string
func (p *Provider) String() string {
	s := `provider "` + p.Provider + `" {
	organisation = "` + p.Organisation + `"
}`
	return s
}
