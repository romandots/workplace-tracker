package app

import (
	"context"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// App holds external connections used by handlers.
type App struct {
	DB    *pgxpool.Pool
	Redis *redis.Client
	Bot   *tgbotapi.BotAPI
}

// New creates a new App from environment variables.
func New(ctx context.Context) (*App, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://app:password@localhost:5432/app?sslmode=disable"
	}
	db, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return nil, err
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	var bot *tgbotapi.BotAPI
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken != "" {
		bot, err = tgbotapi.NewBotAPI(botToken)
		if err != nil {
			return nil, err
		}
	}

	return &App{DB: db, Redis: rdb, Bot: bot}, nil
}

// Close closes external connections.
func (a *App) Close() {
	if a.DB != nil {
		a.DB.Close()
	}
	if a.Redis != nil {
		a.Redis.Close()
	}
}
