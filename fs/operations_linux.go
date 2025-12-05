//go:build linux || darwin

package fs

import (
	"syscall"
)

// GetDiskInfo возвращает информацию о диске/разделе (Linux/macOS)
func GetDiskInfo(path string) (*DiskInfo, error) {
	var stat syscall.Statfs_t
	err := syscall.Statfs(path, &stat)
	if err != nil {
		return nil, err
	}

	total := stat.Blocks * uint64(stat.Bsize)
	free := stat.Bfree * uint64(stat.Bsize)
	used := total - free
	usedPercent := float64(used) / float64(total) * 100

	return &DiskInfo{
		Name:        path,
		TotalSize:   total,
		FreeSpace:   free,
		UsedSpace:   used,
		UsedPercent: usedPercent,
	}, nil
}
