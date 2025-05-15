package localfilesrepo

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/romanpitatelev/clothing-service/internal/entity"
)

type Files struct {
	workdir string
}

func New(workdir string) *Files {
	return &Files{
		workdir: workdir,
	}
}

func (f *Files) ListDir(path, ext string) ([]string, error) {
	entries, err := os.ReadDir(f.workdir + "/" + path)
	if err != nil {
		return nil, fmt.Errorf("error reading dir %s: %w", path, err)
	}

	result := make([]string, 0, len(entries))

	for _, entry := range entries {
		if entry.IsDir() || ext != "" && !strings.HasSuffix(entry.Name(), ext) {
			continue
		}

		result = append(result, entry.Name())
	}

	return result, nil
}

func (f *Files) GetFile(path string) (io.ReadSeekCloser, error) {
	file, err := os.Open(f.workdir + "/" + path) //nolint:gosec
	if err != nil {
		if os.IsNotExist(err) {
			return nil, entity.ErrFileNotFound
		}
	}

	return file, nil
}
