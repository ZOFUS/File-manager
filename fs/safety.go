package fs

import (
	"errors"
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
	// Объединяем базовый путь и пользовательский путь
	fullPath := filepath.Join(BaseDir, userPath)
	// Очищаем путь от "." и ".."
	cleanedFullPath := filepath.Clean(fullPath)

	// Проверяем, что результат находится внутри sandbox
	if !strings.HasPrefix(cleanedFullPath, BaseDir) {
		return "", errors.New("доступ запрещён: попытка обхода пути (path traversal)")
	}

	return cleanedFullPath, nil
}
