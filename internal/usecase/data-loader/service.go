package dataloader

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/google/uuid"
	"github.com/romanpitatelev/clothing-service/internal/entity"
	"github.com/romanpitatelev/clothing-service/internal/utils"
	"github.com/rs/zerolog/log"
)

type Config struct {
	SkipPhotos bool
	PoolSize   int
	Dir        string
	Files      []string
}

type files interface {
	ListDir(path string) ([]string, error)
	GetFile(path string) (io.ReadSeekCloser, error)
}

type productsStore interface {
	UpsertBrand(ctx context.Context, brand entity.Brand) (entity.Brand, error)
	UpsertProduct(ctx context.Context, product entity.Product) (entity.Product, error)
	UpsertVariant(ctx context.Context, variant entity.Variant) error
	SetRelatedPhotos(ctx context.Context) error
}

type mediaStore interface {
	UploadReader(data io.ReadSeeker, fileName string) error
}

type Service struct {
	cfg           Config
	files         files
	productsStore productsStore
	mediaStore    mediaStore
	dataPool      *utils.WorkerPool
	mediaPool     *utils.WorkerPool
}

func New(cfg Config, files files, productsStore productsStore, mediaStore mediaStore) *Service {
	return &Service{
		cfg:           cfg,
		files:         files,
		productsStore: productsStore,
		mediaStore:    mediaStore,
		dataPool:      utils.NewPool(cfg.PoolSize),
		mediaPool:     utils.NewPool(cfg.PoolSize),
	}
}

func (s *Service) Run(ctx context.Context) error {
	defer s.dataPool.Close()
	defer s.mediaPool.Close()

	var err error

	if len(s.cfg.Files) == 0 {
		s.cfg.Files, err = s.files.ListDir(s.cfg.Dir)
		if err != nil {
			return err
		}
	}

	for _, file := range s.cfg.Files {
		log.Info().Str("file", file).Msg("loading file")

		if err = s.ProcessFile(ctx, file); err != nil {
			return err
		}
	}

	if err = s.productsStore.SetRelatedPhotos(ctx); err != nil {
		return fmt.Errorf("failed to set relatedPhotos: %w", err)
	}

	return nil
}

func (s *Service) ProcessFile(ctx context.Context, fileName string) error {
	file, err := s.files.GetFile(fileName)
	if err != nil {
		return fmt.Errorf("failed to get file %s: %w", fileName, err)
	}

	defer func() {
		if err := file.Close(); err != nil {
			log.Warn().Err(err).Str("file", fileName).Msg("failed to close file")
		}
	}()

	var data []entity.JSONProduct

	if err = json.NewDecoder(file).Decode(&data); err != nil {
		return fmt.Errorf("failed to decode file %s: %w", fileName, err)
	}

	if len(data) == 0 || len(data[0].Variants) == 0 {
		log.Info().Str("file", fileName).Msg("file has no variants")

		return nil
	}

	var mapping entity.JSONImageMapping

	mappingFileName, dirName := mappingFileFromJSONFileName(fileName)

	if !s.cfg.SkipPhotos {
		mappingFile, err := s.files.GetFile(mappingFileName)
		if err != nil {
			return fmt.Errorf("failed to get file %s: %w", fileName, err)
		}

		defer func() {
			if err := mappingFile.Close(); err != nil {
				log.Warn().Err(err).Str("file", fileName).Msg("failed to close mapping file")
			}
		}()

		if err = json.NewDecoder(mappingFile).Decode(&mapping); err != nil {
			return fmt.Errorf("failed to decode mapping file %s: %w", fileName, err)
		}
	}

	brand, err := s.productsStore.UpsertBrand(ctx, entity.Brand{
		ID:   entity.BrandID(uuid.New()),
		Name: data[0].Variants[0].BrandName,
	})
	if err != nil {
		return fmt.Errorf("failed to upsert brand: %w", err)
	}

	for i := range data {
		if err = s.saveItem(ctx, data[i], brand.ID, mapping, dirName); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) saveItem(
	ctx context.Context,
	item entity.JSONProduct,
	brandID entity.BrandID,
	mapping entity.JSONImageMapping,
	dirName string,
) error {
	product, variants := item.Products()
	product.BrandID = brandID

	product, err := s.productsStore.UpsertProduct(ctx, product)
	if err != nil {
		log.Warn().Err(err).Msg("failed to upsert product")
	}

	for _, variant := range variants {
		variant.ProductID = product.ID

		s.dataPool.Put(func() {
			if err := s.productsStore.UpsertVariant(ctx, variant); err != nil {
				log.Warn().Err(err).Msg("failed to upsert variant")
			}
		})

		if s.cfg.SkipPhotos {
			continue
		}

		for i := range variant.Photos {
			s.mediaPool.Put(func() {
				photoName, ok := mapping[variant.PhotoOriginalURL[i]]
				if !ok {
					log.Warn().Msg("photo not found in mapping file")

					return
				}

				file, err := s.files.GetFile(dirName + photoName.Path)
				if err != nil {
					log.Warn().Err(err).Msg("failed to get media file")

					return
				}

				defer func() {
					if err := file.Close(); err != nil {
						log.Warn().Err(err).Msg("failed to close file")
					}
				}()

				if err := s.mediaStore.UploadReader(file, variant.Photos[i]); err != nil {
					log.Warn().Err(err).Str("name", variant.Photos[i]).Msg("failed to upload reader")
				}
			})
		}
	}

	return nil
}

func mappingFileFromJSONFileName(fileName string) (string, string) {
	cutName, ok := strings.CutSuffix(fileName, ".json")
	if !ok {
		log.Panic().Str("file", fileName).Msg("name doesnt contain .json")
	}

	return cutName + "-images/map.json", cutName + "-images/"
}
