package manager

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

func ConnectionManager( 
					token string,
					botName string,
					channel string,
					msgChan chan string,
					wg *sync.WaitGroup,
					ctx context.Context,
					connChan chan string,
					reconnectCh <-chan struct{}){
						
	defer wg.Done()
	defer close(msgChan)
	var delay time.Duration = 1 * time.Second
	for{
		conn, err:= net.Dial("tcp", "irc.chat.twitch.tv:6667")
		if err !=nil{
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
				case msg, ok := <-connChan:
					if !ok {
						return
					}
					conn.SetWriteDeadline(time.Now().Add(3*time.Second))
					_, err:= conn.Write([]byte(msg))
					if err !=nil{
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
		fmt.Fprintf(conn, "PASS oauth:%s\r\n", token)
		fmt.Println("отправил токен")
		fmt.Fprintf(conn, "NICK %s\r\n", botName)
		fmt.Println("отправил ник бота")
		fmt.Fprintf(conn, "JOIN #%s\r\n", channel)

		input:= bufio.NewScanner(conn)
		
		go func(input *bufio.Scanner, 
			conn net.Conn, 
			msgChan chan string,
			childCancel context.CancelFunc){
			for input.Scan(){
				select{
				case<-childctx.Done():
					fmt.Println("завершено юзером")
					return
				case msgChan<- input.Text():
				}
			}
			if err:= input.Err(); err != nil{
				childcancel()
				fmt.Println("Ошибка чтения из Twitch:", err)
				return
			}
		}(input, conn, msgChan, childcancel)
		go func( ctx context.Context, connChan chan string){
			ticker := time.NewTicker(12 * time.Second)
			for{
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
		}(childctx, connChan)
		// go func() {
		// 	ticker := time.NewTicker(30 * time.Second)
		// 	for{
		// 		select {
		// 			case _, ok := <-connChan:
		// 			if !ok {
		// 				return
		// 			}
		// 		case <-ticker.C:
		// 			childcancel()
		// 			log.Println("бездейсвтие больше 30 секунд")
		// 			return
		// 		}
		// 	}
		// }()
		select {
		case <-ctx.Done():
			childcancel()
			return
		case<-childctx.Done():
			childcancel()
			conn.Close()
			if waitBackoff(ctx, &delay){return}
			continue
		case <-reconnectCh:
			fmt.Println("Сигнал переподключения (idle таймаут)")
			childcancel()
			conn.Close()
			continue
		}
	}
}

func waitBackoff(ctx context.Context, delay *time.Duration) bool{
	if *delay==0 {
		*delay = 1 * time.Second
	}
	if *delay > 30*time.Second { 
		*delay = 30*time.Second 
	}
	select{
	case <-ctx.Done():
		return true
	case <-time.After(*delay):
		*delay= *delay * 2
		return false
	}
}