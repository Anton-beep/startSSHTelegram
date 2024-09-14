package main

import (
	"os"
	"startSSHTelegram/telegramBot"
	"startSSHTelegram/telegramLogic"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func InitLogger(debug bool) {
	cfg := zap.NewDevelopmentConfig()
	if debug {
		cfg.Level.SetLevel(zap.DebugLevel)
	}
	logger, err := cfg.Build()
	if err != nil {
		panic(err)
	}

	zap.ReplaceGlobals(logger)
	zap.S().Info("Start")
}

func loadEnv() {
	err := godotenv.Load("../.env")
	if err != nil {
		zap.S().Panic("cannot read .env, try to change .env.template")
	}
}

func parseGoodChats(chats string) []int {
	res := []int{}
	for _, el := range strings.Split(chats, ",") {
		val, err := strconv.Atoi(el)
		if err != nil {
			zap.S().Panic("not a number in good chats .env")
		}
		res = append(res, val)
	}

	return res
}

func main() {
	InitLogger(true)
	loadEnv()

	tgBot := telegramBot.New(os.Getenv("TOKEN"))

	tgLogic, err := telegramLogic.New(tgBot, parseGoodChats(os.Getenv("GOOD_CHATS")), os.Getenv("NGROK_PATH"))
	if err != nil {
		panic(err)
	}
	tgLogic.Run()
}
