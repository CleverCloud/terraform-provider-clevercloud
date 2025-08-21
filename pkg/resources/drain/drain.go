package drain

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"go.clever-cloud.dev/client"
)

type ResourceDrain[T DrainAttributes] struct {
	drainKind string
	t         T // not instantiated, used only for static methods

	cc  *client.Client
	org string
}

func NewDatadogDrain() resource.Resource {
	return &ResourceDrain[*DatadogDrain]{drainKind: "datadog", t: &DatadogDrain{}}
}

func NewNewRelicDrain() resource.Resource {
	return &ResourceDrain[*NewRelicDrain]{drainKind: "newrelic", t: &NewRelicDrain{}}
}

func NewElasticsearchDrain() resource.Resource {
	return &ResourceDrain[*ElasticsearchDrain]{drainKind: "elasticsearch", t: &ElasticsearchDrain{}}
}

func NewSyslogUDPDrain() resource.Resource {
	return &ResourceDrain[*SyslogUDPDrain]{drainKind: "syslog_udp", t: &SyslogUDPDrain{}}
}

func NewSyslogTCPDrain() resource.Resource {
	return &ResourceDrain[*SyslogTCPDrain]{drainKind: "syslog_tcp", t: &SyslogTCPDrain{}}
}

func NewHTTPDrain() resource.Resource {
	return &ResourceDrain[*HTTPDrain]{drainKind: "http", t: &HTTPDrain{}}
}

func NewOVHDrain() resource.Resource {
	return &ResourceDrain[*OVHDrain]{drainKind: "ovh", t: &OVHDrain{}}
}

func (r *ResourceDrain[T]) Metadata(ctx context.Context, req resource.MetadataRequest, res *resource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_drain_" + r.drainKind
}
