package fs

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	MaxDecompressedSize = 100 * 1024 * 1024 // 100 MB Limit
	MaxCompressionRatio = 100               // 100:1 Ratio Limit
)

func CreateZip(source, target string) error {
	safeSource, err := ResolvePath(source)
	if err != nil {
		return err
	}
	safeTarget, err := ResolvePath(target)
	if err != nil {
		return err
	}

	zipFile, err := os.Create(safeTarget)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	archive := zip.NewWriter(zipFile)
	defer archive.Close()

	info, err := os.Stat(safeSource)
	if err != nil {
		return err
	}

	var baseDir string
	if info.IsDir() {
		baseDir = filepath.Base(safeSource)
	}

	filepath.Walk(safeSource, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		if baseDir != "" {
			header.Name = filepath.Join(baseDir, strings.TrimPrefix(path, safeSource))
		}

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(writer, file)
		return err
	})

	return err
}

func Unzip(src, dest string) error {
	safeSrc, err := ResolvePath(src)
	if err != nil {
		return err
	}
	safeDest, err := ResolvePath(dest)
	if err != nil {
		return err
	}

	r, err := zip.OpenReader(safeSrc)
	if err != nil {
		return err
	}
	defer r.Close()

	var totalSize int64

	for _, f := range r.File {
		// Zip Bomb Protection 1: Check Ratio
		if f.UncompressedSize64 > 0 && float64(f.UncompressedSize64)/float64(f.CompressedSize64) > MaxCompressionRatio {
			return fmt.Errorf("zip bomb detected: compression ratio too high for %s", f.Name)
		}

		// Zip Bomb Protection 2: Check Total Size
		totalSize += int64(f.UncompressedSize64)
		if totalSize > MaxDecompressedSize {
			return errors.New("zip bomb detected: total decompressed size exceeds limit")
		}

		fpath := filepath.Join(safeDest, f.Name)

		// Check for Zip Slip (Path Traversal inside Zip)
		if !strings.HasPrefix(fpath, filepath.Clean(safeDest)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", fpath)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		// Limit reader to prevent infinite stream
		limitReader := io.LimitReader(rc, MaxDecompressedSize)
		_, err = io.Copy(outFile, limitReader)

		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}
	return nil
}
