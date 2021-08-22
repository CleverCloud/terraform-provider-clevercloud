package clevercloud

import "github.com/hashicorp/terraform-plugin-framework/types"

type Application struct {
	ID             types.String           `tfsdk:"id"`
	Name           types.String           `tfsdk:"name"`
	Description    types.String           `tfsdk:"description"`
	Type           types.String           `tfsdk:"type"`
	Zone           types.String           `tfsdk:"zone"`
	DeployType     types.String           `tfsdk:"deploy_type"`
	OrganizationID types.String           `tfsdk:"organization_id"`
	Scalability    ApplicationScalability `tfsdk:"scalability"`
	Properties     ApplicationProperties  `tfsdk:"properties"`
	Build          ApplicationBuild       `tfsdk:"build"`
	// Environment        types.List `tfsdk:"environment"`
	// ExposedEnvironment types.List `tfsdk:"exposed_environment"`
	// Dependencies       types.List `tfsdk:"dependencies"`
	// VHosts             types.List `tfsdk:"vhosts"`
	Favorite types.Bool `tfsdk:"favorite"`
	Archived types.Bool `tfsdk:"archived"`
	Tags     types.List `tfsdk:"tags"`
}

type ApplicationScalability struct {
	MinInstances        types.Number `tfsdk:"min_instances"`
	MaxInstances        types.Number `tfsdk:"max_instances"`
	MaxAllowedInstances types.Number `tfsdk:"max_instances"`
	MinFlavor           types.String `tfsdk:"min_flavor"`
	MaxFlavor           types.String `tfsdk:"max_flavor"`
}

type ApplicationProperties struct {
	Homogeneous    types.Bool `tfsdk:"homogeneous"`
	StickySessions types.Bool `tfsdk:"sticky_sessions"`
	CancelOnPush   types.Bool `tfsdk:"cancel_on_push"`
	ForceHTTPS     types.Bool `tfsdk:"force_https"`
}

type ApplicationBuild struct {
	SeparateBuild types.Bool   `tfsdk:"separate_build"`
	BuildFlavor   types.String `tfsdk:"build_flavor"`
}
