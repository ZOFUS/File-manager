//go:build windows

package fs

import (
	"errors"
)

// GetDiskInfo возвращает информацию о диске/разделе (Windows заглушка)
// На Windows возвращает ошибку, т.к. приложение предназначено для Docker/Linux
func GetDiskInfo(path string) (*DiskInfo, error) {
	return nil, errors.New("информация о диске недоступна на Windows (запустите в Docker)")
}
