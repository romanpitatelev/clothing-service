package usersrepo

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/romanpitatelev/clothing-service/internal/entity"
	"github.com/romanpitatelev/clothing-service/internal/repository/store"
)

type database interface {
	Exec(ctx context.Context, sq string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sq string, arguments ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, arguments ...any) pgx.Row
	GetTXFromContext(ctx context.Context) store.Transaction
}

type Repo struct {
	db database
}

func New(db database) *Repo {
	return &Repo{
		db: db,
	}
}

func (r *Repo) CreateUser(ctx context.Context, user entity.User) (entity.User, error) {}

func (r *Repo) GetUser(ctx context.Context, userID entity.UserID) (entity.User, error) {
	var user entity.User

	query := `
SELECT id, first_name, last_name, nick_name, gender, age, email, created_at, updated_at, deleted_at
FROM users
WHERE TRUE
	AND id = $1
	AND deleted_at IS NULL`

	var db store.Transaction

	db = r.db.GetTXFromContext(ctx)

	if db == nil {
		db = r.db
	}

	err := r.db.QueryRow(ctx, query, userID).Scan(
		&user.UserID,
		&user.FirstName,
		&user.LastName,
		&user.NickName,
		&user.Gender,
		&user.Age,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.User{}, entity.ErrUserNotFound
		}

		return entity.User{}, fmt.Errorf("failed to get user info: %w", err)
	}

	return user, nil
}

func (r *Repo) UpdateUser(ctx context.Context, userID entity.UserID, updatedUser entity.UserUpdate) (entity.User, error) {
	tx := r.db.GetTXFromContext(ctx)

	query := `
UPDATE users
SET first_name = $1, last_name = $2, nick_name = $3, email = $4, updated_at = $5
WHERE TRUE
	AND id = $5
	AND deleted_at IS NULL
RETURNING id, first_name, last_name, nick_name, gender, age, email, created_at, updated_at, deleted_at`

	updatedAt := time.Now()

	row := tx.QueryRow(ctx, query,
		updatedUser.FirstName,
		updatedUser.LastName,
		updatedUser.NickName,
		updatedUser.Email,
		updatedAt,
		userID,
	)

	var user entity.User
	err := row.Scan(
		&user.UserID,
		&user.FirstName,
		&user.LastName,
		&user.NickName,
		&user.Gender,
		&user.Age,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.User{}, entity.ErrUserNotFound
		}

		return entity.User{}, fmt.Errorf("failed to get user info: %w", err)
	}

	return user, nil
}

func (r *Repo) DeleteUser(ctx context.Context, userID entity.UserID) error {
	_, err := r.GetUser(ctx, userID)
	if err != nil {
		if errors.Is(err, entity.ErrUserNotFound) {
			return entity.ErrUserNotFound
		}

		return fmt.Errorf("failed to fetch user in DeleteUser() function: %w", err)
	}

	query := `
UPDATE users
SET deleted_at = NOW()
WHERE TRUE
	AND id = $1
	AND deleted_at IS NULL`

	_, err = r.db.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("error deleting user %s: %w", uuid.UUID(userID), err)
	}

	return nil
}

func (r *Repo) GetUsers(ctx context.Context, request entity.GetRequestParams) ([]entity.User, error) {
	var (
		users []entity.User
		rows  pgx.Rows
		err   error
	)

	query, args := r.getUsersQuery(request)

	if rows, err = r.db.Query(ctx, query, args...); err != nil {
		return nil, fmt.Errorf("error getting users info: %w", err)
	}

	defer rows.Close()

	for rows.Next() {
		var user entity.User

		err = rows.Scan(
			&user.UserID,
			&user.FirstName,
			&user.LastName,
			&user.NickName,
			&user.Gender,
			&user.Age,
			&user.Email,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error when scanning user: %w", err)
		}

		users = append(users, user)

		if err = rows.Err(); err != nil {
			return nil, fmt.Errorf("rows.Err(): %w", err)
		}

		if len(users) == 0 {
			return []entity.User{}, nil
		}
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err(): %w", err)
	}

	if len(users) == 0 {
		return []entity.User{}, nil
	}

	return users, nil
}

func (r *Repo) getUsersQuery(request entity.GetRequestParams) (string, []any) {
	var (
		sb              strings.Builder
		args            []any
		validSortParams = map[string]string{
			"last_name":  "last_name",
			"first_name": "first_name",
			"nick_name":  "nick_name",
			"age":        "age",
		}
	)

	sb.WriteString(`SELECT id, first_name, last_name, nick_name, gender, age, email, created_at, updated_at
							FROM users
							WHERE deleted_at IS NULL`)

	if request.Filter != "" {
		args = append(args, "%"+request.Filter+"%")
		sb.WriteString(fmt.Sprintf(` AND concat_ws('', id, first_name, last_name, nick_name, gender, age, email, created_at, updated_at)
ILIKE $%d`, len(args)))
	}

	sorting, ok := validSortParams[request.Sorting]
	if !ok {
		sorting = "last_name"
	}

	sb.WriteString(" ORDER BY " + sorting)

	if request.Descending {
		sb.WriteString(" DESC")
	}

	args = append(args, request.Limit)

	sb.WriteString(fmt.Sprintf(" LIMIT $%d", len(args)))

	if request.Offset > 0 {
		args = append(args, request.Offset)
		sb.WriteString(fmt.Sprintf(" OFFSET $%d", len(args)))
	}

	return sb.String(), args
}
