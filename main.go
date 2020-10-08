package main

import (
	"log"
	"os"
	"time"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

const token = "TelegramBotToken"

func main() {
	if _, err := os.Stat("logs"); os.IsNotExist(err) {
		os.Mkdir("logs", 0644)
	}

	fileName := time.Now().Format("20060102_150405")
	f, err := os.OpenFile("logs/"+fileName+".log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	defer f.Close()
	if err != nil {
		log.Panic(err)
	}

	log.SetOutput(f)

	//---

	err = ReadSettings()
	if err != nil {
		log.Panic(err)
	}

	go SettingsWatcher()

	//---

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = false

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		command := FindBotCommand(update.Message.Text)
		var msg tgbotapi.MessageConfig

		switch command {
		case bcStart:
			msg = tgbotapi.NewMessage(update.Message.Chat.ID, cmnd_Start(update.Message.From.ID))
		case bcHelp:
			msg = tgbotapi.NewMessage(update.Message.Chat.ID, cmnd_Help(update.Message.From.ID))
		case bcInfo:
			msg = tgbotapi.NewMessage(update.Message.Chat.ID, cmnd_Info(update.Message.From.ID))
		case bcStatus:
			msg = tgbotapi.NewMessage(update.Message.Chat.ID, cmnd_Status(update.Message.From.ID))
		case bcBook:
			msg = tgbotapi.NewMessage(update.Message.Chat.ID, cmnd_Book(update.Message.Text, update.Message.From.ID, update.Message.From.UserName))
		case bcRelease:
			msg = tgbotapi.NewMessage(update.Message.Chat.ID, cmnd_Release(update.Message.Text, update.Message.From.ID))
		case bcMyId:
			msg = tgbotapi.NewMessage(update.Message.Chat.ID, cmnd_MyId(update.Message.From.ID))
		case bcWl:
			msg = tgbotapi.NewMessage(update.Message.Chat.ID, cmnd_Wl(update.Message.From.ID))
		case bcWlAdd:
			msg = tgbotapi.NewMessage(update.Message.Chat.ID, cmnd_WlAdd(update.Message.Text, update.Message.From.ID))
		case bcWlDel:
			msg = tgbotapi.NewMessage(update.Message.Chat.ID, cmnd_WlDel(update.Message.Text, update.Message.From.ID))
		case bcSerAdd:
			msg = tgbotapi.NewMessage(update.Message.Chat.ID, cmnd_SerAdd(update.Message.Text, update.Message.From.ID))
		case bcSerEdit:
			msg = tgbotapi.NewMessage(update.Message.Chat.ID, cmnd_SerEdit(update.Message.Text, update.Message.From.ID))
		case bcSerDel:
			msg = tgbotapi.NewMessage(update.Message.Chat.ID, cmnd_SerDel(update.Message.Text, update.Message.From.ID))
		}

		log.Printf("Unswer: " + msg.Text)

		if msg.Text != "" {
			bot.Send(msg)
		}
	}
}

func SettingsWatcher() {
	for {
		for i, v := range settings.Servers {
			if !v.IsFree && v.To < time.Now().Unix() {
				ReleaseServer(i)
				UpdateSettings()
			}
		}

		time.Sleep(time.Second * 5)
	}
}
