package main

import (
	"context"
	"flag"
	"strings"

	"github.com/rs/zerolog/log"
	migrate "github.com/rubenv/sql-migrate"

	"github.com/romanpitatelev/clothing-service/internal/configs"
	localfilesrepo "github.com/romanpitatelev/clothing-service/internal/repository/local-files-repo"
	filesrepo "github.com/romanpitatelev/clothing-service/internal/repository/objects-repo"
	clothesrepo "github.com/romanpitatelev/clothing-service/internal/repository/products-repo"
	"github.com/romanpitatelev/clothing-service/internal/repository/store"
	dataloader "github.com/romanpitatelev/clothing-service/internal/usecase/data-loader"
)

var (
	dir, filesStr string
	skipPhotos    bool
)

func init() {
	flag.StringVar(&dir, "d", ".", "working directory")
	flag.StringVar(&filesStr, "f", "", "comma-separated files to load. If none, all files are being processed")
	flag.BoolVar(&skipPhotos, "s", false, "skip photos")
}

const poolSize = 10

func main() {
	flag.Parse()

	cfg := configs.New(false)

	db, err := store.New(context.Background(), store.Config{Dsn: cfg.PostgresDSN})
	if err != nil {
		log.Panic().Err(err).Msg("failed to connect to database")
	}

	if err := db.Migrate(migrate.Up); err != nil {
		log.Panic().Err(err).Msg("failed to migrate")
	}

	log.Info().Msg("successful migration")

	clothesRepo := clothesrepo.New(db)

	filesRepo, err := filesrepo.New(filesrepo.S3Config{
		Address: cfg.S3Address,
		Bucket:  cfg.S3Bucket,
		Access:  cfg.S3Access,
		Secret:  cfg.S3Secret,
		Region:  cfg.S3Region,
	})
	if err != nil {
		log.Panic().Err(err).Msg("failed to connect to objects-repo")
	}

	localFilesRepo := localfilesrepo.New(dir)

	var files []string

	if filesStr != "" {
		files = strings.Split(filesStr, ",")
	}

	loader := dataloader.New(dataloader.Config{
		SkipPhotos: skipPhotos,
		PoolSize:   poolSize,
		Dir:        dir,
		Files:      files,
	},
		localFilesRepo,
		clothesRepo,
		filesRepo,
	)

	if err = loader.Run(context.Background()); err != nil {
		log.Panic().Err(err).Msg("failed to load files")
	}
}
