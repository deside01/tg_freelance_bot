package middlewares

import (
	"context"
	"sync"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func SingleFlight(next bot.HandlerFunc) bot.HandlerFunc {
	sf := sync.Map{}
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		if update.CallbackQuery != nil {
			key := update.CallbackQuery.Message.Message.ID
			if _, loaded := sf.LoadOrStore(key, struct{}{}); loaded {
				return
			}
			defer sf.Delete(key)
			next(ctx, b, update)
		}
	}
}