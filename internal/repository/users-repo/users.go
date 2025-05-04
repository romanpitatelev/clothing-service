package usersrepo

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgerrcode"
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

func (r *Repo) CreateUnverifiedUser(ctx context.Context, user entity.User, otp string) (entity.User, error) { //nolint:funlen
	db := r.db.GetTXFromContext(ctx)

	var (
		sb      strings.Builder
		args    []interface{}
		columns []string
		values  []string
	)

	columns = append(columns, "id")
	args = append(args, user.ID)
	values = append(values, fmt.Sprintf("$%d", len(args)))

	columns = append(columns, "otp")
	args = append(args, otp)
	values = append(values, fmt.Sprintf("$%d", len(args)))

	columns = append(columns, "otp_created_at")
	args = append(args, time.Now())
	values = append(values, fmt.Sprintf("$%d", len(args)))

	columns = append(columns, "phone")
	args = append(args, user.Phone)
	values = append(values, fmt.Sprintf("$%d", len(args)))

	if user.NickName == "" {
		user.NickName = user.Phone
	}

	columns = append(columns, "nick_name")
	args = append(args, user.NickName)
	values = append(values, fmt.Sprintf("$%d", len(args)))

	if user.FirstName != nil {
		columns = append(columns, "first_name")
		args = append(args, *user.FirstName)
		values = append(values, fmt.Sprintf("$%d", len(args)))
	}

	if user.LastName != nil {
		columns = append(columns, "last_name")
		args = append(args, *user.LastName)
		values = append(values, fmt.Sprintf("$%d", len(args)))
	}

	if user.Gender != nil {
		columns = append(columns, "gender")
		args = append(args, *user.Gender)
		values = append(values, fmt.Sprintf("$%d", len(args)))
	}

	if user.BirthDate != nil {
		columns = append(columns, "birth_date")
		args = append(args, *user.BirthDate)
		values = append(values, fmt.Sprintf("$%d", len(args)))
	}

	if user.Email != nil {
		columns = append(columns, "email")
		args = append(args, *user.Email)
		values = append(values, fmt.Sprintf("$%d", len(args)))
	}

	sb.WriteString("INSERT INTO users (")
	sb.WriteString(strings.Join(columns, ", "))
	sb.WriteString(") VALUES (")
	sb.WriteString(strings.Join(values, ", "))
	sb.WriteString(`)
RETURNING id, first_name, last_name, nick_name, gender, birth_date, email, email_verified, phone, phone_verified, created_at, updated_at`)

	if err := pgxscan.Get(ctx, db, &user, sb.String(), args...); err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return entity.User{}, entity.ErrDuplicateContact
		}

		return entity.User{}, fmt.Errorf("failed to create unverified user: %w", err)
	}

	return user, nil
}

func (r *Repo) VerifyUserWithOTP(ctx context.Context, validateUserRequest entity.ValidateUserRequest, otpLifetime time.Duration) (entity.User, error) {
	db := r.db.GetTXFromContext(ctx)

	query := `
UPDATE users
SET phone_verified = true
WHERE TRUE
	AND id = $1
	AND otp = $2
	AND otp_created_at > NOW() - INTERVAL '1 SECOND'*$3
	RETURNING id, first_name, last_name, nick_name, gender, birth_date, email, email_verified, phone, phone_verified, created_at, updated_at`

	var user entity.User

	if err := pgxscan.Get(ctx, db, &user, query, validateUserRequest.UserID.String(), validateUserRequest.OTP, otpLifetime.Seconds()); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.User{}, entity.ErrInvalidOTP
		}

		return entity.User{}, fmt.Errorf("error verifying user %s: %w", uuid.UUID(validateUserRequest.UserID), err)
	}

	return user, nil
}

func (r *Repo) GetUser(ctx context.Context, userID entity.UserID) (entity.User, error) {
	db := r.db.GetTXFromContext(ctx)

	var user entity.User

	query := `
SELECT id, first_name, last_name, nick_name, gender, birth_date, email, email_verified, phone, phone_verified, created_at, updated_at
FROM users
WHERE TRUE
	AND id = $1
	AND deleted_at IS NULL`

	if err := pgxscan.Get(ctx, db, &user, query, userID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.User{}, entity.ErrUserNotFound
		}

		return entity.User{}, fmt.Errorf("failed to get user info: %w", err)
	}

	return user, nil
}

func (r *Repo) UpdateUser(ctx context.Context, userID entity.UserID, updatedUser entity.UserUpdate) (entity.User, error) { //nolint:funlen
	db := r.db.GetTXFromContext(ctx)

	var (
		sb     strings.Builder
		params []interface{}
	)

	updates := make([]string, 0, maxUpdates)

	sb.WriteString("UPDATE users SET ")

	if updatedUser.FirstName != nil {
		params = append(params, *updatedUser.FirstName)
		updates = append(updates, fmt.Sprintf("first_name = $%d", len(params)))
	}

	if updatedUser.LastName != nil {
		params = append(params, *updatedUser.LastName)
		updates = append(updates, fmt.Sprintf("last_name = $%d", len(params)))
	}

	if updatedUser.NickName != nil {
		params = append(params, *updatedUser.NickName)
		updates = append(updates, fmt.Sprintf("nick_name = $%d", len(params)))
	}

	if updatedUser.Gender != nil {
		params = append(params, *updatedUser.Gender)
		updates = append(updates, fmt.Sprintf("gender = $%d", len(params)))
	}

	if updatedUser.BirthDate != nil {
		params = append(params, *updatedUser.BirthDate)
		updates = append(updates, fmt.Sprintf("birth_date = $%d", len(params)))
	}

	if updatedUser.Email != nil {
		params = append(params, *updatedUser.Email)
		updates = append(updates, fmt.Sprintf("email = $%d", len(params)))
	}

	if updatedUser.EmailVerified != nil {
		params = append(params, *updatedUser.EmailVerified)
		updates = append(updates, fmt.Sprintf("email_verified = $%d", len(params)))
	}

	if updatedUser.Phone != nil {
		params = append(params, *updatedUser.Phone)
		updates = append(updates, fmt.Sprintf("phone = $%d", len(params)))
	}

	if updatedUser.PhoneVerified != nil {
		params = append(params, *updatedUser.PhoneVerified)
		updates = append(updates, fmt.Sprintf("phone_verified = $%d", len(params)))
	}

	if updatedUser.OTP != nil {
		params = append(params, *updatedUser.OTP)
		updates = append(updates, fmt.Sprintf("otp = $%d", len(params)))
	}

	if updatedUser.OTPCreatedAt != nil {
		params = append(params, *updatedUser.OTPCreatedAt)
		updates = append(updates, fmt.Sprintf("otp_created_at = $%d", len(params)))
	}

	if len(updates) == 0 {
		return r.GetUser(ctx, userID)
	}

	params = append(params, time.Now())
	updates = append(updates, fmt.Sprintf("updated_at = $%d", len(params)))

	sb.WriteString(strings.Join(updates, ", "))

	params = append(params, userID)
	sb.WriteString(fmt.Sprintf(" WHERE id = $%d AND deleted_at IS NULL", len(params)))

	sb.WriteString(" RETURNING id, first_name, last_name, nick_name, gender, birth_date, email, email_verified, phone, phone_verified, created_at, updated_at")

	var user entity.User

	if err := pgxscan.Get(ctx, db, &user, sb.String(), params...); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.User{}, entity.ErrUserNotFound
		}

		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return entity.User{}, entity.ErrDuplicateContact
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
