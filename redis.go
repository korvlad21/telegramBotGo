package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

// Redis структура для хранения подключения
type Redis struct {
	Client *redis.Client
	ctx    context.Context
}

// NewRedis функция для инициализации подключения к Redis
func NewRedis() *Redis {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	client := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "", // Пароль (если есть)
		DB:       0,  // Используемая БД (по умолчанию 0)
	})

	ctx := context.Background()

	// Проверяем подключение
	pong, err := client.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Ошибка подключения к Redis: %v", err)
	}
	fmt.Println("Redis подключен:", pong)

	return &Redis{Client: client, ctx: ctx}
}

// Set записывает значение в Redis
func (r *Redis) Set(key, value string) error {
	return r.Client.Set(r.ctx, key, value, 0).Err()
}

// Get получает значение из Redis
func (r *Redis) Get(key string) (string, error) {
	return r.Client.Get(r.ctx, key).Result()
}

// Close закрывает соединение с Redis
func (r *Redis) Close() {
	if err := r.Client.Close(); err != nil {
		log.Printf("Ошибка при закрытии соединения с Redis: %v", err)
	}
}

func (r *Redis) ListPush(key, value string) error {
	return r.Client.RPush(r.ctx, key, value).Err()
}

// ListGetAll получает все элементы списка по ключу
func (r *Redis) ListGetAll(key string) ([]string, error) {
	return r.Client.LRange(r.ctx, key, 0, -1).Result()
}
