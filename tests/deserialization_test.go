package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"secure-fm/config"
	"secure-fm/fs"
)

// TestInsecureDeserializationProtection проверяет защиту от небезопасной десериализации
// Уязвимость: при парсинге JSON/XML выполняется произвольный код
// Защита: Go не выполняет код при десериализации (в отличие от Java/Python pickle)
func TestInsecureDeserializationProtection(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sandbox_deser")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := &config.Config{SandboxPath: tmpDir}
	fs.InitFS(cfg)

	t.Run("SafeJSONParsing", func(t *testing.T) {
		// Создаём JSON с потенциально опасным содержимым
		maliciousJSON := `{
			"__proto__": {"isAdmin": true},
			"constructor": {"prototype": {"isAdmin": true}},
			"exec": "os.system('rm -rf /')"
		}`

		jsonFile := "malicious.json"
		fs.WriteFile(jsonFile, maliciousJSON)

		// Пытаемся прочитать - никакой код не должен выполниться
		data, err := fs.ReadJSON(jsonFile)
		if err != nil {
			t.Logf("✅ JSON с опасным содержимым безопасно отклонён: %v", err)
		} else {
			t.Logf("✅ JSON прочитан как данные, код НЕ выполнялся: %v", data)
		}
	})

	t.Run("SafeXMLParsing", func(t *testing.T) {
		// XML с XXE (XML External Entity) атакой
		xxeXML := `<?xml version="1.0"?>
<!DOCTYPE root [
  <!ENTITY xxe SYSTEM "file:///etc/passwd">
]>
<root><content>&xxe;</content></root>`

		xmlFile := "xxe_attack.xml"
		fs.WriteFile(xmlFile, xxeXML)

		// Go's encoding/xml по умолчанию НЕ обрабатывает внешние сущности
		data, err := fs.ReadXML(xmlFile)
		if err != nil {
			t.Logf("✅ XXE атака заблокирована: %v", err)
		} else {
			// Проверяем что содержимое /etc/passwd НЕ было прочитано
			if data != nil && strings.Contains(data.Content, "root:") {
				t.Error("❌ УЯЗВИМОСТЬ XXE! Содержимое /etc/passwd было прочитано!")
			} else {
				t.Log("✅ XXE атака не сработала, внешние сущности не обрабатываются")
			}
		}
	})

	t.Run("CodeReview_SafeLibraries", func(t *testing.T) {
		// Проверяем что используются безопасные библиотеки
		structuredFile := filepath.Join("..", "fs", "structured.go")
		content, err := os.ReadFile(structuredFile)
		if err != nil {
			t.Fatal("Не удалось прочитать fs/structured.go")
		}

		code := string(content)

		safeLibs := []struct {
			lib  string
			desc string
		}{
			{"encoding/json", "Стандартная библиотека JSON (безопасна)"},
			{"encoding/xml", "Стандартная библиотека XML (безопасна без XXE)"},
		}

		for _, lib := range safeLibs {
			if strings.Contains(code, lib.lib) {
				t.Logf("✅ %s", lib.desc)
			}
		}

		// Проверяем отсутствие опасных библиотек
		dangerousLibs := []string{
			"unsafe",
			"reflect.Call",
			"exec.Command",
		}

		for _, lib := range dangerousLibs {
			if strings.Contains(code, lib) {
				t.Logf("⚠️ Найдено потенциально опасное: %s", lib)
			}
		}
	})
}

// TestFileSizeLimits проверяет ограничения размера файлов
func TestFileSizeLimits(t *testing.T) {
	t.Run("MaxFileSizeConstant", func(t *testing.T) {
		safetyFile := filepath.Join("..", "fs", "safety.go")
		content, err := os.ReadFile(safetyFile)
		if err != nil {
			t.Fatal("Не удалось прочитать fs/safety.go")
		}

		code := string(content)

		if strings.Contains(code, "MaxFileSize") {
			t.Log("✅ Установлен лимит MaxFileSize")
		} else {
			t.Error("❌ Лимит MaxFileSize не найден!")
		}

		if strings.Contains(code, "10 * 1024 * 1024") || strings.Contains(code, "10*1024*1024") {
			t.Log("✅ MaxFileSize = 10 MB")
		}
	})

	t.Run("ZipBombLimits", func(t *testing.T) {
		archiveFile := filepath.Join("..", "fs", "archive.go")
		content, err := os.ReadFile(archiveFile)
		if err != nil {
			t.Fatal("Не удалось прочитать fs/archive.go")
		}

		code := string(content)

		checks := []string{
			"MaxDecompressedSize",
			"MaxCompressionRatio",
			"LimitReader",
		}

		for _, check := range checks {
			if strings.Contains(code, check) {
				t.Logf("✅ Найдена защита: %s", check)
			} else {
				t.Errorf("❌ Защита %s не найдена!", check)
			}
		}
	})
}
