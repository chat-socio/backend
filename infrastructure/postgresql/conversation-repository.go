package postgresql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/chat-socio/backend/internal/domain"
)

type conversationRepository struct {
	db *sql.DB
}

// CreateConversation implements domain.ConversationRepository.
func (c *conversationRepository) CreateConversation(ctx context.Context, conversation *domain.Conversation, conversationMembers []*domain.ConversationMember) (*domain.Conversation, error) {
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO conversation (id, created_at, type, title, avatar, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err = tx.ExecContext(ctx, query, conversation.ID, conversation.CreatedAt, conversation.Type, conversation.Title, conversation.Avatar, conversation.UpdatedAt)
	if err != nil {
		return nil, err
	}

	query = `
		INSERT INTO conversation_member (id, conversation_id, user_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	for _, conversationMember := range conversationMembers {
		_, err = tx.ExecContext(ctx, query, conversationMember.ID, conversationMember.ConversationID, conversationMember.UserID, conversationMember.CreatedAt, conversationMember.UpdatedAt)
		if err != nil {
			return nil, err
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return conversation, nil
}

// GetConversationByID implements domain.ConversationRepository.
func (c *conversationRepository) GetConversationByID(ctx context.Context, id string) (*domain.Conversation, []*domain.ConversationMemberWithUser, error) {
	var conversation domain.Conversation
	var conversationMembers []*domain.ConversationMemberWithUser

	fields, values := conversation.MapFields()
	// query get conversation and conversation members
	query := fmt.Sprintf(`SELECT %s FROM conversation WHERE id = $1`, strings.Join(fields, ", "))
	row := c.db.QueryRowContext(ctx, query, id)
	if row.Err() != nil {
		return nil, nil, row.Err()
	}

	if err := row.Scan(values...); err != nil {
		return nil, nil, err
	}

	query = `SELECT cm.conversation_id, cm.user_id, ui.full_name, ui.avatar, ui.type FROM conversation_member cm
		INNER JOIN user_info ui ON cm.user_id = ui.id
		WHERE cm.conversation_id = $1`
	rows, err := c.db.QueryContext(ctx, query, id)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var conversationMember domain.ConversationMemberWithUser
		if err := rows.Scan(&conversationMember.ConversationID, &conversationMember.UserID, &conversationMember.FullName, &conversationMember.Avatar, &conversationMember.UserType); err != nil {
			return nil, nil, err
		}
		conversationMembers = append(conversationMembers, &conversationMember)
	}

	return &conversation, conversationMembers, nil
}

// GetListConversationByUserID implements domain.ConversationRepository.
func (c *conversationRepository) GetListConversationByUserID(ctx context.Context, userID string, lastMessageID string, limit int) ([]*domain.Conversation, error) {
	var conversations []*domain.Conversation
	var conditionLastMessageID string
	var params []any
	params = append(params, userID)
	if lastMessageID != "" {
		conditionLastMessageID = `AND last_message_id < $2`
		params = append(params, lastMessageID)
	} 
	// Add NULL handling for last_message_id
	fieldsWithCoalesce := []string{
		"c.id",
		"c.created_at", 
		"c.type",
		"c.title", 
		"c.avatar",
		"c.updated_at",
		"c.deleted_at",
		"COALESCE(c.last_message_id::text, '')", // Handle NULL last_message_id
		"COALESCE(m.id::text, '')", // Handle NULL message id
		"COALESCE(m.conversation_id::text, '')",
		"COALESCE(m.user_id::text, '')",
		"COALESCE(m.type::text, '')",
		"COALESCE(m.body::text, '')",
		"COALESCE(m.created_at, NULL)",
		"COALESCE(m.updated_at, NULL)", 
		"COALESCE(m.reply_to::text, '')",
		"COALESCE(ui.full_name::text, '')",
		"COALESCE(ui.avatar::text, '')",
		"COALESCE(ui.type::text, '')",
	}

	query := fmt.Sprintf(`SELECT %s FROM conversation c
		LEFT JOIN message m ON c.last_message_id = m.id
		LEFT JOIN user_info ui ON m.user_id = ui.id
		WHERE c.id IN (SELECT DISTINCT conversation_id FROM conversation_member WHERE user_id = $1) %s
		ORDER BY last_message_id DESC LIMIT %d`, strings.Join(fieldsWithCoalesce, ", "), conditionLastMessageID, limit)


	rows, err := c.db.QueryContext(ctx, query, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var conversation domain.Conversation
		var message domain.Message
		var userInfo domain.UserInfo
		values := []any{
			&conversation.ID,
			&conversation.CreatedAt,
			&conversation.Type,
			&conversation.Title,
			&conversation.Avatar,
			&conversation.UpdatedAt,
			&conversation.DeletedAt,
			&conversation.LastMessageID,
			&message.ID,
			&message.ConversationID,
			&message.UserID,
			&message.Type,
			&message.Body,
			&message.CreatedAt,
			&message.UpdatedAt,
			&message.ReplyTo,
			&userInfo.FullName,
			&userInfo.Avatar,
			&userInfo.Type,
		}
		if err := rows.Scan(values...); err != nil {
			return nil, err
		}
		if message.ID == "" {
			conversation.LastMessage = nil
		} else {
			if userInfo.FullName == "" && userInfo.Avatar == "" && userInfo.Type == "" {
				message.User = nil
			} else {
				message.User = &userInfo
			}
			conversation.LastMessage = &message
		}
		conversations = append(conversations, &conversation)
	}

	return conversations, nil
}

// UpdateLastMessageID implements domain.ConversationRepository.
func (c *conversationRepository) UpdateLastMessageID(ctx context.Context, conversationID string, lastMessageID string) (*domain.Conversation, error) {
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	var params []any
	params = append(params, conversationID)
	var lastMessageIDQuery string
	var updatedAtQuery *time.Time
	query := `SELECT last_message_id, updated_at FROM conversation WHERE id = $1 FOR UPDATE`
	err = tx.QueryRowContext(ctx, query, params...).Scan(&lastMessageIDQuery, &updatedAtQuery)
	if err != nil {
		return nil, err
	}
	query = `
		UPDATE conversation
		SET last_message_id = $1, updated_at = NOW()
		WHERE id = $2 AND last_message_id = $3 AND updated_at = $4
		RETURNING id, created_at, type, title, avatar, updated_at, deleted_at, last_message_id
	`

	var conversation domain.Conversation
	_, values := conversation.MapFields()

	err = tx.QueryRowContext(ctx, query, lastMessageID, conversationID, lastMessageIDQuery, updatedAtQuery).Scan(values...)
	if err != nil {
		return nil, err
	}

	return &conversation, nil
}

var _ domain.ConversationRepository = &conversationRepository{}

func NewConversationRepository(db *sql.DB) domain.ConversationRepository {
	return &conversationRepository{db: db}
}
