package main

import (
	"log"
	"os"
	"strconv"
	"strings"

	"kuzma975/project_7648/handler"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gopkg.in/yaml.v3"
)

type config struct {
	Telegram struct {
		Token string `yaml:"token"`
		Debug bool   `yaml:"debug"`
		Test  bool   `yaml:"test"`
	} `yaml:"telegram"`
}

func main() {
	configurationFileName := "config.yaml"
	file, err := os.Open(configurationFileName)
	if err != nil {
		log.Panic(err)
	}
	defer file.Close()
	var conf config
	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(&conf)
	if err != nil {
		log.Panic(err)
	}

	bot, err := tgbotapi.NewBotAPI(conf.Telegram.Token)
	if err != nil {
		log.Panic("Error occured during creating new bot client: ", err)
	}
	bot.Debug = conf.Telegram.Debug

	log.Printf("Authorized on account %s", bot.Self.UserName)
	test := conf.Telegram.Test
	offset := func() int {
		if test {
			return -1
		} else {
			return 0
		}
	}()
	u := tgbotapi.NewUpdate(offset)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)
	var latest *tgbotapi.Message
	if test {
		start := <-updates
		if start.Message != nil {
			handler.HandleMessage(start, *bot, db, test)
			defer func() {
				log.Printf("latest id is %d", latest.MessageID)
				log.Printf("first id %d", start.Message.MessageID)
				for i := start.Message.MessageID; i <= latest.MessageID; i++ {
					toDelete := tgbotapi.NewDeleteMessage(start.Message.Chat.ID, i)
					if resp, err := bot.Request(toDelete); err != nil {
						log.Printf("Cloud not delete message %s", err)
					} else {
						log.Printf("Response is %v", resp)
					}
				}
			}()
		}
	}

	// log.Printf("config file is: %s", conf)
	for update := range updates {
		if update.Message != nil { // If we got a message
			latest = update.Message
			if handler.HandleMessage(update, *bot, db, false) {
				return
			}
		}
	}
}
