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

const (
	maxUpdates = 6
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

//nolint:funlen
func (r *Repo) CreateUnverifiedUser(ctx context.Context, user entity.User, otp string, otpExpiresAt time.Time) error {
	tx := r.db.GetTXFromContext(ctx)

	var (
		sb      strings.Builder
		params  []interface{}
		columns []string
		values  []string
	)

	columns = append(columns, "id", "is_verified", "created_at", "otp", "otp_expires_at")
	values = append(values,
		fmt.Sprintf("$%d", len(params)+1),
		fmt.Sprintf("$%d", len(params)+2), //nolint:mnd
		fmt.Sprintf("$%d", len(params)+3), //nolint:mnd
		fmt.Sprintf("$%d", len(params)+4), //nolint:mnd
		fmt.Sprintf("$%d", len(params)+5), //nolint:mnd
	)

	params = append(params,
		user.UserID,
		false,
		time.Now(),
		otp,
		otpExpiresAt,
	)

	if user.FirstName != nil {
		columns = append(columns, "first_name")
		values = append(values, fmt.Sprintf("$%d", len(params)+1))
		params = append(params, *user.FirstName)
	}

	if user.LastName != nil {
		columns = append(columns, "last_name")
		values = append(values, fmt.Sprintf("$%d", len(params)+1))
		params = append(params, *user.LastName)
	}

	if user.NickName != "" {
		columns = append(columns, "nick_name")
		values = append(values, fmt.Sprintf("$%d", len(params)+1))
		params = append(params, user.NickName)
	}

	if user.Gender != nil {
		columns = append(columns, "gender")
		values = append(values, fmt.Sprintf("$%d", len(params)+1))
		params = append(params, *user.Gender)
	}

	if user.Age != nil {
		columns = append(columns, "age")
		values = append(values, fmt.Sprintf("$%d", len(params)+1))
		params = append(params, *user.Age)
	}

	if user.Email != nil {
		columns = append(columns, "email")
		values = append(values, fmt.Sprintf("$%d", len(params)+1))
		params = append(params, *user.Email)
	}

	if user.Phone != nil {
		columns = append(columns, "phone")
		values = append(values, fmt.Sprintf("$%d", len(params)+1))
		params = append(params, *user.Phone)
	}

	sb.WriteString("INSERT INTO users (")
	sb.WriteString(strings.Join(columns, ", "))
	sb.WriteString(") VALUES (")
	sb.WriteString(strings.Join(values, ", "))
	sb.WriteString(")")

	_, err := tx.Exec(ctx, sb.String(), params...)
	if err != nil {
		return fmt.Errorf("failed to create unverified user: %w", err)
	}

	return nil
}

func (r *Repo) VerifyUserWithOTP(ctx context.Context, unverifiedUser entity.User) error {
	tx := r.db.GetTXFromContext(ctx)

	if unverifiedUser.OTPExpiresAt.Before(time.Now()) {
		return entity.ErrOTPExpired
	}

	query := `
UPDATE users
SET is_verified = true
WHERE TRUE
	AND id = $1
	AND otp = $2
	AND otp_expires_at > NOW()`

	result, err := tx.Exec(ctx, query, unverifiedUser.UserID, unverifiedUser.OTP, unverifiedUser.OTPExpiresAt)
	if err != nil {
		return fmt.Errorf("error verifying user %s: %w", uuid.UUID(unverifiedUser.UserID), err)
	}

	if result.RowsAffected() == 0 {
		return entity.ErrInvalidOTP
	}

	return nil
}

func (r *Repo) GetUser(ctx context.Context, userID entity.UserID) (entity.User, error) {
	var user entity.User

	query := `
SELECT id, first_name, last_name, nick_name, gender, age, email, phone, created_at, is_verified, updated_at, deleted_at
FROM users
WHERE TRUE
	AND id = $1
	AND deleted_at IS NULL`

	var db store.Transaction

	db = r.db.GetTXFromContext(ctx)

	if db == nil {
		db = r.db
	}

	err := db.QueryRow(ctx, query, userID).Scan(
		&user.UserID,
		&user.FirstName,
		&user.LastName,
		&user.NickName,
		&user.Gender,
		&user.Age,
		&user.Email,
		&user.Phone,
		&user.CreatedAt,
		&user.IsVerified,
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

//nolint:funlen
func (r *Repo) UpdateUser(ctx context.Context, userID entity.UserID, updatedUser entity.UserUpdate) (entity.User, error) {
	tx := r.db.GetTXFromContext(ctx)

	var (
		sb     strings.Builder
		params []interface{}
	)

	paramCount := 1
	updates := make([]string, 0, maxUpdates)

	sb.WriteString("UPDATE users SET ")

	if updatedUser.FirstName != nil {
		updates = append(updates, fmt.Sprintf("first_name = $%d", paramCount))
		params = append(params, *updatedUser.FirstName)
		paramCount++
	}

	if updatedUser.LastName != nil {
		updates = append(updates, fmt.Sprintf("last_name = $%d", paramCount))
		params = append(params, *updatedUser.LastName)
		paramCount++
	}

	if updatedUser.NickName != nil {
		updates = append(updates, fmt.Sprintf("nick_name = $%d", paramCount))
		params = append(params, *updatedUser.NickName)
		paramCount++
	}

	if updatedUser.Email != nil {
		updates = append(updates, fmt.Sprintf("email = $%d", paramCount))
		params = append(params, *updatedUser.Email)
		paramCount++
	}

	if updatedUser.Phone != nil {
		updates = append(updates, fmt.Sprintf("phone = $%d", paramCount))
		params = append(params, *updatedUser.Phone)
		paramCount++
	}

	if len(updates) == 0 {
		return r.GetUser(ctx, userID)
	}

	updates = append(updates, fmt.Sprintf("updated_at = $%d", paramCount))
	params = append(params, time.Now())
	paramCount++

	sb.WriteString(strings.Join(updates, ", "))

	sb.WriteString(fmt.Sprintf(" WHERE id = $%d AND deleted_at IS NULL", paramCount))

	params = append(params, userID)

	sb.WriteString(" RETURNING id, first_name, last_name, nick_name, gender, age, email, created_at, updated_at, deleted_at")

	row := tx.QueryRow(ctx, sb.String(), params...)

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
