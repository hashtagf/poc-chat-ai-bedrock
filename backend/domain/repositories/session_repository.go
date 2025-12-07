package repositories

import (
	"context"

	"github.com/bedrock-chat-poc/backend/domain/entities"
)

// SessionRepository defines the interface for session persistence
type SessionRepository interface {
	Create(ctx context.Context, session *entities.Session) error
	FindByID(ctx context.Context, id string) (*entities.Session, error)
	List(ctx context.Context) ([]*entities.Session, error)
	Update(ctx context.Context, session *entities.Session) error
	Delete(ctx context.Context, id string) error
	AddMessage(ctx context.Context, message *entities.Message) error
	GetMessages(ctx context.Context, sessionID string) ([]*entities.Message, error)
	IsExpired(session *entities.Session) bool
}
