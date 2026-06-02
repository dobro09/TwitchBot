package usecase

import (
	"context"
	"twbot/internal/model"
)

type ChatUsecase interface {
	SaveMessage(ctx context.Context, msg model.Message) error     // сохранить, проверить сообщение
	CompleteCommand(ctx context.Context, cmd model.Command) (string, error) // выполнить команду, вернуть ответ
}

type TwitchAPIClient interface {
    GetBroadcasterID(channelName string) (string, error)
    GetClips(broadcasterID string) ([]string, error)
    RandomClip(urls []string) (string, error)
}