package postgresql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/chat-socio/backend/internal/domain"
)

type accountRepository struct {
	db *sql.DB
}

// CreateAccount implements domain.AccountRepository.
func (a *accountRepository) CreateAccount(ctx context.Context, account *domain.Account) error {
	query := fmt.Sprintf(`INSERT INTO %s (id,username, password, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)`, account.TableName())
	stmt, err := a.db.PrepareContext(ctx, query)
	if err != nil {
		return err
	}

	defer stmt.Close()
	_, err = stmt.ExecContext(ctx, account.Username, account.Password, account.CreatedAt, account.UpdatedAt)
	if err != nil {
		return err
	}
	return nil
}

// CreateAccountUser implements domain.AccountRepository.
func (a *accountRepository) CreateAccountUser(ctx context.Context, account *domain.Account, user *domain.UserInfo) error {
	tx, err := a.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	query := fmt.Sprintf(`INSERT INTO %s (id, username, password, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)`, account.TableName())
	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, account.ID, account.Username, account.Password, account.CreatedAt, account.UpdatedAt)
	if err != nil {
		return err
	}

	query = fmt.Sprintf(`INSERT INTO %s (id,account_id, type, email, full_name, avatar, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`, user.TableName())
	stmt, err = tx.PrepareContext(ctx, query)
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, user.ID, account.ID, user.Type, user.Email, user.FullName, user.Avatar, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

// GetAccountByID implements domain.AccountRepository.
func (a *accountRepository) GetAccountByID(ctx context.Context, id string) (*domain.Account, error) {
	account := &domain.Account{}
	fields, values := account.MapFields()
	query := fmt.Sprintf(`SELECT %s FROM %s WHERE id = $1`, strings.Join(fields, ","), account.TableName())
	stmt, err := a.db.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}

	defer stmt.Close()
	row := stmt.QueryRowContext(ctx, id)
	err = row.Scan(values...)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, err
	}
	return account, nil
}

// GetAccountByUsername implements domain.AccountRepository.
func (a *accountRepository) GetAccountByUsername(ctx context.Context, username string) (*domain.Account, error) {
	account := &domain.Account{}
	fields, values := account.MapFields()
	query := fmt.Sprintf(`SELECT %s FROM %s WHERE username = $1`, strings.Join(fields, ","), account.TableName())
	stmt, err := a.db.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}

	defer stmt.Close()
	row := stmt.QueryRowContext(ctx, username)
	err = row.Scan(values...)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, err
	}
	return account, nil
}

// UpdatePassword implements domain.AccountRepository.
func (a *accountRepository) UpdatePassword(ctx context.Context, id string, password string) error {
	var temp domain.Account
	query := fmt.Sprintf(`UPDATE %s SET password = $1 WHERE id = $2`, temp.TableName())
	stmt, err := a.db.PrepareContext(ctx, query)
	if err != nil {
		return err
	}

	defer stmt.Close()
	_, err = stmt.ExecContext(ctx, password, id)
	if err != nil {
		return err
	}
	return nil
}

var _ domain.AccountRepository = (*accountRepository)(nil)

// NewAccountRepository creates a new instance of AccountRepository.
func NewAccountRepository(db *sql.DB) domain.AccountRepository {
	return &accountRepository{
		db: db,
	}
}
