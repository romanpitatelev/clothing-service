package usersrepo

import (
	"context"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/romanpitatelev/clothing-service/internal/entity"
)

type database interface {
	Exec(ctx context.Context, sq string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sq string, arguments ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, arguments ...any) pgx.Row
}

type Repo struct {
	db database
}

func New(db database) *Repo {
	return &Repo{
		db: db,
	}
}

func (r *Repo) CreateUser(ctx context.Context, user entity.User) (entity.User, error)  {}
func (r *Repo) GetUser(ctx context.Context, userID entity.UserID) (entity.User, error) {}
func (r *Repo) UpdateUser(ctx context.Context, userID entity.UserID, updatedUser entity.UserUpdate) (entity.User, error) {
}
func (r *Repo) DeleteUser(ctx context.Context, userID entity.User) error                            {}
func (r *Repo) GetUsers(ctx context.Context, request entity.GetUsersRequest) ([]entity.User, error) {}
