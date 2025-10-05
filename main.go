package main

import (
	"context"
	"errors"
	"fmt"
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
	// Page          int
	CancelChannel chan struct{}
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
	b.RegisterHandler(bot.HandlerTypeMessageText, "clear", bot.MatchTypeCommand, clearHandler)

	log.Println("Bot started")

	b.Start(ctx)
}

func clearHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	err := config.DB.ClearOrders(ctx)
	if err != nil {
		log.Printf("err: %v", err)
		return
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Таблица успешно очищена",
	})
}

func checkerHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	chatID := update.Message.Chat.ID

	session := getUserSession(chatID)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// getOrders(ctx, b, update)

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			go getOrders(ctx, b, update)
		case <-session.CancelChannel:
			log.Println("закончили2")
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text:   "Парсинг остановлен",
			})
			return
		}
	}
}

func getOrders(ctx context.Context, b *bot.Bot, update *models.Update) {
	err := check(ctx, b, update)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return
		}

		log.Printf("check err: %v", err)
	}
}

func check(ctx context.Context, b *bot.Bot, update *models.Update) error {
	chatID := update.Message.Chat.ID

	data, err := scraper.GetOrders2()
	if err != nil {
		log.Fatalf("gg: %v", err)
	}

	if len(data) > 0 {
		_, err = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   fmt.Sprintf("Найдено %v новых заказов!", len(data)),
		})
		if err != nil {
			log.Printf("lenErr: %v", err)
		}
	}

	for _, v := range data {
		text := fmt.Sprintf("*%v*\n%v\n[Ссылка](%v)", v.Title, v.Description, v.Link)

		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    chatID,
			Text:      text,
			ParseMode: models.ParseModeMarkdownV1,
			LinkPreviewOptions: &models.LinkPreviewOptions{
				IsDisabled: bot.True(),
			},
		})

		if err != nil {
			return fmt.Errorf("err: %v", err)
		}

		select {
		case <-time.After(3 * time.Second):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return nil
}

func cancelHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	session := getUserSession(update.Message.Chat.ID)
	sessionMutex.Lock()
	defer sessionMutex.Unlock()

	close(session.CancelChannel)
	session.CancelChannel = make(chan struct{})
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
		// Page:          1,
		CancelChannel: make(chan struct{}, 1),
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
