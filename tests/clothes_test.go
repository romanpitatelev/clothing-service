package tests

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/romanpitatelev/clothing-service/internal/entity"
	"github.com/romanpitatelev/clothing-service/internal/utils"
)

func (s *IntegrationTestSuite) TestClothing() {
	s.Run("get clothing broken uuid", func() {
		s.sendRequest(http.MethodGet, clothesPath+"/16763be4-6022-406e-a950-fcd5", http.StatusBadRequest, nil, nil, entity.User{})
	})

	s.Run("get absent clothing", func() {
		s.sendRequest(http.MethodGet, clothesPath+"/16763be4-6022-406e-a950-fcd5018633ca", http.StatusNotFound, nil, nil, entity.User{})
	})

	id, err := uuid.Parse("16763be4-6022-406e-a950-fcd5018633ca")
	s.Require().NoError(err)

	clothing := entity.Clothing{
		ID:           entity.ClothingID(id),
		Brand:        "12 storeez",
		Category:     "Толстовки и худи",
		Color:        "Серый",
		Name:         "Худи меланж",
		Price:        17000,
		Sizes:        "хs, s, m, l",
		ProductURL:   "https://12storeez.com/catalog/sweatshirts/womencollection/hudi-melanz-128504",
		PrimaryPhoto: "https://image.12storeez.com/images/800xP_90_out/uploads/images/september/128504/66f5650365f8a-16-09-2024-katya-21033.jpg",
		ExtraPhotos: []*string{
			utils.Pointer("https://image.12storeez.com/images/800xP_90_out/uploads/images/september/128504/66f5650365f8a-16-09-2024-katya-21033.jpg"),
			utils.Pointer("https://image.12storeez.com/images/800xP_90_out/uploads/images/september/128504/66f5650365f8a-16-09-2024-katya-21033.jpg"),
			utils.Pointer("https://image.12storeez.com/images/800xP_90_out/uploads/images/september/128504/66f5650365f8a-16-09-2024-katya-21033.jpg"),
		},
		RelatedProducts: []entity.RelatedProduct{
			{
				URL:   utils.Pointer("https://12storeez.com/catalog/sweatshirts/womencollection/hudi-melanz-128504"),
				Photo: utils.Pointer("https://image.12storeez.com/images/800xP_90_out/uploads/images/september/128504/66f5650365f8a-16-09-2024-katya-21033.jpg"),
			},
			{
				URL:   utils.Pointer("https://12storeez.com/catalog/sweatshirts/womencollection/hudi-melanz-128504"),
				Photo: utils.Pointer("https://image.12storeez.com/images/800xP_90_out/uploads/images/september/128504/66f5650365f8a-16-09-2024-katya-21033.jpg"),
			},
			{
				URL: utils.Pointer("https://12storeez.com/catalog/sweatshirts/womencollection/hudi-melanz-128504"),
			},
		},
	}

	s.Run("test repo layer get absent", func() {
		_, err := s.clothesRepo.GetClothing(s.T().Context(), clothing.ID)
		s.Require().ErrorIs(err, entity.ErrClothingNotFound)
	})

	s.Run("test repo layer create", func() {
		err = s.clothesRepo.UpsertClothing(s.T().Context(), clothing)
		s.Require().NoError(err)
	})

	var createdClothing entity.Clothing

	s.Run("test repo layer get", func() {
		createdClothing, err = s.clothesRepo.GetClothing(s.T().Context(), clothing.ID)
		s.Require().NoError(err)
		s.Require().Equal(clothing.ID, createdClothing.ID)
		s.Require().Equal(clothing.Brand, createdClothing.Brand)
		s.Require().Equal(clothing.Category, createdClothing.Category)
		s.Require().Equal(clothing.Color, createdClothing.Color)
		s.Require().Equal(clothing.Name, createdClothing.Name)
		s.Require().Equal(clothing.Price, createdClothing.Price)
		s.Require().Equal(clothing.Sizes, createdClothing.Sizes)
		s.Require().Equal(clothing.ProductURL, createdClothing.ProductURL)
		s.Require().Equal(clothing.PrimaryPhoto, createdClothing.PrimaryPhoto)
		s.Require().Equal(len(clothing.ExtraPhotos), len(createdClothing.ExtraPhotos))
		s.Require().Equal(len(clothing.RelatedProducts), len(createdClothing.RelatedProducts))
	})

	s.Run("get successfully", func() {
		s.sendRequest(http.MethodGet, clothesPath+"/16763be4-6022-406e-a950-fcd5018633ca", http.StatusOK, nil, &createdClothing, entity.User{})
		s.Require().NoError(err)
		s.Require().Equal(clothing.ID, createdClothing.ID)
		s.Require().Equal(clothing.Brand, createdClothing.Brand)
		s.Require().Equal(clothing.Category, createdClothing.Category)
		s.Require().Equal(clothing.Color, createdClothing.Color)
		s.Require().Equal(clothing.Name, createdClothing.Name)
		s.Require().Equal(clothing.Price, createdClothing.Price)
		s.Require().Equal(clothing.Sizes, createdClothing.Sizes)
		s.Require().Equal(clothing.ProductURL, createdClothing.ProductURL)
		s.Require().Equal(clothing.PrimaryPhoto, createdClothing.PrimaryPhoto)
		s.Require().Equal(len(clothing.ExtraPhotos), len(createdClothing.ExtraPhotos))
		s.Require().Equal(len(clothing.RelatedProducts), len(createdClothing.RelatedProducts))
	})
}
