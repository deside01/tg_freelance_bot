package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/deside01/tg_freelance_bot/internal/config"
	"github.com/deside01/tg_freelance_bot/internal/scraper"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/joho/godotenv"
)

type UserSession struct {
	Checker context.CancelFunc
	Page    int
}

var (
	sessions     = make(map[int64]*UserSession)
	sessionMutex = &sync.RWMutex{}
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	config.SetupDB()

	token := strings.TrimSpace(os.Getenv("BOT_TOKEN"))

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	opts := []bot.Option{
		bot.WithCallbackQueryDataHandler("button", bot.MatchTypePrefix, callbackHandler),
		// bot.WithCallbackQueryDataHandler("page", bot.MatchTypePrefix, paginationHandler),
	}

	b, err := bot.New(token, opts...)
	if err != nil {
		panic(err)
	}

	b.RegisterHandler(bot.HandlerTypeMessageText, "start", bot.MatchTypeCommand, checkerHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "stop", bot.MatchTypeCommand, cancelHandler)

	log.Println("Bot started")

	b.Start(ctx)
}

func checkerHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	// config.DB.ClearOrders(ctx)

	checkerCtx, cancel := context.WithCancel(context.Background())
	session := getUserSession(update.Message.Chat.ID)
	session.Checker = cancel

	data, err := scraper.GetOrders2()
	if err != nil {
		log.Fatalf("gg: %v", err)
	}

	log.Println(len(data))

	for _, v := range data {
		select {
		case <-checkerCtx.Done():
			_, err := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   "cancel",
			})
			if err != nil {
				log.Println("hz", err)
			}

			return
		default:
			_, err := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   v.Title,
			})

			if err != nil {
				log.Fatalf("err: %v", err)
			}

			time.Sleep(5 * time.Second)
		}
	}
}

func cancelHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	session := getUserSession(update.Message.Chat.ID)

	if session.Checker != nil {
		session.Checker()
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "lasx",
		})
	}
}

func callbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		ShowAlert:       false,
	})

	b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:    update.CallbackQuery.Message.Message.Chat.ID,
		MessageID: update.CallbackQuery.Message.Message.ID,
		Text:      "You selected the button: " + update.CallbackQuery.Data,
	})
}

func getUserSession(chatID int64) *UserSession {
	sessionMutex.Lock()
	defer sessionMutex.Unlock()

	session, ok := sessions[chatID]
	if ok {
		return session
	}

	newSession := &UserSession{
		Page: 1,
	}
	sessions[chatID] = newSession

	return newSession
}

// func pagination[T any](data []T, page, limit int) []T {
// 	var result []T
// 	dataLength := len(data)

// 	if dataLength == 0 {
// 		return result
// 	}

// 	if limit < 1 {
// 		limit = 10
// 	}

// 	if page < 1 {
// 		page = 1
// 	}

// 	start := (page - 1) * limit
// 	end := start + limit

// 	if start > dataLength {
// 		return result
// 	}

// 	if end > dataLength {
// 		end = dataLength
// 	}

// 	return data[start:end]
// }
