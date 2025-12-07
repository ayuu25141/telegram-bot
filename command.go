// package main

// import (
// 	"fmt"
// 	"log"
// 	"strings"
// 		"github.com/joho/godotenv"
// "os"

// 	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
// )

// func main() {
// 	// 🔐 Replace with your credentials
// 		// Load the .env file
// 	err := godotenv.Load()
// 	if err != nil {
// 		log.Fatal("❌ Error loading .env file")
// 	}
// 	botToken :=  os.Getenv("Bt")
// 	adminID := int64(6386418509) // your own Telegram user ID

// 	bot, err := tgbotapi.NewBotAPI(botToken)
// 	if err != nil {
// 		log.Fatal("Bot init error:", err)
// 	}

// 	bot.Debug = false
// 	log.Printf("🤖 Bot @%s started successfully", bot.Self.UserName)



// 	// 🛰️ Update configuration
// 	u := tgbotapi.NewUpdate(0)
// 	u.Timeout = 60
// 	updates := bot.GetUpdatesChan(u)

// 	for update := range updates {
// 		if update.Message == nil {
// 			continue
// 		}

// 		chatID := update.Message.Chat.ID
// 		text := update.Message.Text

// 		// 👮‍♂️ Only admin can upload or register videos
// 		if update.Message.Video != nil {
// 			if chatID != adminID {
// 				bot.Send(tgbotapi.NewMessage(chatID, "🚫 Sorry, you are not allowed to upload media."))
// 				continue
// 			}

// 			fileID := update.Message.Video.FileID
// 			caption := strings.TrimSpace(update.Message.Caption)
// 			if caption == "" {
// 				caption = fmt.Sprintf("video%d", len(videos)+1)
// 			}

// 			videos[caption] = fileID

// 			link := fmt.Sprintf("https://t.me/%s?start=%s", bot.Self.UserName, caption)
// 			msg := fmt.Sprintf("✅ Video saved!\n\n🎬 *Payload:* `%s`\n🪪 *FileID:* `%s`\n🔗 *Deep Link:*\n%s",
// 				caption, fileID, link)

// 			reply := tgbotapi.NewMessage(chatID, msg)
// 			reply.ParseMode = "Markdown"
// 			bot.Send(reply)

// 			log.Printf("💾 Admin uploaded '%s' (%s)", caption, fileID)
// 			continue
// 		}

// 		// 🚫 Block all other user commands except deep-links
// 		if strings.HasPrefix(text, "/") && !strings.HasPrefix(text, "/start") {
// 			bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Commands are disabled for users."))
// 			continue
// 		}

// 		// 🎯 Handle deep link like /start video1
// 		if strings.HasPrefix(text, "/start") {
// 			payload := strings.TrimSpace(strings.TrimPrefix(text, "/start"))

// 			// Case: user clicked bot link without payload
// 			if payload == "" {
// 				if chatID == adminID {
// 					msg := "👋 Welcome, Admin!\nYou can upload videos directly to save them."
// 					bot.Send(tgbotapi.NewMessage(chatID, msg))
// 				} else {
// 					bot.Send(tgbotapi.NewMessage(chatID, "👋 Welcome! You can only access videos via special links."))
// 				}
// 				continue
// 			}

// 			// Case: payload found
// 			fileID, exists := videos[payload]
// 			if !exists {
// 				bot.Send(tgbotapi.NewMessage(chatID, "❌ No video found for: "+payload))
// 				continue
// 			}

// 			video := tgbotapi.NewVideo(chatID, tgbotapi.FileID(fileID))
// 			video.Caption = fmt.Sprintf("🎬 Here’s your video: %s")
// 			_, err := bot.Send(video)
// 			if err != nil {
// 				log.Println("❗ Error sending video:", err)
// 				bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Failed to send video, please try again later."))
// 			}
// 			continue
// 		}

// 		// 🧍 For normal users who try to chat or send anything else
// 		if chatID != adminID {
// 			bot.Send(tgbotapi.NewMessage(chatID, "💡 Please use the link provided to access videos."))
// 			continue
// 		}

// 		// 👑 Admin default message
// 		bot.Send(tgbotapi.NewMessage(chatID, "✅ Admin mode active. You can upload videos with captions to name them."))
// 	}
// }



package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	// "github.com/joho/godotenv"
	_ "github.com/lib/pq"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	// Load environment variables (.env)
	// err := godotenv.Load()
	// if err != nil {
	// 	log.Fatal("❌ Error loading .env file")
	// }

	// Connect to PostgreSQL
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("❌ Failed to connect DB:", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal("❌ Database not reachable:", err)
	}
	fmt.Println("✅ Connected to PostgreSQL Database succesfully")

	// Telegram bot setup
	botToken := os.Getenv("Bt")
	adminID := int64(6386418509) // your Telegram ID (admin)
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatal("❌ Bot init failed:", err)
	}

	log.Printf("🤖 Bot @%s started", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		text := update.Message.Text
		chatID := update.Message.Chat.ID

		// Handle deep link: /start <deeplinkid>
		if strings.HasPrefix(text, "/start") {
			payload := strings.TrimSpace(strings.TrimPrefix(text, "/start"))

			// No payload → greeting message
			if payload == "" {
				if chatID == adminID {
					bot.Send(tgbotapi.NewMessage(chatID, "👋 Welcome, Admin! You can upload and manage videos."))
				} else {
					bot.Send(tgbotapi.NewMessage(chatID, "👋 Welcome! Please use a valid video link."))
				}
				continue
			}

			// Fetch from DB by deeplinkid
			var videourl string
			err := db.QueryRow(`SELECT videourl FROM videodata WHERE deeplinkid = $1`, payload).Scan(&videourl)

			if err != nil {
				if err == sql.ErrNoRows {
					bot.Send(tgbotapi.NewMessage(chatID, "❌ Invalid or expired video link."))
				} else {
					log.Println("DB error:", err)
					bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Database error. Please try later."))
				}
				continue
			}

			// Send the video by its file_id
			video := tgbotapi.NewVideo(chatID, tgbotapi.FileID(videourl))
			video.Caption = "🎬 Here’s your video!"
			_, err = bot.Send(video)
			if err != nil {
				log.Println("❗ Error sending video:", err)
				bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Could not send video. Try again later."))
			}
			continue
		}

		// Restrict users from sending random messages
		if chatID != adminID {
			bot.Send(tgbotapi.NewMessage(chatID, "💡 Please use a valid bot link to access your video."))
			continue
		}

		// Admin default response
		bot.Send(tgbotapi.NewMessage(chatID, "✅ Admin mode active."))
	}
}

