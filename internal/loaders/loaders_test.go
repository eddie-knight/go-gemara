package loaders

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// Dummy struct for testing
type dummyStruct struct {
	Field string `yaml:"field" json:"field"`
}

func TestDecodeYAMLFromFile(t *testing.T) {
	// Create a temporary YAML file
	file, err := os.CreateTemp("", "test-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() {
		if err := os.Remove(file.Name()); err != nil {
			t.Logf("failed to remove temp file: %v", err)
		}
	}()
	if _, err := file.WriteString("field: value\n"); err != nil {
		t.Fatalf("failed to write to file: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("failed to close file: %v", err)
	}

	var target dummyStruct
	err = decodeYAMLFromFile(file.Name(), &target)
	if err != nil {
		t.Errorf("decodeYAMLFromFile() error = %v, wantErr %v", err, false)
	}
	if target.Field != "value" {
		t.Errorf("decodeYAMLFromFile() got = %v, want %v", target.Field, "value")
	}
}

func TestDecodeYAMLFromReader(t *testing.T) {
	reader := strings.NewReader("field: value\n")
	var target dummyStruct
	err := decodeYAMLFromReader(reader, &target)
	if err != nil {
		t.Errorf("decodeYAMLFromReader() error = %v, wantErr %v", err, false)
	}
	if target.Field != "value" {
		t.Errorf("decodeYAMLFromReader() got = %v, want %v", target.Field, "value")
	}
}

func TestDecodeJSONFromFile(t *testing.T) {
	file, err := os.CreateTemp("", "test-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() {
		if err := os.Remove(file.Name()); err != nil {
			t.Logf("failed to remove temp file: %v", err)
		}
	}()
	if _, err := file.WriteString("{\"field\": \"value\"}"); err != nil {
		t.Fatalf("failed to write to file: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("failed to close file: %v", err)
	}

	var target dummyStruct
	err = decodeJSONFromFile(file.Name(), &target)
	if err != nil {
		t.Errorf("decodeJSONFromFile() error = %v, wantErr %v", err, false)
	}
	if target.Field != "value" {
		t.Errorf("decodeJSONFromFile() got = %v, want %v", target.Field, "value")
	}
}

func TestDecodeJSONFromReader(t *testing.T) {
	reader := strings.NewReader("{\"field\": \"value\"}")
	var target dummyStruct
	err := decodeJSONFromReader(reader, &target)
	if err != nil {
		t.Errorf("decodeJSONFromReader() error = %v, wantErr %v", err, false)
	}
	if target.Field != "value" {
		t.Errorf("decodeJSONFromReader() got = %v, want %v", target.Field, "value")
	}
}

func TestMarshalUnmarshalYAML(t *testing.T) {
	obj := dummyStruct{Field: "value"}
	bytes, err := MarshalYAML(obj)
	if err != nil {
		t.Errorf("MarshalYAML() error = %v, wantErr %v", err, false)
	}
	var target dummyStruct
	err = UnmarshalYAML(bytes, &target)
	if err != nil {
		t.Errorf("UnmarshalYAML() error = %v, wantErr %v", err, false)
	}
	if target.Field != "value" {
		t.Errorf("UnmarshalYAML() got = %v, want %v", target.Field, "value")
	}
}
func TestLoadYAML_FileScheme(t *testing.T) {
	file, err := os.CreateTemp("", "test-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() {
		if err := os.Remove(file.Name()); err != nil {
			t.Logf("failed to remove temp file: %v", err)
		}
	}()
	if _, err := file.WriteString("field: value\n"); err != nil {
		t.Fatalf("failed to write to file: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("failed to close file: %v", err)
	}

	var target dummyStruct
	err = LoadYAML("file://"+file.Name(), &target)
	if err != nil {
		t.Errorf("LoadYAML() error = %v, wantErr %v", err, false)
	}
	if target.Field != "value" {
		t.Errorf("LoadYAML() got = %v, want %v", target.Field, "value")
	}
}

func TestLoadYAML_UnsupportedScheme(t *testing.T) {
	var target dummyStruct
	err := LoadYAML("ftp://example.com/file.yaml", &target)
	if err == nil {
		t.Errorf("LoadYAML() error = %v, wantErr %v", err, true)
	}
}

func TestLoadNotAURL(t *testing.T) {
	var target dummyStruct
	err := LoadYAML("ftpyaml", &target)
	if err == nil {
		t.Errorf("LoadYAML() error = %v, wantErr %v", err, true)
	}
}

func TestLoadYAML_HTTPS(t *testing.T) {
	// Use a local HTTPS server so this test suite never makes network calls.
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("title: Access Control\n"))
	}))
	defer srv.Close()

	originalTransport := http.DefaultTransport
	http.DefaultTransport = srv.Client().Transport
	defer func() { http.DefaultTransport = originalTransport }()

	type FauxCatalog struct {
		Title string `yaml:"title"`
	}
	var target FauxCatalog
	err := LoadYAML(srv.URL+"/baseline/OSPS-AC.yaml", &target)
	if err != nil {
		t.Errorf("LoadYAML() error = %v, wantErr %v", err, false)
	}
	if target.Title != "Access Control" {
		t.Errorf("LoadYAML() failed to decode expected title, got = %v", target.Title)
	}
}

func TestLoadInvalidURL(t *testing.T) {
	// Verify we return an error for a non-200 response without hitting the network.
	srv := httptest.NewTLSServer(http.NotFoundHandler())
	defer srv.Close()

	originalTransport := http.DefaultTransport
	http.DefaultTransport = srv.Client().Transport
	defer func() { http.DefaultTransport = originalTransport }()

	var target dummyStruct
	err := LoadYAML(srv.URL+"/doesnotexist.yaml", &target)
	if err == nil {
		t.Errorf("LoadYAML() error = %v, wantErr %v", err, true)
	}
}

func TestLoadJSON_FileScheme(t *testing.T) {
	file, err := os.CreateTemp("", "test-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() {
		if err := os.Remove(file.Name()); err != nil {
			t.Fatalf("failed to remove file: %v", err)
		}
	}()
	if _, err := file.WriteString("{\"field\": \"value\"}"); err != nil {
		t.Fatalf("failed to write to file: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("failed to close file: %v", err)
	}

	var target dummyStruct
	err = LoadJSON("file://"+file.Name(), &target)
	if err != nil {
		t.Errorf("LoadJSON() error = %v, wantErr %v", err, false)
	}
	if target.Field != "value" {
		t.Errorf("LoadJSON() got = %v, want %v", target.Field, "value")
	}
}

func TestLoadJSON_UnsupportedScheme(t *testing.T) {
	var target dummyStruct
	err := LoadJSON("ftp://example.com/file.json", &target)
	if err == nil {
		t.Errorf("LoadJSON() error = %v, wantErr %v", err, true)
	}
}

func TestLoadJSON_InvalidURL(t *testing.T) {
	// Verify we return an error for a non-200 response without hitting the network.
	srv := httptest.NewTLSServer(http.NotFoundHandler())
	defer srv.Close()

	originalTransport := http.DefaultTransport
	http.DefaultTransport = srv.Client().Transport
	defer func() { http.DefaultTransport = originalTransport }()

	var target dummyStruct
	err := LoadJSON(srv.URL+"/doesnotexist.json", &target)
	if err == nil {
		t.Errorf("LoadJSON() error = %v, wantErr %v", err, true)
	}
}
