package drain_test

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.clever-cloud.com/terraform-provider/pkg/resources/drain"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

func datadogAPIDrain(t *testing.T, url string) tmp.Drain {
	t.Helper()
	recipient, err := json.Marshal(map[string]string{
		"type": "DatadogRecipient",
		"url":  url,
	})
	if err != nil {
		t.Fatalf("marshal recipient: %v", err)
	}
	return tmp.Drain{
		ID:            "drain_id",
		ApplicationID: "app_id",
		Kind:          tmp.DRAIN_KIND_LOG,
		Recipient:     json.RawMessage(recipient),
	}
}

func TestDatadogDrainFromAPI_DecodesPercentEncodedSegments(t *testing.T) {
	// %2D -> '-', %2F -> '/', %3D -> '=': encoded inside the key segment,
	// so the literal "/" split still lands on the endpoint/key boundary.
	apiDrain := datadogAPIDrain(t, "https://http-intake.logs.datadoghq.eu/v1/input/fake%2Dapi%2Dkey%2F123%3D%3D")

	d := &drain.DatadogDrain{}
	if err := d.FromAPI(apiDrain); err != nil {
		t.Fatalf("FromAPI returned error: %v", err)
	}

	if got := d.Endpoint.ValueString(); got != "https://http-intake.logs.datadoghq.eu/v1/input" {
		t.Errorf("Endpoint = %q, want %q", got, "https://http-intake.logs.datadoghq.eu/v1/input")
	}
	if got := d.APIKey.ValueString(); got != "fake-api-key/123==" {
		t.Errorf("APIKey = %q, want %q", got, "fake-api-key/123==")
	}
}

func TestDatadogDrainFromAPI_UnencodedURLUnchanged(t *testing.T) {
	apiDrain := datadogAPIDrain(t, "https://http-intake.logs.datadoghq.com/v1/input/fake-api-key-plain-123")

	d := &drain.DatadogDrain{}
	if err := d.FromAPI(apiDrain); err != nil {
		t.Fatalf("FromAPI returned error: %v", err)
	}

	if got := d.Endpoint.ValueString(); got != "https://http-intake.logs.datadoghq.com/v1/input" {
		t.Errorf("Endpoint = %q, want %q", got, "https://http-intake.logs.datadoghq.com/v1/input")
	}
	if got := d.APIKey.ValueString(); got != "fake-api-key-plain-123" {
		t.Errorf("APIKey = %q, want %q", got, "fake-api-key-plain-123")
	}
}

func TestDatadogDrainFromAPI_PreservesExistingState(t *testing.T) {
	// API returns a different (and differently-encoded) URL, but since the
	// state already has known values, FromAPI must not overwrite them.
	apiDrain := datadogAPIDrain(t, "https://http-intake.logs.datadoghq.eu/v1/input/other%2Dkey")

	d := &drain.DatadogDrain{
		Endpoint: types.StringValue("https://http-intake.logs.datadoghq.com/v1/input"),
		APIKey:   types.StringValue("state-api-key-123"),
	}
	if err := d.FromAPI(apiDrain); err != nil {
		t.Fatalf("FromAPI returned error: %v", err)
	}

	if got := d.Endpoint.ValueString(); got != "https://http-intake.logs.datadoghq.com/v1/input" {
		t.Errorf("Endpoint = %q, want state value preserved, got %q", got, "https://http-intake.logs.datadoghq.com/v1/input")
	}
	if got := d.APIKey.ValueString(); got != "state-api-key-123" {
		t.Errorf("APIKey = %q, want state value preserved, got %q", got, "state-api-key-123")
	}
}

func TestDatadogDrainFromAPI_InvalidEscapeKeepsRawValue(t *testing.T) {
	// "%zz" is not a valid percent-escape; PathUnescape errors and the raw
	// segment must be kept rather than failing the Read.
	apiDrain := datadogAPIDrain(t, "https://http-intake.logs.datadoghq.eu/v1/input/fake%zzkey")

	d := &drain.DatadogDrain{}
	if err := d.FromAPI(apiDrain); err != nil {
		t.Fatalf("FromAPI returned error: %v", err)
	}

	if got := d.APIKey.ValueString(); got != "fake%zzkey" {
		t.Errorf("APIKey = %q, want raw value %q preserved on invalid escape", got, "fake%zzkey")
	}
}
