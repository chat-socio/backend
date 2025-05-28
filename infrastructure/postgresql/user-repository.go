package postgresql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/chat-socio/backend/internal/domain"
)

type userRepository struct {
	db *sql.DB
}

// GetListUser implements domain.UserRepository.
func (u *userRepository) GetListUser(ctx context.Context, keyword string, limit int, lastID string) ([]*domain.UserInfo, error) {
	var users []*domain.UserInfo
	var user domain.UserInfo
	fields, _ := user.MapFields()
	query := fmt.Sprintf(`SELECT %s FROM %s`, strings.Join(fields, ","), user.TableName())

	args := []any{}
	conditions := []string{}

	if keyword != "" {
		conditions = append(conditions, fmt.Sprintf("(full_name ILIKE $%d OR email ILIKE $%d)", len(args)+1, len(args)+2))
		args = append(args, "%"+keyword+"%", "%"+keyword+"%")
	}

	if lastID != "" {
		conditions = append(conditions, fmt.Sprintf("id < $%d", len(args)+1))
		args = append(args, lastID)
	}

	conditions = append(conditions, "type = 'EXTERNAL'")

	if len(conditions) > 0 {
		query = fmt.Sprintf("%s WHERE %s", query, strings.Join(conditions, " AND "))
	}

	query = fmt.Sprintf("%s ORDER BY id DESC LIMIT $%d", query, len(args)+1)
	args = append(args, limit)

	rows, err := u.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var user domain.UserInfo
		_, values := user.MapFields()
		if err := rows.Scan(values...); err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	return users, nil
}

// CreateUser implements domain.UserRepository.
func (u *userRepository) CreateUser(ctx context.Context, user *domain.UserInfo) error {
	query := `INSERT INTO user_info (id ,account_id, type, email, full_name, avatar, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	stmt, err := u.db.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, user.AccountID, user.Type, user.Email, user.FullName, user.Avatar, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		return err
	}
	return nil
}

// GetUserByAccountID implements domain.UserRepository.
func (u *userRepository) GetUserByAccountID(ctx context.Context, accountID string) (*domain.UserInfo, error) {
	var user domain.UserInfo
	fields, values := user.MapFields()
	query := fmt.Sprintf(`SELECT %s FROM %s WHERE account_id = $1`, strings.Join(fields, ","), user.TableName())
	stmt, err := u.db.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	row := stmt.QueryRowContext(ctx, accountID)
	err = row.Scan(values...)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByEmail implements domain.UserRepository.
func (u *userRepository) GetUserByEmail(ctx context.Context, email string) (*domain.UserInfo, error) {
	var user domain.UserInfo
	fields, values := user.MapFields()
	query := fmt.Sprintf(`SELECT %s FROM %s WHERE email = $1`, strings.Join(fields, ","), user.TableName())
	stmt, err := u.db.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	row := stmt.QueryRowContext(ctx, email)
	err = row.Scan(values...)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByID implements domain.UserRepository.
func (u *userRepository) GetUserByID(ctx context.Context, id string) (*domain.UserInfo, error) {
	var user domain.UserInfo
	fields, values := user.MapFields()
	query := fmt.Sprintf(`SELECT %s FROM %s WHERE id = $1`, strings.Join(fields, ","), user.TableName())
	stmt, err := u.db.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	row := stmt.QueryRowContext(ctx, id)
	err = row.Scan(values...)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUser implements domain.UserRepository.
func (u *userRepository) UpdateUser(ctx context.Context, user *domain.UserInfo) error {
	query := `UPDATE user_info SET full_name = $1, avatar = $2, updated_at = NOW() WHERE id = $3`
	stmt, err := u.db.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.ExecContext(ctx, user.FullName, user.Avatar, user.ID)
	if err != nil {
		return err
	}
	return nil
}

func NewUserRepository(db *sql.DB) domain.UserRepository {
	return &userRepository{
		db: db,
	}
}

var _ domain.UserRepository = (*userRepository)(nil)
