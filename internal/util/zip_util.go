package util

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

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
			return err
		}
		defer fileReader.Close()

		// Extract regular file since not a directory
		fmt.Println("Extracting file:", file.Name)

		// Open an output file for writing
		targetFile, err := os.OpenFile(
			path,
			os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
			file.Mode(),
		)
		if err != nil {
			return err
		}
		defer targetFile.Close()

		if _, err = io.Copy(targetFile, fileReader); err != nil {
			return err
		}
	}

	return nil
}

func UnzipFileWithoutDefers(archive, target string) (err error) {
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

		// Extract regular file since not a directory
		fmt.Println("Extracting file:", file.Name)

		// Open an output file for writing
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

	var baseDir string
	if info.IsDir() {
		baseDir = filepath.Base(source)
		fmt.Printf("baseDir: %s\n", baseDir)
	}

	filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		fmt.Printf("Path: %s, Info: %v or Error: %v\n", path, info, err)
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		fmt.Printf("header: %v or Error: %v\n", header, err)
		if err != nil {
			return err
		}

		if baseDir != "" {
			header.Name = filepath.Join("", strings.TrimPrefix(path, source))
			fmt.Printf("header.Name: %s\n", header.Name)
		}

		if info.IsDir() {
			header.Name += "/"
			fmt.Printf("header.Name is Dir: %s\n", header.Name)
		} else {
			header.Method = zip.Deflate
		}

		if header.Name != "/" {
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
		}

		return err
	})

	return err
}
