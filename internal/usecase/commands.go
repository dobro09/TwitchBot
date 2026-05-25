package usecase

import (
	"context"
	"fmt"
	"strings"
	"twbot/internal/model"
	"twbot/internal/store"
)


type chatUsecase struct{
	store store.MessageStore
	twitchapi TwitchAPIClient
	commands map[string]CommandHandler
}

type CommandHandler func(ctx context.Context, cmd model.Command) (string, error)


func NewChatUsecase(store store.MessageStore, twitchapi TwitchAPIClient)(ChatUsecase){
	cu := chatUsecase{
        store:     store,
        twitchapi: twitchapi,
        commands:  make(map[string]CommandHandler),
    }
	cu.commands["help"] = cu.helpCommand
    cu.commands["ping"] = cu.pingCommand
	cu.commands["echo"] = cu.echoCommand
	cu.commands["top"] = cu.topCommand
	cu.commands["randomclip"] = cu.randomClipCommand
	return &cu
}

func(c *chatUsecase) helpCommand(ctx context.Context, cmd model.Command) (string, error) {
    return "Доступные команды бота: !ping, !echo, !top, !randomclip.", nil
}

func(c *chatUsecase) pingCommand(ctx context.Context, cmd model.Command) (string, error) {
    return "Pong!", nil
}

func(c *chatUsecase) echoCommand(ctx context.Context, cmd model.Command) (string, error) {
    return cmd.Args, nil
}

func(c *chatUsecase) topCommand(ctx context.Context, cmd model.Command)(string, error){
	dbresult, err := c.store.TopUsers(ctx, cmd.Channel, 3)
	if err !=nil{
			return "", err
		}
	if len(dbresult) == 0{
			return "Сообщений пока нет.", nil
		}
		
	var builder strings.Builder
	for _, val := range dbresult{
		fmt.Fprintf(&builder,"%s: %d сообщений, ", val.UserName, val.MessageCount)
		}
	end := builder.String()
	return strings.TrimRight(end, ", "), nil
}

func(c *chatUsecase) randomClipCommand(ctx context.Context, cmd model.Command)(string, error){
	broadcasterID, err:=c.twitchapi.GetBroadcasterID(cmd.Args)
	if err!= nil{
		return "не найден канал/ошибка при попытки найти канал", nil
	}
	clips, err := c.twitchapi.GetClips(broadcasterID)
	if err!=nil{
		return "нет клипов на канале/ошибка при попытки найти клипы", nil
	}
	randomClip, err := c.twitchapi.RandomClip(clips)
	if err !=nil{	
		return "нет клипов на канале/ошибка при попытки найти клипы", nil
	}
	return randomClip, nil
}

func(c *chatUsecase) SaveMessage(ctx context.Context, msg model.Message) error {
	if err := c.store.InsertMessage(ctx, msg); err != nil{
		return err
	}
	return nil
}

func(c *chatUsecase) CompleteCommand(ctx context.Context, cmd model.Command) (string, error){
	handler, ok := c.commands[cmd.Command]
	if !ok{
		return "", fmt.Errorf("неправильная команда")
	}
	res, err:= handler(ctx, cmd)
	if err != nil{
		return "", fmt.Errorf("ошибка выполнения запроса %v", err)
	}
	return res, nil
}