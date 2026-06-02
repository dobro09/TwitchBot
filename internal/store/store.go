package store

import (
	"context"
	"twbot/internal/model"
)

type MessageStore interface {
    InsertMessage(ctx context.Context, msg model.Message) error
    TopUsers(ctx context.Context, channel string, limit int) ([]model.UserStat, error)
}