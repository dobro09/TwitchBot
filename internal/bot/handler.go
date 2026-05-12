package bot

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
	"twbot/internal/irc"
	"twbot/internal/store"
)

func Handler(
			connChan chan string,
			ctx context.Context,
			wg *sync.WaitGroup,
			msgChan chan string,
			cancel context.CancelFunc,
			pgStore store.MessageStore,
			accessToken string,
			clientID string,
			reconnectCh chan<- struct{}) {
	defer wg.Done()
	defer close(connChan)
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	lastActivity := time.Now()
	lastPONGfromTwitch:= time.Now()
	for{
		select{
		case <-ctx.Done():
			fmt.Println("завершено юзером")
			return
		case <-ticker.C:
			if time.Since(lastActivity) > 60*time.Second {
    			select {
    			case reconnectCh <- struct{}{}:
					fmt.Println("бездействие дольше 60 секунд")
    			default:
    			}
			}
			if time.Since(lastPONGfromTwitch) > 15*time.Second{
				select {
    			case reconnectCh <- struct{}{}:
					fmt.Println("нет ответа на мой ПИНГ больше 15 секунд")
    			default:
    			}
			
			}
			
		// case cmdmsg, ok:= <-cmdChan:
		// 	if !ok{
		// 		fmt.Println("cmdChan закрыт")
        // 		cmdChan = nil
    	// 		continue
		// 	}
		// 	if cmdmsg == "exit"{
		// 		fmt.Println("OTMENA")
		// 		cancel()
		// 		return
		// 	}
		case twmsg,ok := <-msgChan:
			
			if !ok {
        		fmt.Println("msgChan закрыт")
        		return
    		}
			lastActivity = time.Now()
			fmt.Println("RAW:", twmsg)
			msg:= irc.MsgParser(twmsg)
			switch msg.Command{
			case "PONG":
				lastPONGfromTwitch = time.Now()
			case "PING":
				pong := "PONG :" + msg.Text + "\r\n"
				fmt.Println("PONG ответ отправлен:", pong)
				connChan <- pong
			case "PRIVMSG":
				fmt.Println(msg.User,":", msg.Text)
				storemsg:= store.Message {
						UserID: msg.UserID,
						UserName: msg.User,
						Channel: msg.Channel,
						Text: msg.Text,
						Timestamp: time.Now(),
					}
				if err := pgStore.InsertMessage(ctx, storemsg); err != nil{
					fmt.Println("Ошибка сохранения сообщения:", err)
				}
				if strings.HasPrefix(msg.Text, "!"){
					cmd:=Command{
						UserID: msg.UserID,
						User: msg.User,
						Channel: msg.Channel,
						Text: msg.Text[1:],
						ParentMessageID: msg.MessageID,
					}
					go HandleCommand(cmd, connChan, pgStore, ctx, accessToken,clientID)
				}
			}
		}	
	}
}