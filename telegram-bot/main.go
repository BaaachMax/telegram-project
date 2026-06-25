package main

import (
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Укажите здесь HTTPS-адрес вашего Mini App (например, из ngrok или после деплоя)
const miniAppURL = "https://example.com" // <-- замените на свой URL

func main() {
	botToken := os.Getenv("BOT_TOKEN")
	if botToken == "" {
		log.Fatal("Переменная окружения BOT_TOKEN не установлена. Установите её перед запуском:\n  export BOT_TOKEN=\"ваш_токен\"")
	}

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatalf("Не удалось авторизоваться: %v", err)
	}

	bot.Debug = false
	log.Printf("Бот запущен как @%s", bot.Self.UserName)

	// Настраиваем long polling — бот будет получать новые сообщения
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 30

	updates := bot.GetUpdatesChan(updateConfig)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		switch update.Message.Command() {
		case "start":
			handleStart(bot, update.Message)
		default:
			handleEcho(bot, update.Message)
		}
	}
}

// handleStart отправляет приветствие и кнопку для открытия Mini App
func handleStart(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	msg := tgbotapi.NewMessage(message.Chat.ID, "👋 Привет! Нажми кнопку ниже, чтобы открыть приложение.")

	// Кнопка, открывающая Mini App прямо внутри Telegram
	webAppButton := tgbotapi.KeyboardButton{
		Text:   "Открыть приложение",
		WebApp: &tgbotapi.WebappInfo{URL: miniAppURL},
	}

	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(webAppButton),
	)
	keyboard.ResizeKeyboard = true

	msg.ReplyMarkup = keyboard

	if _, err := bot.Send(msg); err != nil {
		log.Printf("Ошибка отправки сообщения: %v", err)
	}
}

// handleEcho — простой обработчик для всех остальных сообщений (пример)
func handleEcho(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	msg := tgbotapi.NewMessage(message.Chat.ID, "Я получил: "+message.Text+"\n\nНапиши /start, чтобы увидеть кнопку приложения.")
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Ошибка отправки сообщения: %v", err)
	}
}
