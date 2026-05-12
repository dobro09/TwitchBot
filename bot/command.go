package bot

import (
	"context"
	"fmt"
	"strings"
	"twbot/store"
	api "twbot/twitchapi"
)

type Command struct {
	UserID string    // ID юзера
    User    string   // ник отправителя
    Channel string   // канал, куда отвечать
    Text    string   // полный текст после "!"
	ParentMessageID string //  ID СООБЩЕНИЯ юзера которое было отправлено
}

func HandleCommand(cmd Command, 
				connChan chan<- string, 
				pgStore store.MessageStore, 
				ctx context.Context,
				token string,
				clientID string,
				){
	parts:= strings.SplitN(cmd.Text, " ", 2)
	command:= parts[0]
	var args string
	if len(parts)>1{
		args= parts[1]
	}
	prefix:= "PRIVMSG " + cmd.Channel + " :"
	if cmd.ParentMessageID != ""{
		prefix= "@reply-parent-msg-id=" + cmd.ParentMessageID + " " + prefix
	}
	switch command{
	case "ping":
		connChan<- prefix + "Pong!" + "\r\n"
	case "echo":
		connChan<- prefix + args + "\r\n"
	case "top":
		///// статистика канала, топ 3 кто писал сообщения в канал - выводит имя юзера, кол-во сообщений
		dbresult, err := pgStore.TopUsers(ctx, cmd.Channel, 3)
		if len(dbresult) == 0{
			connChan <- prefix + "Сообщений пока нет.\r\n"
			return
		}
		if err !=nil{
			connChan<- prefix + "не получается получить статистику с БД" + "\r\n"
			return
		}
		var builder strings.Builder
		for _, val := range dbresult{
			fmt.Fprintf(&builder,"%s: %d сообщений, ", val.UserName, val.MessageCount)
			// builder.WriteString(fmt.Sprintf("%s: %d,", val.UserName, val.MessageCount))
		}
		end := builder.String()
		connChan<- prefix + strings.TrimRight(end, ",") + "\r\n"
	case "userstats":
		/// статистика юзера, кол-во его сообщений, самые повторяемое сообщение
	case "randomclip": // получить ссылку на рандомный клип с канала пользователя (указывается после) (randomclip user_channel)
		broadcasterID, err:=api.GetBroadcasterID(token, clientID, args)
		if err!= nil{
			connChan<- prefix + "не найден канал/ошибка при попытки найти канал" + "\r\n"
			return
		}
		clips, err := api.GetClips(token,clientID, broadcasterID)
		if err!=nil{
			connChan<- prefix + "нет клипов на канале/ошибка при попытки найти клипы" + "\r\n"
			return
		}
		randomClip, err := api.RandomClip(clips)
		if err !=nil{
			connChan<- prefix + "нет клипов на канале/ошибка при попытки найти клипы" + "\r\n"
			return
		}
		connChan<- prefix + randomClip + "\r\n"
	case "help":
		connChan<- prefix + "Доступные команды бота: !ping, !echo, !top, !randomclip" + "\r\n"
	case "topemoteonchannel":
		// читает в канале(user_channel) сообщения, содержащие одно слово(бол-во сообщений с одним словом это емоут,
		// подсчитаывает сколько раз он был использован и выводит  топ 3 емоутов и сколько раз они были написаны в канале
	}
}