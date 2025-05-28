package postgresql

import (
	"context"
	"database/sql"
	"time"

	"github.com/chat-socio/backend/internal/domain"
)

type sessionRepository struct {
	db *sql.DB
}

// CreateSession implements domain.SessionRepository.
func (s *sessionRepository) CreateSession(ctx context.Context, session *domain.Session) error {
	query := `INSERT INTO session (session_token, account_id, created_at, updated_at, expired_at, is_active, user_agent, ip_address) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	stmt, err := s.db.PrepareContext(ctx, query)
	if err != nil {
		return err
	}

	defer stmt.Close()
	_, err = stmt.ExecContext(ctx, session.SessionToken, session.AccountID, session.CreatedAt, session.UpdatedAt, session.ExpiredAt, session.IsActive, session.UserAgent, session.IPAddress)
	if err != nil {
		return err
	}

	return nil
}

// DeactivateSession implements domain.SessionRepository.
func (s *sessionRepository) DeactivateSession(ctx context.Context, token string) error {
	query := `UPDATE session SET is_active = false WHERE session_token = $1`
	stmt, err := s.db.PrepareContext(ctx, query)
	if err != nil {
		return err
	}

	defer stmt.Close()
	_, err = stmt.ExecContext(ctx, token)
	if err != nil {
		return err
	}

	return nil
}

// DeactiveAllSessionByAccountID implements domain.SessionRepository.
func (s *sessionRepository) DeactiveAllSessionByAccountID(ctx context.Context, accountID string) error {
	query := `UPDATE session SET is_active = false WHERE account_id = $1`
	stmt, err := s.db.PrepareContext(ctx, query)
	if err != nil {
		return err
	}

	defer stmt.Close()
	_, err = stmt.ExecContext(ctx, accountID)
	if err != nil {
		return err
	}

	return nil
}

// GetListSessionByAccountID implements domain.SessionRepository.
func (s *sessionRepository) GetListSessionByAccountID(ctx context.Context, accountID string) ([]*domain.Session, error) {
	var sessions []*domain.Session
	query := `SELECT session_token, account_id, created_at, updated_at, expired_at, is_active, user_agent, ip_address FROM session WHERE account_id = $1 AND is_active = true`
	stmt, err := s.db.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}

	defer stmt.Close()
	rows, err := stmt.QueryContext(ctx, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var session domain.Session
		err = rows.Scan(&session.SessionToken, &session.AccountID, &session.CreatedAt, &session.UpdatedAt, &session.ExpiredAt, &session.IsActive, &session.UserAgent, &session.IPAddress)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, &session)
	}

	return sessions, nil
}

// GetSessionByToken implements domain.SessionRepository.
func (s *sessionRepository) GetSessionByToken(ctx context.Context, token string) (*domain.Session, error) {
	query := `SELECT session_token, account_id, created_at, updated_at, expired_at, is_active, user_agent, ip_address FROM session WHERE session_token = $1`
	stmt, err := s.db.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}

	defer stmt.Close()
	row := stmt.QueryRowContext(ctx, token)
	var session domain.Session
	err = row.Scan(&session.SessionToken, &session.AccountID, &session.CreatedAt, &session.UpdatedAt, &session.ExpiredAt, &session.IsActive, &session.UserAgent, &session.IPAddress)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, err
	}

	return &session, nil
}

// UpdateExpiredAt implements domain.SessionRepository.
func (s *sessionRepository) UpdateExpiredAt(ctx context.Context, token string, newExpiredAt *time.Time) error {
	query := `UPDATE session SET expired_at = $1 WHERE session_token = $2`
	stmt, err := s.db.PrepareContext(ctx, query)
	if err != nil {
		return err
	}

	defer stmt.Close()
	_, err = stmt.ExecContext(ctx, newExpiredAt, token)
	if err != nil {
		return err
	}

	return nil
}

func NewSessionRepository(db *sql.DB) *sessionRepository {
	return &sessionRepository{
		db: db,
	}
}

var _ domain.SessionRepository = (*sessionRepository)(nil)
