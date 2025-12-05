package fs

import (
	"errors"
	"net/url"
	"path/filepath"
	"secure-fm/config"
	"strings"
)

// BaseDir — базовая директория sandbox, относительно которой работают все операции
var BaseDir string

// MaxFileSize — максимальный размер файла для записи (10 MB)
// Защита от DoS-атаки через загрузку больших файлов
const MaxFileSize = 10 * 1024 * 1024 // 10 MB

// InitFS инициализирует файловую систему, устанавливая путь к sandbox
func InitFS(cfg *config.Config) {
	BaseDir = cfg.SandboxPath
	// Преобразуем путь в абсолютный для надёжности
	abs, err := filepath.Abs(BaseDir)
	if err == nil {
		BaseDir = abs
	}
}

// ResolvePath проверяет и преобразует пользовательский путь в безопасный
// Защита от атаки Path Traversal (обход пути)
func ResolvePath(userPath string) (string, error) {
	// Защита #1: декодируем URL-encoded символы (%2F, %2E и т.д.)
	decodedPath, err := url.QueryUnescape(userPath)
	if err != nil {
		decodedPath = userPath
	}

	// Защита #2: проверка на null byte (попытка обрезать строку)
	if strings.Contains(decodedPath, "\x00") {
		return "", errors.New("доступ запрещён: недопустимые символы в пути")
	}

	// Защита #3: запрет абсолютных путей
	if filepath.IsAbs(decodedPath) || strings.HasPrefix(decodedPath, "/") || strings.HasPrefix(decodedPath, "\\") {
		return "", errors.New("доступ запрещён: абсолютные пути запрещены")
	}

	// Защита #4: запрет явного обхода через ".."
	if strings.Contains(decodedPath, "..") {
		return "", errors.New("доступ запрещён: попытка обхода пути (path traversal)")
	}

	// Объединяем базовый путь и пользовательский путь
	fullPath := filepath.Join(BaseDir, decodedPath)
	// Очищаем путь от "." и ".."
	cleanedFullPath := filepath.Clean(fullPath)

	// Защита #5: финальная проверка что результат внутри sandbox
	if !strings.HasPrefix(cleanedFullPath, BaseDir) {
		return "", errors.New("доступ запрещён: попытка обхода пути (path traversal)")
	}

	return cleanedFullPath, nil
}
