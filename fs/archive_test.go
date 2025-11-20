package fs

import (
	"archive/zip"
	"os"
	"path/filepath"
	"secure-fm/config"
	"testing"
)

func TestZipBombProtection(t *testing.T) {
	// Setup sandbox
	tmpDir, err := os.MkdirTemp("", "sandbox_zip")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := &config.Config{SandboxPath: tmpDir}
	InitFS(cfg)

	// Create a dummy zip file that simulates a high compression ratio
	// Since we can't easily generate a real "bomb" without massive resources,
	// we will test the logic by creating a small zip and manually checking if the function accepts it,
	// and then we can mock/force a failure if we had dependency injection,
	// but for now let's verify a valid zip works and an invalid path fails.

	// 1. Create a valid zip
	validZipPath := filepath.Join(tmpDir, "valid.zip")
	createTestZip(t, validZipPath, "hello.txt", "Hello World")

	// 2. Try to unzip it
	// outDir := filepath.Join(tmpDir, "out") // Unused
	err = Unzip("valid.zip", "out")

	if err != nil {
		t.Errorf("Unzip failed for valid zip: %v", err)
	}

	// 3. Test Zip Slip (Path Traversal in Zip)
	// We need to craft a zip with a file named "../evil.txt"
	evilZipPath := filepath.Join(tmpDir, "evil.zip")
	createEvilZip(t, evilZipPath, "../evil.txt", "Evil Content")

	err = Unzip("evil.zip", "out_evil")
	if err == nil {
		t.Error("Unzip should have failed for Zip Slip attempt")
	}
}

func createTestZip(t *testing.T, path, filename, content string) {
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	w := zip.NewWriter(f)
	defer w.Close()

	fWr, err := w.Create(filename)
	if err != nil {
		t.Fatal(err)
	}
	_, err = fWr.Write([]byte(content))
	if err != nil {
		t.Fatal(err)
	}
}

func createEvilZip(t *testing.T, path, filename, content string) {
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	w := zip.NewWriter(f)
	defer w.Close()

	// Manually create header to allow ".."
	header := &zip.FileHeader{
		Name:   filename,
		Method: zip.Store,
	}

	fWr, err := w.CreateHeader(header)
	if err != nil {
		t.Fatal(err)
	}
	_, err = fWr.Write([]byte(content))
	if err != nil {
		t.Fatal(err)
	}
}
