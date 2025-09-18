package tmp

import (
	"context"
	"fmt"
	"iter"
	"net/url"
	"strings"

	"go.clever-cloud.dev/client"
)

type CreateAppRequest struct {
	Name            string `json:"name" example:"SOME_NAME"`
	Deploy          string `json:"deploy" example:"git"`
	Description     string `json:"description" example:"SOME_DESC"`
	InstanceType    string `json:"instanceType" example:"node"`
	InstanceVariant string `json:"instanceVariant" example:"395103fb-d6e2-4fdd-93bc-bc99146f1ea2"`
	InstanceVersion string `json:"instanceVersion" example:"20220330"`
	MinFlavor       string `json:"minFlavor" example:"pico"`
	MaxFlavor       string `json:"maxFlavor" example:"M"`
	SeparateBuild   bool   `json:"separateBuild" example:"true"` // need to be true if BuildFlavor is set
	BuildFlavor     string `json:"buildFlavor" example:"XL"`
	MinInstances    int64  `json:"minInstances" example:"1"`
	MaxInstances    int64  `json:"maxInstances" example:"4"`
	Zone            string `json:"zone" example:"par"`
	CancelOnPush    bool   `json:"cancelOnPush"`
	StickySessions  bool   `json:"stickySessions"`
	ForceHttps      string `json:"forceHttps"`
}

type CreatAppResponse struct {
	ID             string      `json:"id"`
	Name           string      `json:"name"`
	Description    string      `json:"description"`
	Zone           string      `json:"zone"`
	Instance       Instance    `json:"instance"`
	Deployment     Deployment  `json:"deployment"`
	Vhosts         VHosts      `json:"vhosts"`
	CreationDate   int64       `json:"creationDate"`
	LastDeploy     int         `json:"last_deploy"`
	Archived       bool        `json:"archived"`
	StickySessions bool        `json:"stickySessions"`
	Homogeneous    bool        `json:"homogeneous"`
	Favourite      bool        `json:"favourite"`
	CancelOnPush   bool        `json:"cancelOnPush"`
	WebhookURL     string      `json:"webhookUrl"`
	WebhookSecret  string      `json:"webhookSecret"`
	SeparateBuild  bool        `json:"separateBuild"`
	BuildFlavor    BuildFlavor `json:"buildFlavor"`
	OwnerID        string      `json:"ownerId"`
	State          string      `json:"state"`
	CommitID       string      `json:"commitId"`
	Appliance      string      `json:"appliance"`
	Branch         string      `json:"branch"`
	ForceHTTPS     string      `json:"forceHttps"`
	DeployURL      string      `json:"deployUrl"`
}
type Variant struct {
	ID         string `json:"id"`
	Slug       string `json:"slug"`
	Name       string `json:"name"`
	DeployType string `json:"deployType"`
	Logo       string `json:"logo"`
}
type Memory struct {
	Unit      string `json:"unit"`
	Value     int    `json:"value"`
	Formatted string `json:"formatted"`
}
type MinFlavor struct {
	Name            string  `json:"name"`
	Mem             int     `json:"mem"`
	Cpus            int     `json:"cpus"`
	Gpus            int     `json:"gpus"`
	Disk            int     `json:"disk"`
	Price           float64 `json:"price"`
	Available       bool    `json:"available"`
	Microservice    bool    `json:"microservice"`
	MachineLearning bool    `json:"machine_learning"`
	Nice            int     `json:"nice"`
	PriceID         string  `json:"price_id"`
	Memory          Memory  `json:"memory"`
}
type MaxFlavor struct {
	Name            string  `json:"name"`
	Mem             int     `json:"mem"`
	Cpus            int     `json:"cpus"`
	Gpus            int     `json:"gpus"`
	Disk            int     `json:"disk"`
	Price           float64 `json:"price"`
	Available       bool    `json:"available"`
	Microservice    bool    `json:"microservice"`
	MachineLearning bool    `json:"machine_learning"`
	Nice            int     `json:"nice"`
	PriceID         string  `json:"price_id"`
	Memory          Memory  `json:"memory"`
}
type Flavors struct {
	Name            string  `json:"name"`
	Mem             int     `json:"mem"`
	Cpus            int     `json:"cpus"`
	Gpus            int     `json:"gpus"`
	Disk            int     `json:"disk"`
	Price           float64 `json:"price"`
	Available       bool    `json:"available"`
	Microservice    bool    `json:"microservice"`
	MachineLearning bool    `json:"machine_learning"`
	Nice            int     `json:"nice"`
	PriceID         string  `json:"price_id"`
	Memory          Memory  `json:"memory"`
}

type Instance struct {
	Type                string            `json:"type"`
	Version             string            `json:"version"`
	Variant             Variant           `json:"variant"`
	MinInstances        int               `json:"minInstances"`
	MaxInstances        int               `json:"maxInstances"`
	MaxAllowedInstances int               `json:"maxAllowedInstances"`
	MinFlavor           MinFlavor         `json:"minFlavor"`
	MaxFlavor           MaxFlavor         `json:"maxFlavor"`
	Flavors             []Flavors         `json:"flavors"`
	DefaultEnv          map[string]string `json:"defaultEnv"`
	Lifetime            string            `json:"lifetime"`
	InstanceAndVersion  string            `json:"instanceAndVersion"`
}
type Deployment struct {
	Shutdownable bool   `json:"shutdownable"`
	Type         string `json:"type"`
	RepoState    string `json:"repoState"`
	URL          string `json:"url"`
	HTTPURL      string `json:"httpUrl"`
}
type VHosts []VHost
type VHost struct {
	Fqdn string `json:"fqdn"`
}

func (vhosts VHosts) First() *VHost {
	if len(vhosts) == 0 {
		return nil
	}
	return &vhosts[0]
}

func (vhosts VHosts) AsString() []string {
	result := make([]string, len(vhosts))
	for i, vhost := range vhosts {
		result[i] = vhost.Fqdn
	}
	return result
}

func (vhosts VHosts) AllAsString() iter.Seq[string] {
	return func(yield func(string) bool) {
		for _, vhost := range vhosts {
			if !yield(vhost.Fqdn) {
				return
			}
		}
	}
}

// remove default domain (cleverapps one)
// Ex: app-7a1f2c81-bb18-4682-95fc-b7187a056150.cleverapps.io
func (vhosts VHosts) WithoutCleverApps(appId string) VHosts {
	cleverapps := fmt.Sprintf("%s.cleverapps.io/", strings.ReplaceAll(appId, "_", "-"))
	result := []VHost{}

	for _, vhost := range vhosts {
		if vhost.Fqdn == cleverapps {
			continue
		}

		result = append(result, vhost)
	}

	return result
}

func (vhosts VHosts) CleverAppsFQDN(appId string) *VHost {
	cleverapps := fmt.Sprintf("%s.cleverapps.io/", strings.ReplaceAll(appId, "_", "-"))

	for _, vhost := range vhosts {
		if vhost.Fqdn == cleverapps {
			return &vhost
		}
	}

	return nil
}

type BuildFlavor struct {
	Name            string  `json:"name"`
	Mem             int     `json:"mem"`
	Cpus            int     `json:"cpus"`
	Gpus            int     `json:"gpus"`
	Disk            any     `json:"disk"`
	Price           float64 `json:"price"`
	Available       bool    `json:"available"`
	Microservice    bool    `json:"microservice"`
	MachineLearning bool    `json:"machine_learning"`
	Nice            int     `json:"nice"`
	PriceID         string  `json:"price_id"`
	Memory          Memory  `json:"memory"`
}

type Env struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func CreateApp(ctx context.Context, cc *client.Client, organisationID string, app CreateAppRequest) client.Response[CreatAppResponse] {
	path := fmt.Sprintf("/v2/organisations/%s/applications", organisationID)
	return client.Post[CreatAppResponse](ctx, cc, path, app)
}

func GetApp(ctx context.Context, cc *client.Client, organisationID, applicationID string) client.Response[CreatAppResponse] {
	path := fmt.Sprintf("/v2/organisations/%s/applications/%s", organisationID, applicationID)
	return client.Get[CreatAppResponse](ctx, cc, path)
}

func DeleteApp(ctx context.Context, cc *client.Client, organisationID, applicationID string) client.Response[any] {
	path := fmt.Sprintf("/v2/organisations/%s/applications/%s", organisationID, applicationID)
	return client.Delete[any](ctx, cc, path)
}

func GetAppEnv(ctx context.Context, cc *client.Client, organisationID string, applicationID string) client.Response[[]Env] {
	path := fmt.Sprintf("/v2/organisations/%s/applications/%s/env", organisationID, applicationID)
	return client.Get[[]Env](ctx, cc, path)
}

func UpdateAppEnv(ctx context.Context, cc *client.Client, organisationID string, applicationID string, envs map[string]string) client.Response[any] {
	path := fmt.Sprintf("/v2/organisations/%s/applications/%s/env", organisationID, applicationID)
	return client.Put[any](ctx, cc, path, envs)
}

type ProductInstance struct {
	Type          string        `json:"type"`
	Version       string        `json:"version"`
	Name          string        `json:"name"`
	Variant       Variant       `json:"variant"`
	Description   string        `json:"description"`
	Enabled       bool          `json:"enabled"`
	ComingSoon    bool          `json:"comingSoon"`
	MaxInstances  int           `json:"maxInstances"`
	Tags          []string      `json:"tags"`
	Deployments   []string      `json:"deployments"`
	Flavors       []Flavors     `json:"flavors"`
	DefaultFlavor DefaultFlavor `json:"defaultFlavor"`
	BuildFlavor   BuildFlavor   `json:"buildFlavor"`
}

type DefaultFlavor struct {
	Name            string  `json:"name"`
	Mem             int     `json:"mem"`
	Cpus            int     `json:"cpus"`
	Gpus            int     `json:"gpus"`
	Disk            any     `json:"disk"`
	Price           float64 `json:"price"`
	Available       bool    `json:"available"`
	Microservice    bool    `json:"microservice"`
	MachineLearning bool    `json:"machine_learning"`
	Nice            int     `json:"nice"`
	PriceID         string  `json:"price_id"`
	Memory          Memory  `json:"memory"`
}

func GetProductInstance(ctx context.Context, cc *client.Client, ownerId *string) client.Response[[]ProductInstance] {
	path := "/v2/products/instances"
	if ownerId != nil {
		path = fmt.Sprintf("%s?for=%s", path, *ownerId)
	}
	return client.Get[[]ProductInstance](ctx, cc, path)
}

type UpdateAppReq struct {
	Name            string `json:"name" example:"SOME_NAME"`
	Deploy          string `json:"deploy" example:"git"`
	Description     string `json:"description" example:"SOME_DESC"`
	InstanceType    string `json:"instanceType" example:"node"`
	InstanceVariant string `json:"instanceVariant" example:"395103fb-d6e2-4fdd-93bc-bc99146f1ea2"`
	InstanceVersion string `json:"instanceVersion" example:"20220330"`
	MinFlavor       string `json:"minFlavor" example:"pico"`
	MaxFlavor       string `json:"maxFlavor" example:"M"`
	BuildFlavor     string `json:"buildFlavor" example:"XL"`
	MinInstances    int64  `json:"minInstances" example:"1"`
	MaxInstances    int64  `json:"maxInstances" example:"4"`
	Zone            string `json:"zone" example:"par"`
	CancelOnPush    bool   `json:"cancelOnPush"`
	StickySessions  bool   `json:"stickySessions"`
	ForceHttps      string `json:"forceHttps"`
}

func UpdateApp(ctx context.Context, cc *client.Client, organisationID, applicationID string, req UpdateAppReq) client.Response[CreatAppResponse] {
	path := fmt.Sprintf("/v2/organisations/%s/applications/%s", organisationID, applicationID)
	return client.Put[CreatAppResponse](ctx, cc, path, req)
}

func GetAppVhosts(ctx context.Context, cc *client.Client, organisationID, applicationID string) client.Response[VHosts] {
	path := fmt.Sprintf("/v2/organisations/%s/applications/%s/vhosts", organisationID, applicationID)
	return client.Get[VHosts](ctx, cc, path)
}

func AddAppVHost(ctx context.Context, cc *client.Client, organisationID, applicationID, vhost string) client.Response[any] {
	vhost = url.QueryEscape(vhost)
	path := fmt.Sprintf("/v2/organisations/%s/applications/%s/vhosts/%s", organisationID, applicationID, vhost)
	return client.Put[any](ctx, cc, path, map[string]string{})
}

func DeleteAppVHost(ctx context.Context, cc *client.Client, organisationID, applicationID, vhost string) client.Response[any] {
	vhost = url.QueryEscape(vhost)
	path := fmt.Sprintf("/v2/organisations/%s/applications/%s/vhosts/%s", organisationID, applicationID, vhost)
	return client.Delete[any](ctx, cc, path)
}

func AddAppLinkedAddons(ctx context.Context, cc *client.Client, organisationID, applicationID, addonID string) client.Response[client.Nothing] {
	path := fmt.Sprintf("/v2/organisations/%s/applications/%s/addons", organisationID, applicationID)
	return client.Post[client.Nothing](ctx, cc, path, addonID)
}

func GetAppLinkedAddons(ctx context.Context, cc *client.Client, organisationID, applicationID string) client.Response[[]AddonResponse] {
	path := fmt.Sprintf("/v2/organisations/%s/applications/%s/addons", organisationID, applicationID)
	return client.Get[[]AddonResponse](ctx, cc, path)
}

type RestartAppRes struct {
	ID           int    `json:"id"`
	Message      string `json:"message"`
	Type         string `json:"type"` // error / success
	DeploymentID string `json:"deploymentId"`
}

func RestartApp(ctx context.Context, cc *client.Client, organisationID, applicationID string) client.Response[RestartAppRes] {
	path := fmt.Sprintf("/v2/organisations/%s/applications/%s/instances", organisationID, applicationID)
	return client.Post[RestartAppRes](ctx, cc, path, nil)
}
