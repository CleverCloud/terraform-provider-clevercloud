package kubernetes

import "go.clever-cloud.dev/sdk/models"

// clusterView is a cluster as the fake API of these tests answers it: the
// fields the provider reads, written the way the API writes them.
type clusterView struct {
	CreationDate string                  `json:"creationDate,omitempty"`
	Features     *models.ClusterFeatures `json:"features,omitempty"`
	ID           string                  `json:"id"`
	Name         string                  `json:"name"`
	Status       string                  `json:"status"`
	Tags         []string                `json:"tags"`
	Topology     *apiTopology            `json:"topologyConfig,omitempty"`
	Version      string                  `json:"version"`
	ZoneID       string                  `json:"zoneId,omitempty"`
	Waiting      *clusterWaiting         `json:"waiting,omitempty"`
}

// clusterWaiting tells what a cluster creation is waiting for
type clusterWaiting struct {
	Attempts   int    `json:"attempts"`
	Dependency string `json:"dependency"`
	Since      string `json:"since"`
	Step       string `json:"step,omitempty"`
}
