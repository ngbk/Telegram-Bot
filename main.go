package main

import (
	"context"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
	"log"
	"os"
	telegram "telegram_bot/client"
	event_consumer "telegram_bot/consumer/event-consumer"
	"telegram_bot/db/repository"
	sessions "telegram_bot/db/session"
	"telegram_bot/events"
)

const (
	tgGotHost        = "api.telegram.org"
	apiTokenTelegram = "API_TOKEN_TELEGRAM"
	batchSize        = 100
	offset           = 0
)

func main() {
	loadConfigs()

	tgClient := telegram.New(tgGotHost, getToken(apiTokenTelegram))

	ctx := context.Background()

	db, err := repository.New(repository.ConfigDB{
		Host:     viper.GetString("db.host"),
		Port:     viper.GetString("db.port"),
		Username: viper.GetString("db.username"),
		Password: os.Getenv("PASSWORD_DB"),
		DBName:   viper.GetString("db.db"),
		SSLMode:  viper.GetString("db.sslmode"),
	})
	if err != nil {
		log.Fatal("can't create db: ", err)
	}

	redisdb, err := sessions.NewRedisClient(ctx, sessions.ConfigRedis{
		Addr:     viper.GetString("session.addr"),
		Password: viper.GetString("session.password"),
		Db:       viper.GetInt("session.db"),
	})

	if err != nil {
		log.Fatal("can't connect session", err)
	}

	eventProcessor := events.New(tgClient, db, *redisdb)

	log.Print("service started")

	consumer := event_consumer.New(eventProcessor, eventProcessor, batchSize)

	if err = consumer.Start(); err != nil {
		log.Fatal("service is stopped", err)
	}

}

func loadConfigs() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("can't load configs", err)
	}

	viper.AddConfigPath("./")
	viper.SetConfigName("configs")
	err = viper.ReadInConfig()

	if err != nil {
		log.Fatal("can't load configs", err)
	}
}

func getToken(apiToken string) string {
	return os.Getenv(apiToken)
}
