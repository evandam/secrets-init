package google

import (
	"context"
	"secrets-init/pkg/secrets"
	"strings"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"github.com/pkg/errors"
	secretspb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
)

// SecretsProvider Google Cloud secrets provider
type SecretsProvider struct {
	sm GoogleSecretsManagerAPI
}

// NewGoogleSecretsProvider init Google Secrets Provider
func NewGoogleSecretsProvider(ctx context.Context) (secrets.Provider, error) {
	sp := SecretsProvider{}
	var err error
	sp.sm, err = secretmanager.NewClient(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize Google Cloud SDK")
	}
	return &sp, nil
}

// ResolveSecrets replaces all passed variables values prefixed with 'gcp:secretmanager'
// by corresponding secrets from Google Secret Manager
// The secret name should be in the format (optionally with version)
//    `gcp:secretmanager:projects/{PROJECT_ID}/secrets/{SECRET_NAME}`
//    `gcp:secretmanager:projects/{PROJECT_ID}/secrets/{SECRET_NAME}/versions/{VERSION|latest}`
func (sp SecretsProvider) ResolveSecrets(ctx context.Context, vars []string) ([]string, error) {
	var envs []string
	for _, env := range vars {
		kv := strings.Split(env, "=")
		key, value := kv[0], kv[1]
		if strings.HasPrefix(value, "gcp:secretmanager:") {
			// construct valid secret name
			name := strings.TrimPrefix(value, "gcp:secretmanager:")
			// if no version specified add latest
			if !strings.Contains(name, "/versions/") {
				name += "/versions/latest"
			}
			// get secret value
			req := &secretspb.AccessSecretVersionRequest{
				Name: name,
			}
			secret, err := sp.sm.AccessSecretVersion(ctx, req)
			if err != nil {
				return vars, errors.Wrap(err, "failed to get secret from Google Secret Manager")
			}
			env = key + "=" + string(secret.Payload.GetData())
		}
		envs = append(envs, env)
	}

	return envs, nil
}
