package tests

import (
	"net/http"

	"github.com/romanpitatelev/clothing-service/internal/entity"
)

func (s *IntegrationTestSuite) TestFiles() {
	s.Run("create file on repo layer", func() {
		file, contentType, name := s.getFile(http.MethodGet, imagesPath+"/test_file", http.StatusOK, nil, entity.User{})
		s.Require().Equal("application/octet-stream", contentType)
		s.Require().Equal("test_file", name)
		s.Require().Equal("test data\n", string(file))
	})
}
