package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// ===== Конфигурация =====

var botToken = os.Getenv("BOT_TOKEN") // токен бота, полученный от @BotFather

// ===== Валидация initData =====
//
// Telegram присылает initData как querystring. Алгоритм проверки описан тут:
// https://core.telegram.org/bots/webapps#validating-data-received-via-the-mini-app
func validateInitData(initData string, maxAge time.Duration) (url.Values, bool) {
	values, err := url.ParseQuery(initData)
	if err != nil {
		return nil, false
	}

	hash := values.Get("hash")
	if hash == "" {
		return nil, false
	}
	values.Del("hash")

	// Собираем data_check_string: все пары key=value, отсортированные по key, через \n
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, k+"="+values.Get(k))
	}
	dataCheckString := strings.Join(parts, "\n")

	// secret_key = HMAC_SHA256("WebAppData", bot_token)
	secretKeyMac := hmac.New(sha256.New, []byte("WebAppData"))
	secretKeyMac.Write([]byte(botToken))
	secretKey := secretKeyMac.Sum(nil)

	// итоговый хэш = HMAC_SHA256(secret_key, data_check_string)
	h := hmac.New(sha256.New, secretKey)
	h.Write([]byte(dataCheckString))
	calculatedHash := hex.EncodeToString(h.Sum(nil))

	if calculatedHash != hash {
		return nil, false
	}

	// Доп. проверка: не протухло ли auth_date (защита от replay-атак)
	if maxAge > 0 {
		if authDateStr := values.Get("auth_date"); authDateStr != "" {
			if authDate, err := strconv.ParseInt(authDateStr, 10, 64); err == nil {
				if time.Since(time.Unix(authDate, 0)) > maxAge {
					return nil, false
				}
			}
		}
	}

	return values, true
}

// ===== Middleware: проверка initData из заголовка =====

type ctxKey string

const userCtxKey ctxKey = "tg_user"

func telegramAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		initData := r.Header.Get("X-Telegram-Init-Data")
		if initData == "" {
			http.Error(w, "missing init data", http.StatusUnauthorized)
			return
		}

		values, ok := validateInitData(initData, 24*time.Hour)
		if !ok {
			http.Error(w, "invalid init data", http.StatusUnauthorized)
			return
		}

		// Пользовательские данные лежат в поле "user" как JSON-строка
		userJSON := values.Get("user")
		log.Printf("authenticated request, user payload: %s", userJSON)

		next(w, r)
	}
}

// ===== Хэндлеры =====

func handlePing(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Пример защищённого эндпоинта — доступен только с валидным initData
func handleMe(w http.ResponseWriter, r *http.Request) {
	initData := r.Header.Get("X-Telegram-Init-Data")
	values, _ := validateInitData(initData, 24*time.Hour)

	var user map[string]interface{}
	_ = json.Unmarshal([]byte(values.Get("user")), &user)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"user": user,
	})
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

// ===== CORS (нужен, если фронтенд раздаётся отдельно от backend) =====

func withCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Telegram-Init-Data")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}
}

func main() {
	if botToken == "" {
		log.Println("ВНИМАНИЕ: переменная окружения BOT_TOKEN не установлена. Установите её перед запуском:")
		log.Println(`  export BOT_TOKEN="ваш_токен_от_BotFather"`)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/ping", withCORS(handlePing))
	mux.HandleFunc("/api/me", withCORS(telegramAuthMiddleware(handleMe)))

	addr := ":8080"
	log.Printf("Backend запущен на %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
