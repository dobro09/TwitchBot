package store

import (
	"context"
	"database/sql"
	"fmt"
	"twbot/internal/model"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

type PostgresStore struct {
    db *sql.DB
}

func NewPostgresStore(databaseURL string) (*PostgresStore, error) {
    m, err := migrate.New("file://db/migrations", databaseURL)
    if err != nil {
        return nil, fmt.Errorf("не удалось создать экземпляр migrate: %w", err)
    }
    defer m.Close()

    err = m.Up()
    if err != nil && err != migrate.ErrNoChange {
        return nil, err
    }

    db, err := sql.Open("postgres", databaseURL)
    if err != nil {
        return nil, err
    }
    if err := db.Ping(); err != nil {
        return nil, err
    }
    return &PostgresStore{db: db}, nil
}

func (s *PostgresStore) InsertMessage(ctx context.Context, msg model.Message) error {
    // ... SQL INSERT
	sqlquery := `
	INSERT INTO messages (user_id, user_name, channel, text, timestamp)
	VALUES($1, $2, $3, $4, $5);
	`
	_, err := s.db.ExecContext(ctx, sqlquery, msg.UserID,msg.UserName, msg.Channel, msg.Text,msg.Timestamp)
	return err
}

func (s *PostgresStore) TopUsers(ctx context.Context, channel string, limit int) ([]model.UserStat, error) {
    // ... SQL SELECT с GROUP BY
	sqlquery := `
	SELECT user_id, user_name, COUNT(*) AS number_msgs FROM messages
	WHERE channel = $1
	GROUP BY user_id, user_name
	ORDER BY number_msgs DESC
	LIMIT $2;
	`
	rows, err := s.db.QueryContext(ctx, sqlquery, channel, limit)
	if err !=nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]model.UserStat, 0)

	for rows.Next() {
		var res model.UserStat
		err:= rows.Scan(
			&res.UserID,
			&res.UserName,
			&res.MessageCount,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, res)
	}
	if err := rows.Err(); err != nil { return nil, err }
	return result, nil
}