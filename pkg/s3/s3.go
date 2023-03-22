package s3

import (
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.clever-cloud.com/terraform-provider/pkg/tmp"
)

func minioClientFor(endpoint, id, secret string) (*minio.Client, error) {
	return minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(id, secret, ""),
		Secure: true,
	})
}

type CellarCreds struct {
	Host      string
	KeyID     string
	KeySecret string
}

// Extract S3 credentials from Clever Cloud Cellar exposed env vars
func FromEnvVars(envVars []tmp.EnvVar) *CellarCreds {
	creds := &CellarCreds{}

	for _, envVar := range envVars {
		switch envVar.Name {
		case "CELLAR_ADDON_KEY_SECRET":
			creds.KeySecret = envVar.Value
		case "CELLAR_ADDON_KEY_ID":
			creds.KeyID = envVar.Value
		case "CELLAR_ADDON_HOST":
			creds.Host = envVar.Value
		default:
		}
	}

	return creds
}

func MinioClientFromEnvsFor(envVars []tmp.EnvVar) (*minio.Client, error) {
	creds := FromEnvVars(envVars)
	return minioClientFor(creds.Host, creds.KeyID, creds.KeySecret)
}
