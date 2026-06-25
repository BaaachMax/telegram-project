# Telegram Mini App — стартовый шаблон

## 1. Создание бота
Через @BotFather: /newbot

## 2. Запуск backend
cd backend
export BOT_TOKEN="ваш_токен"
go run main.go

## 3. Запуск frontend
cd frontend
python3 -m http.server 5173

## 4. HTTPS через ngrok
ngrok http 5173

Полученный https://xxxx.ngrok.io указываете в @BotFather как Mini App URL.

## Важно
Никогда не доверяйте initDataUnsafe на фронтенде для решений безопасности — 
все критичные проверки идут через backend с полной валидацией initData (см. main.go).
