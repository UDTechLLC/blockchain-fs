package util

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// TEST: TestUnzipFile
func UnzipFile(archive, target string) (err error) {
	reader, err := zip.OpenReader(archive)
	if err != nil {
		return err
	}
	defer reader.Close()

	if err := os.MkdirAll(target, 0755); err != nil {
		return err
	}

	for _, file := range reader.File {
		path := filepath.Join(target, file.Name)

		if file.FileInfo().IsDir() {
			os.MkdirAll(path, file.Mode())
			continue
		}

		fileReader, err := file.Open()
		if err != nil {
			if fileReader != nil {
				fileReader.Close()
			}

			return err
		}

		targetFile, err := os.OpenFile(
			path,
			os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
			file.Mode(),
		)
		if err != nil {
			fileReader.Close()
			if targetFile != nil {
				targetFile.Close()
			}

			return err
		}
		defer targetFile.Close()

		if _, err = io.Copy(targetFile, fileReader); err != nil {
			fileReader.Close()
			targetFile.Close()

			return err
		}

		fileReader.Close()
		targetFile.Close()
	}

	return nil
}

// TEST: TestZipFile
func ZipFile(source, target string) (err error) {
	zipfile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	info, err := os.Stat(source)
	if err != nil {
		return nil
	}

	sourceIsDir := info.IsDir()

	filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		if sourceIsDir {
			header.Name = strings.TrimPrefix(path, source)
		}

		if header.Name == "" {
			return nil
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
			if file != nil {
				file.Close()
			}
			return err
		}
		_, err = io.Copy(writer, file)
		if file != nil {
			file.Close()
		}

		return err
	})

	return err
}
