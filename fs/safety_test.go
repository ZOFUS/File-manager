package fs

import (
	"os"
	"secure-fm/config"
	"testing"
)

func TestResolvePath(t *testing.T) {
	// Setup temporary sandbox
	tmpDir, err := os.MkdirTemp("", "sandbox")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Init FS with this sandbox
	cfg := &config.Config{SandboxPath: tmpDir}
	InitFS(cfg)

	tests := []struct {
		name      string
		input     string
		wantError bool
	}{
		{"Valid file", "test.txt", false},
		{"Valid nested file", "subdir/test.txt", false},
		{"Path traversal attempt", "../secret.txt", true},
		{"Deep path traversal", "../../etc/passwd", true},
		{"Traversal with valid prefix", "../sandbox_neighbor/test.txt", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ResolvePath(tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("ResolvePath(%q) error = %v, wantError %v", tt.input, err, tt.wantError)
			}
		})
	}
}
