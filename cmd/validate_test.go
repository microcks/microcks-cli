package cmd

import (
	"os"
	"testing"

	"github.com/microcks/microcks-cli/pkg/config"
	"github.com/stretchr/testify/require"
)

const validateTestConfigFilePath = "./testdata/validate.config"

// setupValidateTestConfig writes the shared testConfig (defined in
// context_test.go) to disk so ReadLocalConfig can load it, and returns
// a cleanup func to remove it afterwards.
func setupValidateTestConfig(t *testing.T) func() {
	t.Helper()
	err := os.MkdirAll("./testdata", 0755)
	require.NoError(t, err)
	err = os.WriteFile(validateTestConfigFilePath, []byte(testConfig), 0644)
	require.NoError(t, err)
	return func() {
		_ = os.Remove(validateTestConfigFilePath)
	}
}

func TestReadLocalConfig_ForValidate_ParsesSuccessfully(t *testing.T) {
	cleanup := setupValidateTestConfig(t)
	defer cleanup()

	localConfig, err := config.ReadLocalConfig(validateTestConfigFilePath)
	require.NoError(t, err)
	require.NotNil(t, localConfig)
	require.False(t, localConfig.IsEmpty())
}

func TestResolveContext_ForValidate_ResolvesCurrentContext(t *testing.T) {
	cleanup := setupValidateTestConfig(t)
	defer cleanup()

	localConfig, err := config.ReadLocalConfig(validateTestConfigFilePath)
	require.NoError(t, err)

	ctx, err := localConfig.ResolveContext("")
	require.NoError(t, err)
	require.Equal(t, "http://localhost:8083", ctx.Server.Server)
}

func TestReadLocalConfig_ForValidate_MissingFileReturnsNilNoError(t *testing.T) {
	// Non-existent path should surface as a nil config, not a hard error,
	// matching the same contract other commands (e.g. stop) rely on.
	localConfig, err := config.ReadLocalConfig("./testdata/does-not-exist.config")
	require.NoError(t, err)
	require.Nil(t, localConfig)
}
