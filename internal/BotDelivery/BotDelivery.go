package botdelivery

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"sync"
	"time"

	"twbot/internal/model"
	"twbot/internal/usecase"
	"twbot/internal/utils"
)

type BotDelivery struct {
	token   string
	botName string
	channel string
	msgChan chan string
	connChan chan string
	reconnectCh chan struct{}
	cu usecase.ChatUsecase
}

func (b *BotDelivery) Init(cu usecase.ChatUsecase) {
	b.token = os.Getenv("TWITCH_TOKEN")
	if b.token == "" {
        log.Fatal("TWITCH_TOKEN не задан")
    }
	b.botName = os.Getenv("BOT_NAME")
	b.channel = os.Getenv("CHANNEL")
	b.msgChan= make(chan string, 100)
	b.connChan= make(chan string, 100)
	b.reconnectCh= make(chan struct{})
	b.cu = cu
}

func (b *BotDelivery) Connect(ctx context.Context, wg *sync.WaitGroup) {					
	defer wg.Done()
	defer close(b.msgChan)

	var delay time.Duration = 1 * time.Second

	for {
		conn, err := net.Dial("tcp", "irc.chat.twitch.tv:6667")
		if err !=nil {
			log.Println("не удалось подключиться к Twitch")
			if waitBackoff(ctx, &delay){return}
			continue
		}
		if tcpConn, ok := conn.(*net.TCPConn); ok {
    		tcpConn.SetKeepAlive(true)
    		tcpConn.SetKeepAlivePeriod(30 * time.Second)
		}

		delay = 1 * time.Second
		fmt.Println("подключился к серверу twitch")
		childctx, childcancel:= context.WithCancel(ctx)

		go func() {
    		for {
       			select {
        		case <-childctx.Done():
					return
				case msg, ok := <-b.connChan:
					if !ok {
						return
					}
					conn.SetWriteDeadline(time.Now().Add(3*time.Second))
					_, err := conn.Write([]byte(msg))
					if err != nil {
						log.Println("Ошибка записи в Twitch, инициирую переподключение")
						childcancel()
						return
					}
				}
			}
		}()

		fmt.Fprintf(conn, "CAP REQ :twitch.tv/tags twitch.tv/commands\r\n")
		fmt.Fprintf(conn, "CAP REQ :twitch.tv/membership\r\n")
		fmt.Fprintf(conn, "CAP END\r\n")
		fmt.Fprintf(conn, "PASS oauth:%s\r\n", b.token)
		fmt.Println("отправил токен")
		fmt.Fprintf(conn, "NICK %s\r\n", b.botName)
		fmt.Println("отправил ник бота")
		fmt.Fprintf(conn, "JOIN #%s\r\n", b.channel)

		input:= bufio.NewScanner(conn)
		
		go func(input *bufio.Scanner, 
			conn net.Conn, 
			msgChan chan string,
			childCancel context.CancelFunc) {
			for input.Scan() {
				select {
				case <-childctx.Done():
					fmt.Println("завершено юзером")
					return
				case msgChan<- input.Text():
				}
			}
			if err := input.Err(); err != nil {
				childcancel()
				fmt.Println("Ошибка чтения из Twitch:", err)
				return
			}
		}(input, conn, b.msgChan, childcancel)
		go func(ctx context.Context, connChan chan string) {
			ticker := time.NewTicker(12 * time.Second)
			for {
				select {
				case <-childctx.Done():
					ticker.Stop()
					return
				case <-ticker.C:
					select {
					case connChan<-"PING :tmi.twitch.tv\r\n":
						fmt.Println("PING отправлен")
					default:
						log.Println("канал connChan не готов(заполнен или закрыт)")
					}			
				}
			}
		}(childctx, b.connChan)
		select {
		case <-ctx.Done():
			childcancel()
			return
		case <-childctx.Done():
			childcancel()
			conn.Close()
			if waitBackoff(ctx, &delay){return}
			continue
		case <-b.reconnectCh:
			fmt.Println("Сигнал переподключения (idle таймаут)")
			childcancel()
			conn.Close()
			continue
		}
	}
}
func waitBackoff(ctx context.Context, delay *time.Duration) bool {
	if *delay == 0 {
		*delay = 1 * time.Second
	}
	if *delay > 30*time.Second { 
		*delay = 30*time.Second 
	}
	select {
	case <-ctx.Done():
		return true
	case <-time.After(*delay):
		*delay= *delay * 2
		return false
	}
}

func (b *BotDelivery) Handle(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	defer close(b.connChan)
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	lastActivity := time.Now()
	lastPONGfromTwitch := time.Now()
	for {
		select {
		case <-ctx.Done():
			fmt.Println("завершено юзером")
			return
		case <-ticker.C:
			if time.Since(lastActivity) > 60*time.Second {
    			select {
    			case b.reconnectCh <- struct{}{}:
					fmt.Println("бездействие дольше 60 секунд")
    			default:
    			}
			}
			if time.Since(lastPONGfromTwitch) > 15*time.Second {
				select {
    			case b.reconnectCh <- struct{}{}:
					fmt.Println("нет ответа на мой ПИНГ больше 15 секунд")
    			default:
    			}
			}
		case twmsg ,ok := <-b.msgChan:
			if !ok {
        		fmt.Println("msgChan закрыт")
        		return
    		}
			lastActivity = time.Now()
			fmt.Println("RAW:", twmsg)
			rawmsg := utils.MsgParser(twmsg)
			switch rawmsg.Command {
			case "PONG":
				lastPONGfromTwitch = time.Now()
			case "PING":
				pong := "PONG :" + rawmsg.Text + "\r\n"
				fmt.Println("PONG ответ отправлен:", pong)
				b.connChan <- pong
			case "PRIVMSG":
				fmt.Println(rawmsg.User,":", rawmsg.Text)
				msg := model.Message {
					UserID: rawmsg.UserID,
					UserName: rawmsg.User,
					Channel: rawmsg.Channel,
					Text: rawmsg.Text,
					MessageID: rawmsg.MessageID,
					Timestamp: time.Now(),
				}
				if err := b.cu.SaveMessage(ctx, msg); err != nil {
					fmt.Println("Ошибка сохранения сообщения:", err)
				}
				if strings.HasPrefix(msg.Text, "!") {
					parts := strings.SplitN(msg.Text[1:], " ", 2)
					command := parts[0]
					var args string
					if len(parts) > 1 {
						args = parts[1]
					}
					cmd := model.Command {
						UserID: msg.UserID,
						UserName: msg.UserName,
						Channel: msg.Channel,
						Text: msg.Text[1:],
						ParentMessageID: msg.MessageID,
						Command: command,
						Args: args,
					}
					res, err := b.cu.CompleteCommand(ctx, cmd)
					if err != nil {
						fmt.Println("Ошибка выполнения команды:", err)
					}
					prefix := "PRIVMSG " + cmd.Channel + " :"
					if cmd.ParentMessageID != "" {
						prefix = "@reply-parent-msg-id=" + cmd.ParentMessageID + " " + prefix
					}
					b.connChan <- prefix + res + "\r\n"
				}
			}
		}
	}	
}
