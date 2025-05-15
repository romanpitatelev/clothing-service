package tests

import (
	"bytes"
	"context"
	"net/http"
	"sort"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/rs/zerolog/log"

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

	s.Run("collect alphabet", func() {
		s.T().Skip()

		var lines []string

		err := pgxscan.Select(context.Background(), s.db, &lines, `
SELECT coalesce(photo_1, '') || coalesce(photo_2, '') || coalesce(photo_3, '') || coalesce(photo_4, '') ||
       coalesce(photo_5, '') || coalesce(photo_6, '') || coalesce(photo_7, '') || coalesce(photo_8, '') ||
       coalesce(photo_9, '') || coalesce(photo_10, '')
FROM variants
WHERE photo_1 IS NOT NULL`)
		s.Require().NoError(err)

		m := map[rune]struct{}{}

		for _, line := range lines {
			for _, elem := range line {
				m[elem] = struct{}{}
			}
		}

		runeSlice := make([]rune, 0, len(m))

		for k := range m {
			runeSlice = append(runeSlice, k)
		}

		sort.Slice(runeSlice, func(i, j int) bool {
			return runeSlice[i] < runeSlice[j]
		})

		log.Info().Msg(string(runeSlice))
	})
}
