package integration

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"terrestrial-dtn/pkg/hdtnconfig"
)

// projectRoot returns the path to the project root directory.
func projectRoot(t *testing.T) string {
	t.Helper()
	_, filename, _, _ := runtime.Caller(0)
	// test/integration/smoke_test.go → project root is ../../
	return filepath.Join(filepath.Dir(filename), "..", "..")
}

// TestSmoke_ConfigFilesParseAsValidJSON verifies that the HDTN config files
// for node-a and node-b parse as valid JSON and pass validation.
func TestSmoke_ConfigFilesParseAsValidJSON(t *testing.T) {
	root := projectRoot(t)

	files := []string{
		filepath.Join(root, "configs", "hdtn-config-node-a.json"),
		filepath.Join(root, "configs", "hdtn-config-node-b.json"),
	}

	for _, path := range files {
		t.Run(filepath.Base(path), func(t *testing.T) {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("failed to read %s: %v", path, err)
			}

			var cfg hdtnconfig.HDTNConfig
			if err := json.Unmarshal(data, &cfg); err != nil {
				t.Fatalf("failed to parse %s as JSON: %v", path, err)
			}

			if err := cfg.Validate(); err != nil {
				t.Fatalf("%s fails validation: %v", path, err)
			}
		})
	}
}

// TestSmoke_ScriptsHaveExecutePermission verifies that all scripts/ files
// have execute permission.
func TestSmoke_ScriptsHaveExecutePermission(t *testing.T) {
	root := projectRoot(t)
	scriptsDir := filepath.Join(root, "scripts")

	scripts := []string{
		"build-hdtn.sh",
		"start-node-a.sh",
		"start-node-b.sh",
		"stop-node.sh",
	}

	for _, script := range scripts {
		t.Run(script, func(t *testing.T) {
			path := filepath.Join(scriptsDir, script)
			info, err := os.Stat(path)
			if err != nil {
				t.Fatalf("script %s not found: %v", script, err)
			}

			mode := info.Mode()
			if mode&0111 == 0 {
				t.Fatalf("script %s does not have execute permission (mode: %s)", script, mode)
			}
		})
	}
}

// TestSmoke_ObsoleteCodeRemoved verifies that ax25/, pkg/ion/, and pkg/ionconfig/
// directories do not exist.
func TestSmoke_ObsoleteCodeRemoved(t *testing.T) {
	root := projectRoot(t)

	obsoleteDirs := []string{
		filepath.Join(root, "ax25"),
		filepath.Join(root, "pkg", "ion"),
		filepath.Join(root, "pkg", "ionconfig"),
	}

	for _, dir := range obsoleteDirs {
		t.Run(filepath.Base(dir), func(t *testing.T) {
			if _, err := os.Stat(dir); err == nil {
				t.Fatalf("obsolete directory %s still exists", dir)
			}
		})
	}
}

// TestSmoke_KissPackageExists verifies that the kiss/ package directory exists.
func TestSmoke_KissPackageExists(t *testing.T) {
	root := projectRoot(t)
	kissDir := filepath.Join(root, "kiss")

	info, err := os.Stat(kissDir)
	if err != nil {
		t.Fatalf("kiss/ directory not found: %v", err)
	}
	if !info.IsDir() {
		t.Fatal("kiss/ is not a directory")
	}

	// Verify kiss.go exists
	kissFile := filepath.Join(kissDir, "kiss.go")
	if _, err := os.Stat(kissFile); err != nil {
		t.Fatalf("kiss/kiss.go not found: %v", err)
	}
}

// TestSmoke_HDTNPackagesExist verifies that the new HDTN packages exist.
func TestSmoke_HDTNPackagesExist(t *testing.T) {
	root := projectRoot(t)

	packages := []struct {
		dir  string
		file string
	}{
		{"pkg/hdtn", "lifecycle.go"},
		{"pkg/hdtn", "telemetry.go"},
		{"pkg/hdtn", "contactplan.go"},
		{"pkg/hdtnconfig", "config.go"},
		{"pkg/hdtnconfig", "generate.go"},
	}

	for _, pkg := range packages {
		t.Run(pkg.dir+"/"+pkg.file, func(t *testing.T) {
			path := filepath.Join(root, pkg.dir, pkg.file)
			if _, err := os.Stat(path); err != nil {
				t.Fatalf("%s not found: %v", path, err)
			}
		})
	}
}

// TestSmoke_KISSCLAPluginExists verifies that the C++ KISS CLA plugin files exist.
func TestSmoke_KISSCLAPluginExists(t *testing.T) {
	root := projectRoot(t)
	pluginDir := filepath.Join(root, "plugins", "kiss-cla")

	files := []string{
		"kiss_cla_plugin.h",
		"kiss_cla_plugin.cpp",
		"CMakeLists.txt",
	}

	for _, file := range files {
		t.Run(file, func(t *testing.T) {
			path := filepath.Join(pluginDir, file)
			if _, err := os.Stat(path); err != nil {
				t.Fatalf("%s not found: %v", file, err)
			}
		})
	}
}
