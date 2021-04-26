package history

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"cloud.google.com/go/storage"
	"github.com/googleapis/google-cloud-go-testing/storage/stiface"
	"github.com/mitchellh/go-homedir"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/jwt"
	"google.golang.org/api/option"
)

// GCSStorageConfig is a config for gcs storage.
// This is expected to have the same options as the Terraform gcs backend.
// https://www.terraform.io/docs/backends/types/gcs.html
type GCSStorageConfig struct {
	// The name of the Google Cloud Storage bucket.
	Bucket string `hcl:"bucket"`

	// Prefix is the directory where state files will be saved inside the bucket.
	Prefix string `hcl:"prefix"`

	// Credentials is a local path to Google Cloud Platform account credentials in JSON format.
	// Alternatively, https://cloud.google.com/docs/authentication/production#automatically can be used.
	Credentials string `hcl:"credentials,optional"`

	// AccessToken is a OAuth2 token used for GCP authentication.
	AccessToken string `hcl:"access_token,optional"`

	// ImpersonateServiceAccount is the service account used to impersonate for all Google API Calls.
	ImpersonateServiceAccount string `hcl:"impersonate_service_account,optional"`

	// ImpersonateServiceAccountDelegates is the delegation chain for the impersonated service account.
	ImpersonateServiceAccountDelegates []string `hcl:"impersonate_service_account_delegates,optional"`

	// EncryptionKey is a 32 byte base64 encoded 'customer supplied encryption key' used to encrypt all state.
	EncryptionKey string `hcl:"encryption_key,optional"`
}

// GCSStorageConfig implements a StorageConfig.
var _ StorageConfig = (*GCSStorageConfig)(nil)

// NewStorage returns a new instance of GCSStorage.
func (c *GCSStorageConfig) NewStorage() (Storage, error) {
	return NewGCSStorage(c, nil)
}

var _ Storage = (*GCSStorage)(nil)

// NewGCSStorage returns a new instance of GCSStorage.
func NewGCSStorage(config *GCSStorageConfig, client stiface.Client) (*GCSStorage, error) {
	return &GCSStorage{
		config: config,
		client: client,
	}, nil
}

// GCSStorage is an implementation of Storage for Google GCS.
type GCSStorage struct {
	// config is a storage config for gcs.
	config *GCSStorageConfig

	// Storage can be mocked for testing.
	client stiface.Client

	// encryptionKey is the Customer Encryption Key used to encrypt/decrypt objects
	encryptionKey []byte
}

func (s *GCSStorage) Write(ctx context.Context, b []byte) error {
	if err := s.init(ctx); err != nil {
		return err
	}

	w := s.client.Bucket(s.config.Bucket).
		Object(s.config.Prefix).
		NewWriter(ctx)

	if _, err := w.Write(b); err != nil {
		return fmt.Errorf("failed writing to gcs://%s/%s: %w", s.config.Bucket, s.config.Prefix, err)
	} else if err = w.Close(); err != nil {
		return fmt.Errorf("failed closing write to gcs://%s/%s: %w", s.config.Bucket, s.config.Prefix, err)
	}
	return nil
}

func (s *GCSStorage) Read(ctx context.Context) ([]byte, error) {
	if err := s.init(ctx); err != nil {
		return nil, err
	}

	r, err := s.client.Bucket(s.config.Bucket).
		Object(s.config.Prefix).
		NewReader(ctx)
	if err != nil {
		if err.Error() == "storage: object doesn't exist" {
			// If the key does not exist
			return []byte{}, nil
		}

		// unexpected error
		return nil, fmt.Errorf("failed to start reading gcs://%s/%s: %w", s.config.Bucket, s.config.Prefix, err)
	}

	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed reading gcs://%s/%s: %w", s.config.Bucket, s.config.Prefix, err)
	} else if err = r.Close(); err != nil {
		return nil, fmt.Errorf("failed closing reader for gcs://%s/%s: %w", s.config.Bucket, s.config.Prefix, err)
	}
	return data, nil
}

// adapted from https://github.com/hashicorp/terraform/blob/2635b3b/backend/remote-state/gcs/backend.go#L113-L228
func (s *GCSStorage) init(ctx context.Context) error {
	if s.client != nil {
		return nil
	}
	var opts []option.ClientOption

	var creds string
	var tokenSource oauth2.TokenSource
	if s.config.AccessToken != "" {
		tokenSource = oauth2.StaticTokenSource(&oauth2.Token{
			AccessToken: s.config.AccessToken,
		})
	} else if s.config.Credentials != "" {
		creds = s.config.Credentials
	} else if v := os.Getenv("GOOGLE_BACKEND_CREDENTIALS"); v != "" {
		creds = v
	} else {
		creds = os.Getenv("GOOGLE_CREDENTIALS")
	}

	if tokenSource != nil {
		opts = append(opts, option.WithTokenSource(tokenSource))
	} else if creds != "" {
		var account accountFile

		// to mirror how the provider works, we accept the file path or the contents
		contents, err := readPathOrContents(creds)
		if err != nil {
			return fmt.Errorf("error loading credentials: %s", err)
		}

		if err := json.Unmarshal([]byte(contents), &account); err != nil {
			return fmt.Errorf("error parsing credentials '%s': %s", contents, err)
		}

		conf := jwt.Config{
			Email:      account.ClientEmail,
			PrivateKey: []byte(account.PrivateKey),
			Scopes:     []string{storage.ScopeReadWrite},
			TokenURL:   "https://oauth2.googleapis.com/token",
		}

		opts = append(opts, option.WithHTTPClient(conf.Client(ctx)))
	} else {
		opts = append(opts, option.WithScopes(storage.ScopeReadWrite))
	}

	// Service Account Impersonation
	if svcAccount := s.config.ImpersonateServiceAccount; svcAccount != "" {
		opts = append(opts, option.ImpersonateCredentials(svcAccount))

		if delegates := s.config.ImpersonateServiceAccountDelegates; len(delegates) != 0 {
			opts = append(opts, option.ImpersonateCredentials(svcAccount, delegates...))
		}
	}

	client, err := storage.NewClient(ctx, opts...)
	if err != nil {
		return fmt.Errorf("init failed: %v", err)
	}

	// adapter allows actual GCS client to be used in mocking interface
	s.client = stiface.AdaptClient(client)

	key := s.config.EncryptionKey
	if key == "" {
		key = os.Getenv("GOOGLE_ENCRYPTION_KEY")
	}

	if key != "" {
		kc, err := readPathOrContents(key)
		if err != nil {
			return fmt.Errorf("Error loading encryption key: %s", err)
		}

		// The GCS client expects a customer supplied encryption key to be
		// passed in as a 32 byte long byte slice. The byte slice is base64
		// encoded before being passed to the API. We take a base64 encoded key
		// to remain consistent with the GCS docs.
		// https://cloud.google.com/storage/docs/encryption#customer-supplied
		// https://github.com/GoogleCloudPlatform/google-cloud-go/blob/def681/storage/storage.go#L1181
		k, err := base64.StdEncoding.DecodeString(kc)
		if err != nil {
			return fmt.Errorf("Error decoding encryption key: %s", err)
		}
		s.encryptionKey = k
	}

	return nil
}

// If the argument is a path, Read loads it and returns the contents,
// otherwise the argument is assumed to be the desired contents and is simply
// returned.
// From https://github.com/hashicorp/terraform/blob/7c0ec01/backend/backend.go#L330-L356
func readPathOrContents(poc string) (string, error) {
	if len(poc) == 0 {
		return poc, nil
	}

	path := poc
	if path[0] == '~' {
		var err error
		path, err = homedir.Expand(path)
		if err != nil {
			return path, err
		}
	}

	if _, err := os.Stat(path); err == nil {
		contents, err := ioutil.ReadFile(path)
		if err != nil {
			return string(contents), err
		}
		return string(contents), nil
	}

	return poc, nil
}

// accountFile represents the structure of the account file JSON file.
type accountFile struct {
	PrivateKeyId string `json:"private_key_id"`
	PrivateKey   string `json:"private_key"`
	ClientEmail  string `json:"client_email"`
	ClientId     string `json:"client_id"`
}
