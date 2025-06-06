package domain

import (
	"fmt"
	"time"
)

type Session struct {
	SessionToken string     `json:"session_token,omitempty"` // UUID, generated by the server, primary key
	AccountID    string     `json:"account_id,omitempty"`
	CreatedAt    *time.Time `json:"created_at,omitempty"`
	UpdatedAt    *time.Time `json:"updated_at,omitempty"`
	ExpiredAt    *time.Time `json:"expired_at,omitempty"`
	IsActive     *bool      `json:"is_active,omitempty"`  // default true
	UserAgent    string     `json:"user_agent,omitempty"` // for tracking purposes
	IPAddress    string     `json:"ip_address,omitempty"` // for tracking purposes
	// Location     string `json:"location,omitempty"`   // for tracking purposes
}

func (s *Session) TableName() string {
	return "session"
}

func (s *Session) MapFields() ([]string, []any) {
	return []string{
			"session_token",
			"account_id",
			"created_at",
			"updated_at",
			"expired_at",
			"is_active",
			"user_agent",
			"ip_address",
		}, []any{
			&s.SessionToken,
			&s.AccountID,
			&s.CreatedAt,
			&s.UpdatedAt,
			&s.ExpiredAt,
			&s.IsActive,
			&s.UserAgent,
			&s.IPAddress,
		}
}

func (s *Session) GetSessionToken() string {
	return s.SessionToken
}

func (s *Session) GetAccountID() string {
	return s.AccountID
}

func (s Session) ConvertToMapString() map[string]string {
	result := map[string]string{
		"session_token": s.SessionToken,
		"account_id":    s.AccountID,
		"user_agent":    s.UserAgent,
		"ip_address":    s.IPAddress,
	}

	if s.IsActive != nil {
		result["is_active"] = fmt.Sprintf("%t", *s.IsActive)
	}

	return result
}

func (s *Session) FromMap(m map[string]string) {
	s.SessionToken = m["session_token"]
	s.AccountID = m["account_id"]
	s.UserAgent = m["user_agent"]
	s.IPAddress = m["ip_address"]

	if isActive, ok := m["is_active"]; ok {
		if isActive == "true" {
			active := true
			s.IsActive = &active
		} else if isActive == "false" {
			active := false
			s.IsActive = &active
		}
	}
}
