package palworld

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAdapterSettings(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name     string
		settings string
		port     int
		password string
	}{
		{name: "defaults", port: 8212},
		{name: "custom port", settings: `RESTAPIPort=9000,AdminPassword="test-only"`, port: 9000, password: "test-only"},
		{name: "special characters", settings: `AdminPassword="a,b=(c);#"`, port: 8212, password: "a,b=(c);#"},
		{name: "escaped characters", settings: `AdminPassword="quote\"slash\\end"`, port: 8212, password: "quote\"slash\\end"},
		{name: "unquoted password", settings: `AdminPassword=test-only`, port: 8212, password: "test-only"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			config := filepath.Join(dir, "LinuxServer")
			require.NoError(t, os.Mkdir(config, 0o755))
			ini := "[/Script/Pal.PalGameWorldSettings]\nOptionSettings=(RESTAPIEnabled=True," + test.settings + ")\n"
			require.NoError(t, os.WriteFile(filepath.Join(config, "PalWorldSettings.ini"), []byte(ini), 0o644))

			var adapter Adapter
			matched, err := adapter.Detect(dir)
			require.NoError(t, err)
			require.True(t, matched)
			require.Equal(t, test.port, adapter.restPort)
			require.Equal(t, test.password, adapter.adminPassword)
		})
	}
}

func TestAdapterSettingsError(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name     string
		settings string
	}{
		{name: "zero port", settings: `RESTAPIPort=0`},
		{name: "negative port", settings: `RESTAPIPort=-1`},
		{name: "port out of range", settings: `RESTAPIPort=65536`},
		{name: "invalid port", settings: `RESTAPIPort=never-log-secret`},
		{name: "invalid password escape", settings: `AdminPassword="never-log-secret\q"`},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			config := filepath.Join(dir, "LinuxServer")
			require.NoError(t, os.Mkdir(config, 0o755))
			ini := "[/Script/Pal.PalGameWorldSettings]\nOptionSettings=(RESTAPIEnabled=True," + test.settings + ")\n"
			require.NoError(t, os.WriteFile(filepath.Join(config, "PalWorldSettings.ini"), []byte(ini), 0o644))

			adapter := Adapter{restPort: 9000, adminPassword: "old-password"}
			matched, err := adapter.Detect(dir)
			require.Error(t, err)
			require.False(t, matched)
			require.Zero(t, adapter.restPort)
			require.Empty(t, adapter.adminPassword)
			require.NotContains(t, err.Error(), "never-log-secret")
		})
	}
}

func TestAdapterSettingsClearedOnNoMatch(t *testing.T) {
	t.Parallel()
	adapter := Adapter{restPort: 9000, adminPassword: "old-password"}
	matched, err := adapter.Detect(t.TempDir())
	require.NoError(t, err)
	require.False(t, matched)
	require.Zero(t, adapter.restPort)
	require.Empty(t, adapter.adminPassword)
}

func TestAdapterRedactsPassword(t *testing.T) {
	t.Parallel()
	adapter := Adapter{restPort: 8212, adminPassword: "never-log-secret"}
	for _, value := range []any{adapter, &adapter} {
		for _, format := range []string{"%v", "%+v", "%#v"} {
			output := fmt.Sprintf(format, value)
			require.NotContains(t, output, adapter.adminPassword)
			require.Contains(t, output, "[REDACTED]")
		}
	}
	for _, json := range []bool{false, true} {
		var output bytes.Buffer
		var handler slog.Handler = slog.NewTextHandler(&output, nil)
		if json {
			handler = slog.NewJSONHandler(&output, nil)
		}
		slog.New(handler).Info("adapter settings", "adapter", &adapter)
		require.NotContains(t, output.String(), adapter.adminPassword)
	}
}
