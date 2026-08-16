package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	ycsdk "github.com/yandex-cloud/go-sdk"
)

func TestCredentialsFromKeyFileUsesWorkloadIdentity(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("unexpected request method: %s", r.Method)
		}
		if r.URL.Path != "/computeMetadata/v1/instance/service-accounts/default/token" {
			t.Errorf("unexpected request path: %s", r.URL.Path)
		}
		if r.Header.Get("Metadata-Flavor") != "Google" {
			t.Errorf("unexpected Metadata-Flavor header: %s", r.Header.Get("Metadata-Flavor"))
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"workload-token","expires_in":3600,"token_type":"Bearer"}`))
	}))
	t.Cleanup(server.Close)
	t.Setenv(ycsdk.InstanceMetadataOverrideEnvVar, server.Listener.Addr().String())

	credentials, err := credentialsFromKeyFile("")
	require.NoError(t, err)

	workloadCredentials, ok := credentials.(ycsdk.NonExchangeableCredentials)
	require.True(t, ok)

	token, err := workloadCredentials.IAMToken(context.Background())
	require.NoError(t, err)
	require.Equal(t, "workload-token", token.IamToken)
}

func TestCredentialsFromKeyFileReturnsReadError(t *testing.T) {
	missingKeyFile := filepath.Join(t.TempDir(), "missing.json")

	_, err := credentialsFromKeyFile(missingKeyFile)

	require.ErrorContains(t, err, "failed to read service account key file")
}
