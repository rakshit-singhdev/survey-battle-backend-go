package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port           string
	NodeEnv        string
	MongoURI       string
	RedisURL       string
	AllowedOrigins []string

	JWT       JWTConfig
	Cookie    CookieConfig
	AI        AIConfig
	FastMoney FastMoneyConfig
	Redis     RedisConfig

	MinQuestionsToPublish int
	RefreshTokenGrace     time.Duration
}

type JWTConfig struct {
	AccessSecret  string
	RefreshSecret string
	PlayerSecret  string

	AccessExpiresIn  time.Duration
	RefreshExpiresIn time.Duration
	PlayerExpiresIn  time.Duration
}

type CookieConfig struct {
	Domain   string
	Secure   bool
	SameSite string
}

type AIConfig struct {
	Provider string
	Model    string
	APIKey   string
}

type FastMoneyConfig struct {
	PlayerTime  time.Duration
	TargetScore int
}

type RedisConfig struct {
	GameStateTTL time.Duration
	GameLockTTL  time.Duration
}

func Load() (*Config, error) {
	// Load .env if present.
	// In production, environment variables can be supplied directly.
	_ = godotenv.Load()

	config := &Config{
		Port:     getEnv("PORT", "4000"),
		NodeEnv:  getEnv("NODE_ENV", "development"),
		MongoURI: os.Getenv("MONGO_URI"),
		RedisURL: os.Getenv("REDIS_URL"),

		AllowedOrigins: getList("ALLOWED_ORIGINS"),

		JWT: JWTConfig{
			AccessSecret:     os.Getenv("JWT_ACCESS_SECRET"),
			RefreshSecret:    os.Getenv("JWT_REFRESH_SECRET"),
			PlayerSecret:     os.Getenv("JWT_PLAYER_SECRET"),
			AccessExpiresIn:  getDuration("JWT_ACCESS_EXPIRES_IN", 15*time.Minute),
			RefreshExpiresIn: getDuration("JWT_REFRESH_EXPIRES_IN", 7*24*time.Hour),
			PlayerExpiresIn:  getDuration("JWT_PLAYER_EXPIRES_IN", 12*time.Hour),
		},

		Cookie: CookieConfig{
			Domain:   os.Getenv("COOKIE_DOMAIN"),
			Secure:   getBool("COOKIE_SECURE", false),
			SameSite: getEnv("COOKIE_SAME_SITE", "Lax"),
		},

		AI: AIConfig{
			Provider: os.Getenv("LLM_PROVIDER"),
			Model:    os.Getenv("LLM_MODEL"),
			APIKey:   os.Getenv("LLM_API_KEY"),
		},

		FastMoney: FastMoneyConfig{
			PlayerTime: getDuration(
				"FAST_MONEY_PLAYER_TIME",
				30*time.Second,
			),
			TargetScore: getInt(
				"FAST_MONEY_TARGET_SCORE",
				200,
			),
		},

		Redis: RedisConfig{
			GameStateTTL: getDuration(
				"REDIS_GAME_STATE_TTL",
				24*time.Hour,
			),
			GameLockTTL: getDuration(
				"REDIS_GAME_LOCK_TTL",
				3*time.Second,
			),
		},

		MinQuestionsToPublish: getInt(
			"MIN_QUESTIONS_TO_PUBLISH",
			5,
		),

		RefreshTokenGrace: getDuration(
			"REFRESH_TOKEN_GRACE_SECONDS",
			30*time.Second,
		),
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

func (c *Config) Validate() error {
	var missing []string

	if c.MongoURI == "" {
		missing = append(missing, "MONGO_URI")
	}

	if c.RedisURL == "" {
		missing = append(missing, "REDIS_URL")
	}

	if c.JWT.AccessSecret == "" {
		missing = append(missing, "JWT_ACCESS_SECRET")
	}

	if c.JWT.RefreshSecret == "" {
		missing = append(missing, "JWT_REFRESH_SECRET")
	}

	if c.JWT.PlayerSecret == "" {
		missing = append(missing, "JWT_PLAYER_SECRET")
	}

	if len(missing) > 0 {
		return fmt.Errorf(
			"missing required environment variables: %s",
			strings.Join(missing, ", "),
		)
	}

	return nil
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)

	if value == "" {
		return fallback
	}

	return value
}

func getInt(key string, fallback int) int {
	value := os.Getenv(key)

	if value == "" {
		return fallback
	}

	result, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return result
}

func getBool(key string, fallback bool) bool {
	value := os.Getenv(key)

	if value == "" {
		return fallback
	}

	result, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}

	return result
}

func getDuration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)

	if value == "" {
		return fallback
	}

	// Support plain seconds for variables such as:
	// FAST_MONEY_PLAYER_TIME=30
	// REFRESH_TOKEN_GRACE_SECONDS=30
	if seconds, err := strconv.Atoi(value); err == nil {
		return time.Duration(seconds) * time.Second
	}

	// Also support Go durations:
	// 30s, 15m, 24h
	duration, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}

	return duration
}

func getList(key string) []string {
	value := os.Getenv(key)

	if value == "" {
		return nil
	}

	parts := strings.Split(value, ",")

	result := make([]string, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)

		if part != "" {
			result = append(result, part)
		}
	}

	return result
}
