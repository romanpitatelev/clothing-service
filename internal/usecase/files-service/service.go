package filesservice

import (
	"fmt"
	"io"
)

type filesStore interface {
	DownloadFile(fileName string) (io.ReadCloser, string, error)
}

type Service struct {
	fileStore filesStore
}

func New(fileStore filesStore) *Service {
	return &Service{
		fileStore: fileStore,
	}
}

func (s *Service) GetFile(fileName string) (io.ReadCloser, string, error) {
	reader, contentType, err := s.fileStore.DownloadFile(fileName)
	if err != nil {
		return nil, "", fmt.Errorf("error downloading file %s: %w", fileName, err)
	}

	return reader, contentType, nil
}
