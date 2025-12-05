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
	// MaxDecompressedSize — максимальный размер распакованных данных (защита от ZIP-бомб)
	MaxDecompressedSize = 100 * 1024 * 1024 // 100 MB
	// MaxCompressionRatio — максимальная степень сжатия (защита от ZIP-бомб)
	MaxCompressionRatio = 100 // 100:1
)

// CreateZip создаёт ZIP-архив из файла или директории
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

	// Обходим все файлы и добавляем их в архив
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
			header.Method = zip.Deflate // Используем сжатие Deflate
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

// Unzip распаковывает ZIP-архив с защитой от ZIP-бомб и Zip Slip
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
		// Защита от ZIP-бомб #1: проверка степени сжатия
		if f.UncompressedSize64 > 0 && float64(f.UncompressedSize64)/float64(f.CompressedSize64) > MaxCompressionRatio {
			return fmt.Errorf("обнаружена ZIP-бомба: слишком высокая степень сжатия для %s", f.Name)
		}

		// Защита от ZIP-бомб #2: проверка общего размера
		totalSize += int64(f.UncompressedSize64)
		if totalSize > MaxDecompressedSize {
			return errors.New("обнаружена ZIP-бомба: превышен лимит размера распакованных данных")
		}

		fpath := filepath.Join(safeDest, f.Name)

		// Защита от Zip Slip (Path Traversal внутри архива)
		if !strings.HasPrefix(fpath, filepath.Clean(safeDest)+string(os.PathSeparator)) {
			return fmt.Errorf("недопустимый путь файла: %s", fpath)
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

		// Ограничиваем чтение для предотвращения бесконечного потока
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
