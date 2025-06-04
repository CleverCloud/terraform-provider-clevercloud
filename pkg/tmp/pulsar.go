package tmp

import (
	"context"
	"fmt"

	"go.clever-cloud.dev/client"
)

/*
	{
	  "id": "pulsar_",
	  "owner_id": "user_",
	  "tenant": "user_",
	  "namespace": "pulsar_,
	  "cluster_id": "c3",
	  "token": "aaa",
	  "creation_date": "2025-06-05T08:13:14.101962Z",
	  "ask_for_deletion_date": null,
	  "deletion_date": null,
	  "status": "ACTIVE",
	  "plan": "BETA",
	  "cold_storage_id": "cellar_",
	  "cold_storage_linked": true,
	  "cold_storage_must_be_provided": true
	}
*/
type Pulsar struct {
	ID                        string  `json:"id"`
	OwnerID                   string  `json:"owner_id"`
	Tenant                    string  `json:"tenant"`
	Namespace                 string  `json:"namespace"`
	ClusterID                 string  `json:"cluster_id"`
	Token                     string  `json:"token"`
	CreationDate              string  `json:"creation_date"`
	Plan                      string  `json:"plan"`
	AskForDeletionDate        *string `json:"ask_for_deletion_date"`
	DeletionDate              *string `json:"deletion_date"`
	Status                    string  `json:"status"`
	ColdStorageID             string  `json:"cold_storage_id"`
	ColdStorageLinked         bool    `json:"cold_storage_linked"`
	ColdStorageMustBeProvided bool    `json:"cold_storage_must_be_provided"`
}

func GetPulsar(ctx context.Context, cc *client.Client, organisationID, pulsarID string) client.Response[Pulsar] {
	path := fmt.Sprintf("/v4/addon-providers/addon-pulsar/addons/%s", pulsarID)
	return client.Get[Pulsar](ctx, cc, path)
}

/*
*

	{
	  "id": "c3",
	  "url": "materiamq.eu-fr-1.services.clever-cloud.com",
	  "pulsar_port": 6650,
	  "pulsar_tls_port": 6651,
	  "web_port": 80,
	  "web_tls_port": 443,
	  "version": "4.0.3",
	  "available": true,
	  "zone": "PAR",
	  "support_cold_storage": true,
	  "supported_plans": [
	    "BETA",
	    "ORGANISATION_LOGS",
	    "ORGANISATION_ACCESS_LOGS",
	    "ORGANISATION_AUDIT_LOGS",
	    "ORGANISATION_ACTIONS"
	  ]
	}
*/
type PulsarCluster struct {
	ID                 string   `json:"id"`
	URL                string   `json:"url"`
	PulsarPort         int      `json:"pulsar_port"`
	PulsarTLSPort      int      `json:"pulsar_tls_port"`
	WebPort            int      `json:"web_port"`
	WebTLSPort         int      `json:"web_tls_port"`
	Version            string   `json:"version"`
	Available          bool     `json:"available"`
	Zone               string   `json:"zone"`
	SupportColdStorage bool     `json:"support_cold_storage"`
	SupportedPlans     []string `json:"supported_plans"`
}

func GetPulsarCluster(ctx context.Context, cc *client.Client, clusterID string) client.Response[PulsarCluster] {
	path := fmt.Sprintf("/v4/addon-providers/addon-pulsar/clusters/%s", clusterID)
	return client.Get[PulsarCluster](ctx, cc, path)
}
