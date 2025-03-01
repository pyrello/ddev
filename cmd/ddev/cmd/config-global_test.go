package cmd

import (
	"github.com/drud/ddev/pkg/exec"
	"github.com/drud/ddev/pkg/fileutil"
	"github.com/drud/ddev/pkg/globalconfig"
	asrt "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

// Run with various flags
// Try to create errors
// Validate that what it spits out is what's there

func TestCmdGlobalConfig(t *testing.T) {
	assert := asrt.New(t)

	backupConfig := globalconfig.DdevGlobalConfig
	// Start with no config file
	configFile := globalconfig.GetGlobalConfigPath()
	if fileutil.FileExists(configFile) {
		err := os.Remove(configFile)
		require.NoError(t, err)
	}
	// We need to make sure that the (corrupted, bogus) global config file is removed
	// and then read (empty)
	// nolint: errcheck
	t.Cleanup(func() {
		// Even though the global config is going to be deleted, make sure it's sane before leaving
		args := []string{"config", "global", "--omit-containers", "", "--nfs-mount-enabled", "--disable-http2=false", "--mutagen-enabled=false", "--simple-formatting=false", "--table-style=default", `--required-docker-compose-version=""`, `--use-docker-compose-from-path=false`, `xdebug-ide-location=""`}
		globalconfig.DdevGlobalConfig.OmitContainersGlobal = nil
		_, err := exec.RunHostCommand(DdevBin, args...)
		assert.NoError(err)
		globalconfig.DdevGlobalConfig = backupConfig
		globalconfig.DdevGlobalConfig.OmitContainersGlobal = nil

		err = os.Remove(configFile)
		if err != nil {
			t.Logf("Unable to remove %v: %v", configFile, err)
		}
		err = globalconfig.ReadGlobalConfig()
		if err != nil {
			t.Logf("Unable to ReadGlobalConfig: %v", err)
		}
	})

	// Look at initial config
	args := []string{"config", "global"}
	out, err := exec.RunCommand(DdevBin, args)
	assert.NoError(err)
	assert.Contains(string(out), "Global configuration:\ninstrumentation-opt-in=false\nomit-containers=[]\nweb-environment=[]\nmutagen-enabled=false\nnfs-mount-enabled=false\nrouter-bind-all-interfaces=false\ninternet-detection-timeout-ms=3000\ndisable-http2=false\nuse-letsencrypt=false\nletsencrypt-email=\ntable-style=default\nsimple-formatting=false\nauto-restart-containers=false\nuse-hardened-images=false\nfail-on-hook-fail=false\nrequired-docker-compose-version=\nuse-docker-compose-from-path=false\nproject-tld=\nxdebug-ide-location=\n")

	// Update a config
	// Don't include no-bind-mounts because global testing
	// will turn it on and break this
	args = []string{"config", "global", "--project-tld=ddev.test", "--instrumentation-opt-in=false", "--omit-containers=dba,ddev-ssh-agent", "--mutagen-enabled=true", "--nfs-mount-enabled=true", "--router-bind-all-interfaces=true", "--internet-detection-timeout-ms=850", "--use-letsencrypt", "--letsencrypt-email=nobody@example.com", "--table-style=bright", "--simple-formatting=true", "--auto-restart-containers=true", "--use-hardened-images=true", "--fail-on-hook-fail=true", `--disable-http2`, `--web-environment="SOMEENV=some+val"`, `--xdebug-ide-location=container`}
	out, err = exec.RunCommand(DdevBin, args)
	assert.NoError(err)
	assert.Contains(string(out), "Global configuration:\ninstrumentation-opt-in=false\nomit-containers=[dba,ddev-ssh-agent]\nweb-environment=[\"SOMEENV=some+val\"]\nmutagen-enabled=true\nnfs-mount-enabled=true\nrouter-bind-all-interfaces=true\ninternet-detection-timeout-ms=850\ndisable-http2=true\nuse-letsencrypt=true\nletsencrypt-email=nobody@example.com\ntable-style=bright\nsimple-formatting=true\nauto-restart-containers=true\nuse-hardened-images=true\nfail-on-hook-fail=true\nrequired-docker-compose-version=\nuse-docker-compose-from-path=false")
	assert.Contains(string(out), "xdebug-ide-location=container")

	err = globalconfig.ReadGlobalConfig()
	assert.NoError(err)
	assert.False(globalconfig.DdevGlobalConfig.InstrumentationOptIn)
	assert.Contains(globalconfig.DdevGlobalConfig.OmitContainersGlobal, "ddev-ssh-agent")
	assert.Contains(globalconfig.DdevGlobalConfig.OmitContainersGlobal, "dba")
	assert.True(globalconfig.DdevGlobalConfig.MutagenEnabledGlobal)
	assert.True(globalconfig.DdevGlobalConfig.NFSMountEnabledGlobal)
	assert.Len(globalconfig.DdevGlobalConfig.OmitContainersGlobal, 2)
	assert.Equal("nobody@example.com", globalconfig.DdevGlobalConfig.LetsEncryptEmail)
	assert.Equal("ddev.test", globalconfig.DdevGlobalConfig.ProjectTldGlobal)
	assert.True(globalconfig.DdevGlobalConfig.UseLetsEncrypt)
	assert.True(globalconfig.DdevGlobalConfig.UseHardenedImages)
	assert.True(globalconfig.DdevGlobalConfig.FailOnHookFailGlobal)
	assert.True(globalconfig.DdevGlobalConfig.DisableHTTP2)
	assert.True(globalconfig.DdevGlobalConfig.SimpleFormatting)
	assert.Equal("bright", globalconfig.DdevGlobalConfig.TableStyle)
	assert.Equal("container", globalconfig.DdevGlobalConfig.XdebugIDELocation)

	// Test that variables can be appended to the web environment
	args = []string{"config", "global", `--web-environment-add="FOO=bar"`}
	out, err = exec.RunCommand(DdevBin, args)
	assert.NoError(err)
	assert.Contains(string(out), "web-environment=[\"FOO=bar\",\"SOMEENV=some+val\"]")
}
