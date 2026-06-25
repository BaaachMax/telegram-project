package main

import (
    "encoding/json"
    "log"
    "os"
    "strconv"

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

// handleStart отправляет приветствие и кнопку для открытия Mini App.
// Используем прямой вызов Bot API (sendMessage с reply_markup в виде JSON),
// потому что установленная версия библиотеки tgbotapi не поддерживает
// поле WebApp в InlineKeyboardButton напрямую.
func handleStart(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
    replyMarkup := map[string]interface{}{
        "inline_keyboard": [][]map[string]interface{}{
            {
                {
                    "text": "Открыть приложение",
                    "web_app": map[string]string{
                        "url": miniAppURL,
                    },
                },
            },
        },
    }
    replyMarkupJSON, _ := json.Marshal(replyMarkup)

    params := tgbotapi.Params{}
    params.AddNonEmpty("chat_id", strconv.FormatInt(message.Chat.ID, 10))
    params.AddNonEmpty("text", "👋 Привет! Нажми кнопку ниже, чтобы открыть приложение.")
    params.AddNonEmpty("reply_markup", string(replyMarkupJSON))

    if _, err := bot.MakeRequest("sendMessage", params); err != nil {
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
