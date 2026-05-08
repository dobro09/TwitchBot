package irc

import "strings"

type Message struct{
	Raw string // Полный необработанный текст
	User string // Юзер который отправил сообщение
	Command string // Команда, которая пришла
	Channel string // Канал, на которм пришло сообщение
	Text string // Обработанный текст,(без тегов), который прислал Юзер
	Tags string // Теги
	MessageID string // ID СООБЩЕНИЯ юзера которое было отправлено
	UserID string // ID ЮЗЕРА который отправил сообщение
}

func MsgParser(twmsg string)(msg Message){
	if twmsg == ""{
		return
	}
	restt := twmsg
	msg.Raw = twmsg
	if strings.HasPrefix(twmsg, "@"){
		idx:= strings.IndexByte(twmsg, ' ')
		if idx != -1{
			tagsPart := twmsg[1:idx]
			rest:= twmsg[idx+1:]
			restt = rest
			msg.Tags = tagsPart
			msg.MessageID = getValueMessageID(tagsPart)
			msg.UserID = getValueUserID(tagsPart)
		}
	}
	fieldsStr:= strings.Fields(restt)
	if strings.HasPrefix(restt, ":"){
		nick:= fieldsStr[0]
		nick = strings.TrimPrefix(nick, ":")
		idx:= strings.IndexByte(nick, '!')
		if idx != -1{
			nick= nick[:idx]
		}
		msg.User = nick
		fieldsStr= fieldsStr[1:]
	}
	msg.Command = fieldsStr[0]

	textIdx:= -1
	for i := 1; i < len(fieldsStr); i++ {
		if strings.HasPrefix(fieldsStr[i], ":"){
			textIdx = i
			break
		}
	}
	
	if msg.Command == "PRIVMSG" && len(fieldsStr)>1 {
		msg.Channel = fieldsStr[1]
		if textIdx == -1{
			msg.Text= ""
		} else{
			text := fieldsStr[textIdx][1:]
			if textIdx+1 < len(fieldsStr){
				text += " " + strings.Join(fieldsStr[textIdx+1:], " ")
			}
			msg.Text= text
		}	
	}

	if msg.Command == "PING" && textIdx!= -1{
		msg.Text = fieldsStr[textIdx][1:]
	}
	return msg
}

func getValueMessageID(tag string)(value string){
	parts:= strings.Split(tag, ";")
	for _, val := range parts{
		key, value, _:=strings.Cut(val, "=")
		if key == "id"{
			return value
		}
	}
	return ""
}

func getValueUserID(tag string)(value string){
	parts:= strings.Split(tag, ";")
	for _, val := range parts{
		key, value, _:=strings.Cut(val, "=")
		if key == "user-id"{
			return value
		}
	}
	return ""
}