package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/deside01/tg_freelance_bot/internal/middlewares"
	"github.com/deside01/tg_freelance_bot/internal/scraper"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/joho/godotenv"
)

var arr []string

type UserSession struct {
	Page int
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

	token := strings.TrimSpace(os.Getenv("BOT_TOKEN"))

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	opts := []bot.Option{
		bot.WithCallbackQueryDataHandler("button", bot.MatchTypePrefix, callbackHandler),
		// bot.WithCallbackQueryDataHandler("page", bot.MatchTypePrefix, paginationHandler),
	}

	for i := range 21 {
		arr = append(arr, strconv.Itoa(i))
	}

	b, err := bot.New(token, opts...)
	if err != nil {
		panic(err)
	}

	b.RegisterHandler(bot.HandlerTypeMessageText, "start", bot.MatchTypeCommand, kbHandler)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "page", bot.MatchTypePrefix, paginationHandler, middlewares.SingleFlight)

	log.Println("Bot started")
	go scraper.StartScraper(15 * time.Second)
	b.Start(ctx)
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

	newSession := &UserSession{Page: 1}
	sessions[chatID] = newSession

	return newSession
}

func updateSession(chatID int64, page int) {
	session := getUserSession(chatID)
	session.Page = page
}

func kbHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        "Click by button",
		ReplyMarkup: paginationButtons(),
	})
}

func paginationButtons() (keyboard *models.InlineKeyboardMarkup) {
	return &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{
					Text:         "<",
					CallbackData: "page_left",
				},
				{
					Text:         ">",
					CallbackData: "page_right",
				},
			},
		},
	}
}

func pagination(data []string, page, limit int) []string {
	var result []string
	dataLength := len(data)

	if dataLength == 0 {
		return result
	}

	if limit < 1 {
		limit = 10
	}

	if page < 1 {
		page = 1
	}

	start := (page - 1) * limit
	end := start + limit

	if start > dataLength {
		return result
	}

	if end > dataLength {
		end = dataLength
	}

	return data[start:end]
}

func paginationHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		ShowAlert:       false,
	})

	chatID := update.CallbackQuery.Message.Message.Chat.ID
	session := getUserSession(chatID)

	switch update.CallbackQuery.Data {
	case "page_left":
		if session.Page > 1 {
			session.Page--
		}
	case "page_right":
		session.Page++
	}

	data := pagination(arr, session.Page, 10)

	if session.Page*len(data) == 0 {
		session.Page--
		return
	}

	updateSession(chatID, session.Page)

	b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:      update.CallbackQuery.Message.Message.Chat.ID,
		MessageID:   update.CallbackQuery.Message.Message.ID,
		Text:        strings.Join(data, "\n"),
		ReplyMarkup: paginationButtons(),
	})
}
