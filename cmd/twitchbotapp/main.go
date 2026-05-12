package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	v1 "twbot/api/gen/botstats/v1"
	"twbot/internal/bot"
	"twbot/internal/manager"
	"twbot/internal/store"
	api "twbot/internal/twitchapi"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

func main(){
	err:= godotenv.Load()
	if err != nil {
        log.Println("Файл .env не найден, используются только переменные окружения")
    }
	token := os.Getenv("TWITCH_TOKEN")
    botName := os.Getenv("BOT_NAME")
    channel := os.Getenv("CHANNEL")
	databaseURL := os.Getenv("DATABASE_URL")
	twitctClientID := os.Getenv("TWITCH_CLIENT_ID")
	twitchClientSecret := os.Getenv("TWITCH_CLIENT_SECRET")
	accessToken, err:= api.GetAppAccessToken(twitctClientID,twitchClientSecret)
	if err !=nil{
		log.Println("ошибка получения accessTokena, команды randomclip недоступна")
	}
    if token == "" {
        log.Fatal("TWITCH_TOKEN не задан")
    }

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	sigChan:= make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func(){
		<-sigChan
		fmt.Println("OTMENA")
		cancel()
	}()
	
	msgChan:= make(chan string, 100)
	//cmdChan:= make(chan string, 100)
	connChan:= make(chan string, 100)
	reconnectCh:= make(chan struct{})

	newstore, err := store.NewPostgresStore(databaseURL)
	if err !=nil{
		log.Fatalf("Не удалось подключиться к БД: %v", err)
	}

	grpcServer := v1.NewGRPCServer(newstore)
	s:= grpc.NewServer()
	v1.RegisterBotStatsServer(s, grpcServer)
	lis, err:= net.Listen("tcp", "localhost:50051")
	if err != nil {
    	log.Fatalf("не удалось запустить gRPC слушателя: %v", err)
	}
	go func ()  {
		if err:= s.Serve(lis); err !=nil{
			log.Fatalf("не удалось запустить gRPC слушателя: %v", err)
		}	
	}()
	go func() {
		<-ctx.Done()
    	s.GracefulStop()
	}()

	wg.Add(1)
	go manager.ConnectionManager(token,botName,channel,msgChan,&wg,ctx,connChan,reconnectCh)
	
	// inputStdin:= bufio.NewScanner(os.Stdin)
	// go func (inputStdin *bufio.Scanner)  {
	// 	defer close(cmdChan)

	// 	for inputStdin.Scan(){
	// 		select {
	// 		case <-ctx.Done():
	// 			return
	// 		case cmdChan<- inputStdin.Text():
	// 		}
	// 	}
	// 	if err:= inputStdin.Err(); err != nil {
	// 		fmt.Println("Ошибка чтения из консоли:", err)
   	// 		return
	// 	}
	// }(inputStdin)

	wg.Add(1)
	go bot.Handler(connChan, ctx, &wg, msgChan, cancel, newstore, accessToken, twitctClientID,reconnectCh)
	wg.Wait()
	fmt.Println("закончилося код")
}