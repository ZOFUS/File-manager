package fs

import (
	"io"
	"os"
	"sync"
)

// FileMutex protects access to files to prevent race conditions within the app
var fileMutex sync.RWMutex

func ListDrives() []string {
	// In Docker/Linux, drives are mounts. We can just show root info or specific mounts.
	// For simplicity, we just return "/"
	return []string{"/"}
}

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

func WriteFile(path string, content string) error {
	safePath, err := ResolvePath(path)
	if err != nil {
		return err
	}

	fileMutex.Lock()
	defer fileMutex.Unlock()

	return os.WriteFile(safePath, []byte(content), 0644)
}

func DeleteFile(path string) error {
	safePath, err := ResolvePath(path)
	if err != nil {
		return err
	}

	fileMutex.Lock()
	defer fileMutex.Unlock()

	return os.Remove(safePath)
}

func CopyFile(src, dst string) error {
	safeSrc, err := ResolvePath(src)
	if err != nil {
		return err
	}
	safeDst, err := ResolvePath(dst)
	if err != nil {
		return err
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
