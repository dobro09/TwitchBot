package model

import "time"

type Command struct {
	Command string 	 // команда которая вызывается
	Args 	string   // аргументы после команды
	UserID  string    // ID юзера
    UserName    string   // ник отправителя
    Channel string   // канал, куда отвечать
    Text    string   // полный текст после "!"
	ParentMessageID string //  ID СООБЩЕНИЯ юзера которое было отправлено
}

type Message struct {
	UserID    string
	UserName  string
	MessageID string
	Channel   string
	Text      string
	Timestamp time.Time
}

type UserStat struct {
	UserID       string
	UserName     string
	MessageCount int
}

type RAWMessage struct {
	Raw string // Полный необработанный текст
	User string // Юзер который отправил сообщение
	Command string // Команда, которая пришла
	Channel string // Канал, на которм пришло сообщение
	Text string // Обработанный текст,(без тегов), который прислал Юзер
	Tags string // Теги
	MessageID string // ID СООБЩЕНИЯ юзера которое было отправлено
	UserID string // ID ЮЗЕРА который отправил сообщение
}