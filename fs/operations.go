package fs

import (
	"errors"
	"io"
	"os"
	"sync"
)

// fileMutex защищает доступ к файлам для предотвращения состояния гонки (race condition)
var fileMutex sync.RWMutex

// DiskInfo содержит информацию о диске/разделе
type DiskInfo struct {
	Name        string  // название/путь
	TotalSize   uint64  // общий объём в байтах
	FreeSpace   uint64  // свободное место в байтах
	UsedSpace   uint64  // использовано в байтах
	UsedPercent float64 // процент использования
}

// ListDrives возвращает список доступных дисков/точек монтирования
func ListDrives() []string {
	// В Docker/Linux диски представлены как точки монтирования
	// Для простоты возвращаем корневой раздел "/"
	return []string{"/"}
}

// ListDirectory возвращает список файлов и папок в указанной директории
func ListDirectory(path string) ([]os.FileInfo, error) {
	safePath, err := ResolvePath(path)
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(safePath)
	if err != nil {
		return nil, err
	}

	infos := make([]os.FileInfo, 0, len(entries))
	for _, entry := range entries {
		info, err := entry.Info()
		if err == nil {
			infos = append(infos, info)
		}
	}
	return infos, nil
}

// CreateDirectory создаёт директорию (включая все родительские)
func CreateDirectory(path string) error {
	safePath, err := ResolvePath(path)
	if err != nil {
		return err
	}

	fileMutex.Lock()
	defer fileMutex.Unlock()

	return os.MkdirAll(safePath, 0755)
}

// ReadFile читает содержимое текстового файла
func ReadFile(path string) (string, error) {
	safePath, err := ResolvePath(path)
	if err != nil {
		return "", err
	}

	fileMutex.RLock()
	defer fileMutex.RUnlock()

	content, err := os.ReadFile(safePath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// WriteFile записывает содержимое в файл
func WriteFile(path string, content string) error {
	// Проверка максимального размера файла (защита от переполнения)
	if len(content) > MaxFileSize {
		return errors.New("размер файла превышает максимально допустимый (10 MB)")
	}

	safePath, err := ResolvePath(path)
	if err != nil {
		return err
	}

	fileMutex.Lock()
	defer fileMutex.Unlock()

	return os.WriteFile(safePath, []byte(content), 0644)
}

// DeleteFile удаляет файл
func DeleteFile(path string) error {
	safePath, err := ResolvePath(path)
	if err != nil {
		return err
	}

	fileMutex.Lock()
	defer fileMutex.Unlock()

	return os.Remove(safePath)
}

// CopyFile копирует файл из src в dst
func CopyFile(src, dst string) error {
	safeSrc, err := ResolvePath(src)
	if err != nil {
		return err
	}
	safeDst, err := ResolvePath(dst)
	if err != nil {
		return err
	}

	// Проверка размера исходного файла
	srcInfo, err := os.Stat(safeSrc)
	if err != nil {
		return err
	}
	if srcInfo.Size() > int64(MaxFileSize) {
		return errors.New("размер исходного файла превышает максимально допустимый (10 MB)")
	}

	fileMutex.RLock()
	srcFile, err := os.Open(safeSrc)
	fileMutex.RUnlock()
	if err != nil {
		return err
	}
	defer srcFile.Close()

	fileMutex.Lock()
	dstFile, err := os.Create(safeDst)
	fileMutex.Unlock()
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// MoveFile перемещает (переименовывает) файл из src в dst
func MoveFile(src, dst string) error {
	safeSrc, err := ResolvePath(src)
	if err != nil {
		return err
	}
	safeDst, err := ResolvePath(dst)
	if err != nil {
		return err
	}

	fileMutex.Lock()
	defer fileMutex.Unlock()

	return os.Rename(safeSrc, safeDst)
}

// AppendFile добавляет содержимое в конец существующего файла
func AppendFile(path string, content string) error {
	safePath, err := ResolvePath(path)
	if err != nil {
		return err
	}

	// Проверяем текущий размер файла + новый контент
	fileMutex.RLock()
	info, err := os.Stat(safePath)
	fileMutex.RUnlock()
	if err != nil {
		return err
	}

	if info.Size()+int64(len(content)) > int64(MaxFileSize) {
		return errors.New("итоговый размер файла превысит максимально допустимый (10 MB)")
	}

	fileMutex.Lock()
	defer fileMutex.Unlock()

	file, err := os.OpenFile(safePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(content)
	return err
}
