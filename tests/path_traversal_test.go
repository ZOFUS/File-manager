package tests

import (
	"os"
	"path/filepath"
	"testing"

	"secure-fm/config"
	"secure-fm/fs"
)

// TestPathTraversal проверяет защиту от атак обхода пути (Path Traversal)
// Уязвимость: пользователь пытается получить доступ к файлам вне sandbox
func TestPathTraversal(t *testing.T) {
	// Создаём временную sandbox
	tmpDir, err := os.MkdirTemp("", "sandbox_pathtest")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := &config.Config{SandboxPath: tmpDir}
	fs.InitFS(cfg)

	// Создаём тестовый файл внутри sandbox
	testFile := filepath.Join(tmpDir, "safe.txt")
	os.WriteFile(testFile, []byte("safe content"), 0644)

	testCases := []struct {
		name      string
		path      string
		wantError bool
		desc      string
	}{
		{
			name:      "Valid_SimpleFile",
			path:      "safe.txt",
			wantError: false,
			desc:      "Обычный файл в sandbox - должен работать",
		},
		{
			name:      "Valid_NestedPath",
			path:      "subdir/file.txt",
			wantError: false,
			desc:      "Вложенный путь - должен работать",
		},
		{
			name:      "Attack_ParentDir",
			path:      "../secret.txt",
			wantError: true,
			desc:      "Попытка выйти на уровень выше - ДОЛЖНА БЛОКИРОВАТЬСЯ",
		},
		{
			name:      "Attack_DeepTraversal",
			path:      "../../etc/passwd",
			wantError: true,
			desc:      "Попытка добраться до /etc/passwd - ДОЛЖНА БЛОКИРОВАТЬСЯ",
		},
		{
			name:      "Attack_AbsolutePath",
			path:      "/etc/passwd",
			wantError: true,
			desc:      "Абсолютный путь вне sandbox - ДОЛЖЕН БЛОКИРОВАТЬСЯ",
		},
		{
			name:      "Attack_EncodedTraversal",
			path:      "..%2F..%2Fetc/passwd",
			wantError: true,
			desc:      "URL-encoded traversal - ДОЛЖЕН БЛОКИРОВАТЬСЯ",
		},
		{
			name:      "Attack_MixedSlashes",
			path:      "..\\..\\windows\\system32",
			wantError: true,
			desc:      "Windows-style traversal - ДОЛЖЕН БЛОКИРОВАТЬСЯ",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := fs.ResolvePath(tc.path)

			if tc.wantError && err == nil {
				t.Errorf("УЯЗВИМОСТЬ! %s: путь '%s' должен быть заблокирован, но прошёл", tc.desc, tc.path)
			} else if !tc.wantError && err != nil {
				t.Errorf("Ложное срабатывание: %s: путь '%s' заблокирован ошибочно: %v", tc.desc, tc.path, err)
			} else if tc.wantError && err != nil {
				t.Logf("✅ ЗАЩИТА РАБОТАЕТ: %s", tc.desc)
			} else {
				t.Logf("✅ OK: %s", tc.desc)
			}
		})
	}
}
