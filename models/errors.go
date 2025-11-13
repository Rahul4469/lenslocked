package models

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
)

var (
	ErrEmailTaken      = errors.New("models: email address is already in use")
	ErrAccountNotFound = errors.New("models: no rows in result set")
	ErrNotFound        = errors.New("models: resources could not be found")
)

type FileError struct {
	Issue string
}

func (fe FileError) Error() string {
	return fmt.Sprintf("invalid file: %v", fe.Issue)
}

func checkContentType(r io.Reader, allowedTypes []string) ([]byte, error) {
	testBytes := make([]byte, 512)
	n, err := r.Read(testBytes)
	if err != nil {
		return nil, fmt.Errorf("checking content type: %w", err)
	}

	contentType := http.DetectContentType(testBytes)
	for _, t := range allowedTypes {
		if contentType == t {
			return testBytes[:n], nil
		}
	}
	return nil, FileError{
		Issue: fmt.Sprintf("Invalid content type: %v", contentType),
	}
}

func checkExtension(filename string, allowedExtension []string) error {
	if !hasExtension(filename, allowedExtension) {
		return FileError{
			Issue: fmt.Sprintf("invalid extension: %v", filepath.Ext(filename)),
		}
	}
	return nil
}
