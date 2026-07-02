package cmdCli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/tea4go/jvms/internal/entity"
)

func TestLoadConfigFromPathBacksUpCorruptConfig(t *testing.T) {
	configFile := filepath.Join(t.TempDir(), "jvms.json")
	if err := os.WriteFile(configFile, []byte(`{"web_type":"huawei"}}`), 0644); err != nil {
		t.Fatalf("write corrupt config: %v", err)
	}

	var cfg entity.TConfig
	if err := loadConfigFromPath(configFile, &cfg); err != nil {
		t.Fatalf("expected corrupt config to be backed up and reset, got: %v", err)
	}

	matches, err := filepath.Glob(configFile + ".bad.*")
	if err != nil {
		t.Fatalf("glob corrupt config backup: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("expected one corrupt config backup, got %d", len(matches))
	}
	if _, err := os.Stat(configFile); !os.IsNotExist(err) {
		t.Fatalf("expected corrupt config to be moved away, stat err: %v", err)
	}
}

func TestSaveConfigToPathWritesValidJSON(t *testing.T) {
	configFile := filepath.Join(t.TempDir(), "jvms.json")
	cfg := entity.TConfig{
		JavaHome:          "/tmp/jdk",
		CurrentJDKVersion: "openjdk-21",
		WebType:           "huawei",
		WebAll:            true,
		Store:             "/tmp/store",
		Download:          "/tmp/download",
	}

	if err := saveConfigToPath(configFile, &cfg); err != nil {
		t.Fatalf("save config: %v", err)
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("read saved config: %v", err)
	}
	var decoded entity.TConfig
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("expected saved config to be valid JSON, got: %v\n%s", err, string(data))
	}
	if decoded.WebType != "huawei" {
		t.Fatalf("expected saved web type huawei, got %q", decoded.WebType)
	}
}
