package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	botdelivery "twbot/internal/BotDelivery"
	"twbot/internal/store"
	api "twbot/internal/twitchapi"
	"twbot/internal/usecase"

	"github.com/joho/godotenv"
)

func main(){
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	err := godotenv.Load()
	if err != nil {
        log.Println("Файл .env не найден, используются только переменные окружения")
    }
	twitchClientID := os.Getenv("TWITCH_CLIENT_ID")
	twitchClientSecret := os.Getenv("TWITCH_CLIENT_SECRET")
	twitchClient, err := api.NewTwitchClient(twitchClientID, twitchClientSecret)
	if err != nil {
    	log.Fatalf("Не удалось создать Twitch клиент: %v", err)
	}

	databaseURL := os.Getenv("DATABASE_URL")
	defaultStore := os.Getenv("STORE_TYPE")
	if defaultStore == "" {
    	defaultStore = "postgres"
	}
	storeType := flag.String("store", defaultStore, "Тип хранилища: postgres или memory")
	flag.Parse()

	newstore, err := store.InitStore(*storeType, databaseURL)
	if err != nil {
    	log.Fatalf("Не удалось инициализировать хранилище: %v", err)
	}

	b:= botdelivery.BotDelivery{}
	cu:= usecase.NewChatUsecase(newstore, twitchClient)
	b.Init(cu)
	wg.Add(1)
	go b.Connect(ctx, &wg)
	wg.Add(1)
	go b.Handle(ctx, &wg)
	

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("OTMENA")
		cancel()
	}()
	wg.Wait()
}