package config

import (
	"encoding/hex"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

var config *Config

const (
	DEFAULT_HTTP_CLIENT_TIMEOUT     = 10 * time.Second
	DEFAULT_WATERMILL_MAX_IDLE_TIME = 30 * time.Minute
	DEFAULT_NUM_WORKER              = 10
)

type Config struct {
	ServerConfig  ServerConfig
	ServiceConfig ServiceConfig
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type ServerConfig struct {
	// Debug string
	HTTP HTTPConfig
	// Middleware MiddlewareConfig
}

type HTTPConfig struct {
	Port              string
	ConnectionTimeout time.Duration
	// ReadTimeout       int
	// WriteTimeout      int
	// ProxyURL string
}

type MiddlewareConfig struct {
	AllowCors string
}

type ServiceConfig struct {
	Redis RedisConfig
}

func GetConfig() *Config {
	if config != nil {
		return config
	}
	godotenv.Load()

	doInit()

	return config
}

func GetConfigWithFilename(envFileName string) *Config {

	if godotenv.Load(envFileName) == nil {
		goto INIT_CONFIG
	}
	if godotenv.Load(fmt.Sprintf("../%s", envFileName)) == nil {
		goto INIT_CONFIG
	}
	if godotenv.Load(fmt.Sprintf("../../%s", envFileName)) == nil {
		goto INIT_CONFIG
	}
	if godotenv.Load(fmt.Sprintf("../../../%s", envFileName)) == nil {
		goto INIT_CONFIG
	}

	if godotenv.Load(fmt.Sprintf("../../../../%s", envFileName)) == nil {
		goto INIT_CONFIG
	}
	if godotenv.Load(fmt.Sprintf("../../../../../%s", envFileName)) == nil {
		goto INIT_CONFIG
	}

	log.Fatalln("failed to load .env file")
INIT_CONFIG:

	doInit()
	return config
}

func doInit() {
	config = &Config{
		ServerConfig: ServerConfig{
			HTTP: HTTPConfig{
				Port:              getEnvString("HTTP_PORT", "8888"),
				ConnectionTimeout: getEnvDurationFromSeconds("HTTP_TIMEOUT_SEC", DEFAULT_HTTP_CLIENT_TIMEOUT),
			},
		},
		ServiceConfig: ServiceConfig{
			Redis: RedisConfig{
				Addr:     getEnvString("REDIS_ADDR", "localhost:6379"),
				Password: getEnvString("REDIS_PASSWORD", ""),
				DB:       getEnvInt("REDIS_DB", 0),
			},
		},
	}
}

func Init() {
	GetConfig()
}

//lint:ignore U1000 Ignore unused code, it may require in the future
func getEnvString(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	return value
}

func getEnvDurationString(key string, defaultValue time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	// Parse the duration string
	duration, err := time.ParseDuration(value)
	if err != nil {
		fmt.Println("Error parsing duration:", err)
		return defaultValue
	}
	return duration
}

func getEnvDurationFromHours(key string, defaultValue time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	return time.Duration(intValue) * time.Hour
}

func getEnvDurationFromMillisecond(key string, defaultValue time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	return time.Duration(intValue) * time.Millisecond
}

//lint:ignore U1000 Ignore unused code, it may require in the future
func getEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	return intValue
}

func getEnvFloat64(key string, defaultValue float64) float64 {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	floatValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return defaultValue
	}

	return floatValue
}

//lint:ignore U1000 Ignore unused code, it may require in the future
func getEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}

	return boolValue
}

//lint:ignore U1000 Ignore unused code, it may require in the future
func getEnvStringArray(key string, defaultValue []string) []string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	result := strings.Split(value, ",")
	for i := range result {
		result[i] = strings.TrimSpace(result[i])
	}

	return result
}

//lint:ignore U1000 Ignore unused code, it may require in the future
func getEnvDurationFromSeconds(key string, defaultValue time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	intValue, err := strconv.ParseInt(value, 10, 64)

	if err != nil {
		return defaultValue
	}

	return time.Duration(intValue) * time.Second
}

func getEnvDurationFromMinutes(key string, defaultValue time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	intValue, err := strconv.ParseInt(value, 10, 64)

	if err != nil {
		return defaultValue
	}

	return time.Duration(intValue) * time.Minute
}

//lint:ignore U1000 Ignore unused code, it may require in the future
func getEnvDurationFromSecondsNullable(key string, defaultValue time.Duration) *time.Duration {
	value := os.Getenv(key)
	if value == "" {
		if defaultValue == 0 {
			return nil
		} else {
			return &defaultValue
		}
	}

	intValue, err := strconv.ParseInt(value, 10, 64)

	if err != nil {
		return &defaultValue
	}

	result := time.Duration(intValue) * time.Second
	return &result
}

//lint:ignore U1000 Ignore unused code, it may require in the future
func getEnvBytes(key string, defaultValue []byte) []byte {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	decodedByteArray, err := hex.DecodeString(value)
	if err != nil {
		return defaultValue
	}

	return decodedByteArray
}

//lint:ignore U1000 Ignore unused code, it may require in the future
func getEnvLogLevel(key string, defaultValue slog.Level) slog.Level {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	switch value {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return defaultValue
	}
}
