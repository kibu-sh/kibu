package config

import (
	"cloud.google.com/go/secretmanager/apiv1"
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
	"path/filepath"
	"strings"

	"context"
)

var _ Store = (*GCPSecretManagerStore)(nil)

type GCPSecretManagerStore struct {
	projectID string
	client    *secretmanager.Client
}

// GetByKey is a convenience method for getting a value by key
// A simpler alias interface to Get
func (c GCPSecretManagerStore) GetByKey(ctx context.Context, key string, target any) (*CipherText, error) {
	return c.Get(ctx, GetParams{
		Result: target,
		Path:   key,
	})
}

func (c GCPSecretManagerStore) basePath() string {
	return filepath.Join("projects", c.projectID)
}

// secretPath encodes the secret name into a path that is compatible with GCP Secret Manager
func (c GCPSecretManagerStore) secretPath(name string) string {
	return filepath.Join(c.basePath(), "secrets", c.encodeName(name))
}

// encodeName encodes the secret name into a name that is compatible with GCP Secret Manager
// must match format [[a-zA-Z_0-9]+]
func (c GCPSecretManagerStore) encodeName(name string) string {
	name = strings.Replace(name, ".", "-", -1)
	name = strings.Replace(name, "_", "-", -1)
	name = strings.Replace(name, "/", "-", -1)
	name = strings.Replace(name, "-enc-json", "", -1)
	return name
}

func (c GCPSecretManagerStore) secretVersionPath(name, version string) string {
	return filepath.Join(c.secretPath(name), "versions", version)
}

func (c GCPSecretManagerStore) secretVersionPathLatest(name string) string {
	return c.secretVersionPath(name, "latest")
}

func (c GCPSecretManagerStore) Get(ctx context.Context, params GetParams) (ciphertext *CipherText, err error) {
	secret, err := c.client.AccessSecretVersion(ctx, &secretmanagerpb.AccessSecretVersionRequest{
		Name: c.secretVersionPathLatest(params.Path),
	})
	if err != nil {
		return
	}

	err = json.Unmarshal(secret.Payload.GetData(), params.Result)
	if err != nil {
		return
	}

	ciphertext = &CipherText{
		Data: string(secret.Payload.GetData()),
	}

	return
}

func (c GCPSecretManagerStore) Set(ctx context.Context, params SetParams) (ciphertext *CipherText, err error) {
	data, err := json.MarshalIndent(params.Data, "", "  ")
	if err != nil {
		return
	}

	secret, err := c.client.GetSecret(ctx, &secretmanagerpb.GetSecretRequest{
		Name: c.secretPath(params.Path),
	})

	// create secret if	it doesn't exist
	if err != nil {
		secret, err = c.client.CreateSecret(ctx, &secretmanagerpb.CreateSecretRequest{
			Parent:   c.basePath(),
			SecretId: c.encodeName(params.Path),
			Secret: &secretmanagerpb.Secret{
				Replication: &secretmanagerpb.Replication{
					Replication: &secretmanagerpb.Replication_Automatic_{
						Automatic: &secretmanagerpb.Replication_Automatic{},
					},
				},
			},
		})
	}

	if err != nil {
		return
	}

	secretVersion, err := c.client.AddSecretVersion(ctx, &secretmanagerpb.AddSecretVersionRequest{
		Parent: secret.Name,
		Payload: &secretmanagerpb.SecretPayload{
			Data: data,
		},
	})
	if err != nil {
		return
	}

	ciphertext = &CipherText{
		EncryptionKey:  EncryptionKey{},
		CreatedAt:      lo.ToPtr(secretVersion.CreateTime.AsTime()),
		LastModifiedAt: lo.ToPtr(secretVersion.CreateTime.AsTime()),
	}

	return
}

func NewGCPSecretManagerStore(ctx context.Context, projectID string) (store *GCPSecretManagerStore, err error) {
	if projectID == "" {
		err = errors.New("projectID is required")
		return
	}
	client, err := secretmanager.NewClient(ctx)
	store = &GCPSecretManagerStore{
		projectID: projectID,
		client:    client,
	}
	return
}
