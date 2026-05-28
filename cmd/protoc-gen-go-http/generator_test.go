package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestGoldenStorageHTTP(t *testing.T) {
	if _, err := exec.LookPath("protoc"); err != nil {
		t.Skip("protoc not installed")
	}
	root := findModuleRoot(t)
	golden := filepath.Join(root, "tests", "storage", "storage_http.pb.go")
	tmpDir := t.TempDir()
	work := filepath.Join(tmpDir, "work")
	storageDir := filepath.Join(work, "tests", "storage")
	if err := os.MkdirAll(storageDir, 0o755); err != nil {
		t.Fatal(err)
	}
	copyFile(t, filepath.Join(root, "tests", "storage", "storage.proto"), filepath.Join(storageDir, "storage.proto"))
	copyDir(t, filepath.Join(root, "pluck"), filepath.Join(work, "pluck"))

	annotations := findAnnotationsProto(t)
	plugin := filepath.Join(t.TempDir(), "protoc-gen-go-http")
	if out, buildErr := exec.Command("go", "build", "-o", plugin, ".").CombinedOutput(); buildErr != nil {
		t.Fatalf("build plugin: %v\n%s", buildErr, out)
	}

	cmd := exec.Command("protoc",
		"--proto_path=.",
		"--proto_path="+annotations,
		"--plugin=protoc-gen-go-http="+plugin,
		"--go-http_out=paths=source_relative:.",
		"tests/storage/storage.proto",
	)
	cmd.Dir = work
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("protoc failed: %v\n%s", err, out)
	}

	gotPath := filepath.Join(storageDir, "storage_http.pb.go")
	got, err := os.ReadFile(gotPath)
	if err != nil {
		t.Fatalf("read generated file: %v", err)
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatalf("read golden file: %v", err)
	}
	if normalizeGenerated(string(got)) != normalizeGenerated(string(want)) {
		t.Fatalf("generated output differs from golden file %s", golden)
	}
}

func normalizeGenerated(s string) string {
	lines := strings.Split(s, "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.HasPrefix(line, "// - protoc") {
			continue
		}
		if strings.HasPrefix(line, "// source:") {
			continue
		}
		out = append(out, line)
	}
	return strings.Join(out, "\n")
}

func findModuleRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(wd, "http.tmpl")); err == nil {
		return wd
	}
	if _, err := os.Stat(filepath.Join(wd, "cmd", "protoc-gen-go-http", "http.tmpl")); err == nil {
		return filepath.Join(wd, "cmd", "protoc-gen-go-http")
	}
	t.Fatal("cannot locate protoc-gen-go-http module root")
	return ""
}

func findAnnotationsProto(t *testing.T) string {
	t.Helper()
	out, err := exec.Command("go", "env", "GOMODCACHE").Output()
	if err != nil {
		t.Fatal(err)
	}
	modCache := strings.TrimSpace(string(out))
	candidates := []string{
		filepath.Join(modCache, "github.com/grpc-ecosystem/grpc-gateway@v1.16.0/third_party/googleapis"),
	}
	for _, c := range candidates {
		if _, err := os.Stat(filepath.Join(c, "google", "api", "annotations.proto")); err == nil {
			return c
		}
	}
	t.Skip("google/api/annotations.proto not found in module cache")
	return ""
}

func copyFile(t *testing.T, src, dst string) {
	t.Helper()
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dst, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func copyDir(t *testing.T, src, dst string) {
	t.Helper()
	if err := os.MkdirAll(dst, 0o755); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		copyFile(t, filepath.Join(src, e.Name()), filepath.Join(dst, e.Name()))
	}
}
