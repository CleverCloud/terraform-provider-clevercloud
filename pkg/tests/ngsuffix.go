package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"testing"
)

var (
	ngSuffixOnce sync.Once
	ngSuffix     string
	ngSuffixErr  error
)

// NGMemberFQDN builds a networkgroup member FQDN valid on the target platform:
// label + "." + the first DNS suffix served on the public
// /networkgroup/configuration route ("cc-ng.cloud" on the public platform), so
// acceptance tests need no hardcoded suffix and keep working on deployments
// with another one.
func NGMemberFQDN(t *testing.T, label string) string {
	t.Helper()

	ngSuffixOnce.Do(func() {
		endpoint := os.Getenv("CC_API_ENDPOINT")
		if endpoint == "" {
			endpoint = "https://api.clever-cloud.com"
		}

		res, err := http.Get(endpoint + "/networkgroup/configuration")
		if err != nil {
			ngSuffixErr = err
			return
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusOK {
			ngSuffixErr = fmt.Errorf("GET /networkgroup/configuration: status %d", res.StatusCode)
			return
		}

		configuration := struct {
			DNSSuffixes []string `json:"dnsSuffixes"`
		}{}
		if err := json.NewDecoder(res.Body).Decode(&configuration); err != nil {
			ngSuffixErr = err
			return
		}
		if len(configuration.DNSSuffixes) == 0 {
			ngSuffixErr = fmt.Errorf("GET /networkgroup/configuration: empty dnsSuffixes")
			return
		}
		ngSuffix = configuration.DNSSuffixes[0]
	})

	if ngSuffixErr != nil {
		t.Fatalf("cannot resolve the networkgroup DNS suffix: %s", ngSuffixErr)
	}
	return label + "." + ngSuffix
}
