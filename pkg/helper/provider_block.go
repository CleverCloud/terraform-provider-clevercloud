package helper

import (
	"fmt"

	"go.clever-cloud.com/terraform-provider/pkg"
)

// Provider structur
type Provider struct {
	provider     string
	organisation string
	blocks       []fmt.Stringer
}

// New function type that accepts pointer to Provider
// (~= Signature of option functions)
type ProviderOption func(*Provider)

// Provider constructor:
//   - desc: Build a new Provider and apply specifics ProviderOption functions
//   - args: provider name, ProviderOption function
//   - return: pointer to Provider
func NewProvider(providerName string) *Provider {
	// default values
	const (
		defaultOrganisation = ""
	)

	p := &Provider{
		provider:     providerName,
		organisation: defaultOrganisation,
	}

	return p
}

// Organisation name:
//   - desc: chained function that set Provider.Organisation then return Provider
//   - args: new organisation name
//   - return: pointer to Provider
func (p *Provider) SetOrganisation(orgName string) *Provider {
	p.organisation = orgName
	return p
}

func (p *Provider) Append(blocks ...fmt.Stringer) *Provider {
	p.blocks = blocks
	return p
}

// Provider block
//   - desc: chained function that stringify Provider into a terraform block
//   - args: none
//   - return: string
func (p *Provider) String() string {
	s := `provider "` + p.provider + `" {
	organisation = "` + p.organisation + `"
}
` + pkg.Reduce(p.blocks, "", func(acc string, block fmt.Stringer) string {
		return acc + block.String() + "\n"
	})
	return s
}
