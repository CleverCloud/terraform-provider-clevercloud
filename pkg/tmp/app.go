package tmp

import (
	"context"
	"fmt"

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
	MinInstances    int64  `json:"minInstances" example:"1"`
	MaxInstances    int64  `json:"maxInstances" example:"4"`
	Zone            string `json:"zone" example:"par"`
}

type CreatAppResponse struct {
	ID             string      `json:"id"`
	Name           string      `json:"name"`
	Description    string      `json:"description"`
	Zone           string      `json:"zone"`
	Instance       Instance    `json:"instance"`
	Deployment     Deployment  `json:"deployment"`
	Vhosts         []Vhosts    `json:"vhosts"`
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
type Vhosts struct {
	Fqdn string `json:"fqdn"`
}
type BuildFlavor struct {
	Name            string      `json:"name"`
	Mem             int         `json:"mem"`
	Cpus            int         `json:"cpus"`
	Gpus            int         `json:"gpus"`
	Disk            interface{} `json:"disk"`
	Price           float64     `json:"price"`
	Available       bool        `json:"available"`
	Microservice    bool        `json:"microservice"`
	MachineLearning bool        `json:"machine_learning"`
	Nice            int         `json:"nice"`
	PriceID         string      `json:"price_id"`
	Memory          Memory      `json:"memory"`
}

func CreateApp(ctx context.Context, cc *client.Client, organisationID string, app CreateAppRequest) client.Response[CreatAppResponse] {
	path := fmt.Sprintf("/v2/organisations/%s/applications", organisationID)
	return client.Post[CreatAppResponse](ctx, cc, path, app)
}

func GetApp(ctx context.Context, cc *client.Client, organisationID, applicationID string) client.Response[CreatAppResponse] {
	path := fmt.Sprintf("/v2/organisations/%s/applications/%s", organisationID, applicationID)
	return client.Get[CreatAppResponse](ctx, cc, path)
}

func DeleteApp(ctx context.Context, cc *client.Client, organisationID, applicationID string) client.Response[interface{}] {
	path := fmt.Sprintf("/v2/organisations/%s/applications/%s", organisationID, applicationID)
	return client.Delete[interface{}](ctx, cc, path)
}

type Env struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func ListAppEnv(ctx context.Context, cc *client.Client, organisationID string, applicationID string) client.Response[[]Env] {
	path := fmt.Sprintf("/v2/organisations/%s/applications/%s/env", organisationID, applicationID)
	return client.Get[[]Env](ctx, cc, path)
}

func UpdateAppEnv(ctx context.Context, cc *client.Client, organisationID string, applicationID string, envs map[string]string) client.Response[CreatAppResponse] {
	path := fmt.Sprintf("/v2/organisations/%s/applications/%s/env", organisationID, applicationID)
	return client.Put[CreatAppResponse](ctx, cc, path, envs)
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
	Name            string      `json:"name"`
	Mem             int         `json:"mem"`
	Cpus            int         `json:"cpus"`
	Gpus            int         `json:"gpus"`
	Disk            interface{} `json:"disk"`
	Price           float64     `json:"price"`
	Available       bool        `json:"available"`
	Microservice    bool        `json:"microservice"`
	MachineLearning bool        `json:"machine_learning"`
	Nice            int         `json:"nice"`
	PriceID         string      `json:"price_id"`
	Memory          Memory      `json:"memory"`
}

func GetProductInstance(ctx context.Context, cc *client.Client) client.Response[[]ProductInstance] {
	path := "/v2/products/instances"
	return client.Get[[]ProductInstance](ctx, cc, path)
}

func ListDependencies(ctx context.Context, cc *client.Client, org, app string) client.Response[[]AddonResponse] {
	path := fmt.Sprintf("/v2/organisations/%s/applications/%s/addons", org, app)
	return client.Get[[]AddonResponse](ctx, cc, path)
}

func CreateDependency(ctx context.Context, cc *client.Client, org, app, service string) client.Response[client.Nothing] {
	path := fmt.Sprintf("/v2/organisations/%s/applications/%s/addons", org, app)
	return client.Post[client.Nothing](ctx, cc, path, service)
}

func DeleteDependency(ctx context.Context, cc *client.Client, org, app, service string) client.Response[interface{}] {
	path := fmt.Sprintf("/v2/organisations/%s/applications/%s/addons/%s", org, app, service)
	return client.Delete[interface{}](ctx, cc, path)
}
