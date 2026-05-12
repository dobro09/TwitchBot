package store

import (
	"context"
	"time"
)

type Message struct {
    UserID    string
    UserName  string
    Channel   string
    Text      string
    Timestamp time.Time
}

type UserStat struct {
    UserID       string
    UserName     string
    MessageCount int
}

type MessageStore interface {
    InsertMessage(ctx context.Context, msg Message) error
    TopUsers(ctx context.Context, channel string, limit int) ([]UserStat, error)
}

