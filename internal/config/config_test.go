package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfigValidateAllowsWorkloadIdentity(t *testing.T) {
	config := Config{FolderID: "folder-id"}

	require.NoError(t, config.validate())
}

func TestConfigValidateRequiresFolderID(t *testing.T) {
	config := Config{}

	require.EqualError(t, config.validate(), "folder_id configuration is required")
}
