package tests

import (
	"bytes"
	"net/http"

	"github.com/romanpitatelev/clothing-service/internal/entity"
)

func (s *IntegrationTestSuite) TestFiles() {
	s.Run("get file from object storage", func() {
		file, contentType, name := s.getFile(http.MethodGet, imagesPath+"/test_file", http.StatusOK, nil, entity.User{})
		s.Require().Equal("application/octet-stream", contentType)
		s.Require().Equal("test_file", name)
		s.Require().Equal("test data\n", string(file))
	})

	s.Run("repo layer put object", func() {
		err := s.filesRepo.UploadFile([]byte("biba"), "boba", "")
		s.Require().NoError(err)
	})

	s.Run("repo layer put reader", func() {
		err := s.filesRepo.UploadReader(bytes.NewReader([]byte("biba")),
			"ANNA_PEKUN/platia/Zheltyi/Zheltyi/Plate_SANDY/1.webp")
		s.Require().NoError(err)
	})
}
