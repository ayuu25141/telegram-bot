package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	_ "github.com/lib/pq"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var bot *tgbotapi.BotAPI
var db *sql.DB

func main() {
	var err error

	// Connect to PostgreSQL
	dbURL := strings.TrimSpace(os.Getenv("DB_URL"))
	db, err = sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("❌ Failed to connect DB:", err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatal("❌ DB unreachable:", err)
	}
	fmt.Println("✅ Connected to PostgreSQL")

	// Setup Bot
	botToken := strings.TrimSpace(os.Getenv("BT")) // remove hidden newlines/spaces
	bot, err = tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatal("❌ Bot init failed:", err)
	}
	log.Printf("🤖 Bot @%s Webhook Mode Started", bot.Self.UserName)

	// Set webhook
	webhookURL := strings.TrimSpace("https://telegram-bot-kzen.onrender.com/webhook")
webhookConfig := tgbotapi.NewWebhook(webhookURL)
_, err = bot.Request(webhookConfig)
if err != nil {
    log.Fatal("❌ Failed to set webhook:", err)
}

	// Webhook handler
	http.HandleFunc("/webhook", HandleWebhook)

	// Render auto provides PORT
	port := os.Getenv("PORT")
	if port == "" {
		port = "10000"
	}

	log.Println("🚀 Server running on port", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func HandleWebhook(w http.ResponseWriter, r *http.Request) {
	var update tgbotapi.Update
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		log.Println("Invalid update:", err)
		return
	}

	if update.Message == nil {
		return
	}

	chatID := update.Message.Chat.ID
	text := update.Message.Text
	adminID := int64(6386418509)

	// Handle /start <deeplink>
	if strings.HasPrefix(text, "/start") {
		payload := strings.TrimSpace(strings.TrimPrefix(text, "/start"))

		if payload == "" {
			msg := "👋 Welcome! Use a valid video link."
			if chatID == adminID {
				msg = "👋 Welcome Admin! Upload videos freely."
			}
			bot.Send(tgbotapi.NewMessage(chatID, msg))
			return
		}

		var videourl string
		err := db.QueryRow(`SELECT videourl FROM videodata WHERE deeplinkid=$1`, payload).Scan(&videourl)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "❌ Invalid or expired video link."))
			return
		}

		video := tgbotapi.NewVideo(chatID, tgbotapi.FileID(videourl))
		video.Caption = "🎬 Here’s your video!"
		bot.Send(video)
		return
	}

	// Restrict normal users
	if chatID != adminID {
		bot.Send(tgbotapi.NewMessage(chatID, "💡 Use a valid bot link to access your video."))
		return
	}

	// Admin default response
	bot.Send(tgbotapi.NewMessage(chatID, "✅ Admin mode active."))
}

