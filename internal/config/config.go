package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultHTTPPort          = 8080
	defaultDBHost            = "localhost"
	defaultDBPort            = 5432
	defaultDBUser            = "postgres"
	defaultDBPassword        = "postgres"
	defaultDBName            = "flysoft_flight_service"
	defaultDBSSLMode         = "disable"
	defaultLogLevel          = "info"
	defaultOfferTTL          = 30 * time.Minute
	defaultCommissionPercent = int64(5)
	defaultServiceFeeAdult   = int64(1500)
	defaultServiceFeeChild   = int64(1000)
	defaultServiceFeeInfant  = int64(0)
)

type Config struct {
	HTTPPort          int
	DB                DBConfig
	LogLevel          string
	OfferTTL          time.Duration
	CommissionPercent int64
	ServiceFees       ServiceFeeConfig
}

type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
}

func (c DBConfig) DSN() string {
	user := url.User(c.User)
	if c.Password != "" {
		user = url.UserPassword(c.User, c.Password)
	}

	dsn := url.URL{
		Scheme: "postgres",
		User:   user,
		Host:   net.JoinHostPort(c.Host, strconv.Itoa(c.Port)),
		Path:   c.Name,
	}

	query := dsn.Query()
	query.Set("sslmode", c.SSLMode)
	dsn.RawQuery = query.Encode()

	return dsn.String()
}

type ServiceFeeConfig struct {
	Adult  int64
	Child  int64
	Infant int64
}

func Load() (Config, error) {
	httpPort, err := envInt("HTTP_PORT", defaultHTTPPort)
	if err != nil {
		return Config{}, err
	}

	dbPort, err := envInt("DB_PORT", defaultDBPort)
	if err != nil {
		return Config{}, err
	}

	offerTTL, err := envDuration("OFFER_TTL", defaultOfferTTL)
	if err != nil {
		return Config{}, err
	}

	commissionPercent, err := envInt64("COMMISSION_PERCENT", defaultCommissionPercent)
	if err != nil {
		return Config{}, err
	}

	serviceFeeAdult, err := envInt64("SERVICE_FEE_ADULT", defaultServiceFeeAdult)
	if err != nil {
		return Config{}, err
	}

	serviceFeeChild, err := envInt64("SERVICE_FEE_CHILD", defaultServiceFeeChild)
	if err != nil {
		return Config{}, err
	}

	serviceFeeInfant, err := envInt64("SERVICE_FEE_INFANT", defaultServiceFeeInfant)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		HTTPPort: httpPort,
		DB: DBConfig{
			Host:     envString("DB_HOST", defaultDBHost),
			Port:     dbPort,
			User:     envString("DB_USER", defaultDBUser),
			Password: envString("DB_PASSWORD", defaultDBPassword),
			Name:     envString("DB_NAME", defaultDBName),
			SSLMode:  envString("DB_SSLMODE", defaultDBSSLMode),
		},
		LogLevel:          strings.ToLower(envString("LOG_LEVEL", defaultLogLevel)),
		OfferTTL:          offerTTL,
		CommissionPercent: commissionPercent,
		ServiceFees: ServiceFeeConfig{
			Adult:  serviceFeeAdult,
			Child:  serviceFeeChild,
			Infant: serviceFeeInfant,
		},
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c Config) Validate() error {
	if err := validatePort("HTTP_PORT", c.HTTPPort); err != nil {
		return err
	}
	if err := validatePort("DB_PORT", c.DB.Port); err != nil {
		return err
	}

	for key, value := range map[string]string{
		"DB_HOST":    c.DB.Host,
		"DB_USER":    c.DB.User,
		"DB_NAME":    c.DB.Name,
		"DB_SSLMODE": c.DB.SSLMode,
		"LOG_LEVEL":  c.LogLevel,
	} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s must not be empty", key)
		}
	}

	switch c.DB.SSLMode {
	case "disable", "allow", "prefer", "require", "verify-ca", "verify-full":
	default:
		return fmt.Errorf("DB_SSLMODE must be one of disable, allow, prefer, require, verify-ca, verify-full")
	}

	switch c.LogLevel {
	case "debug", "info", "warn", "error":
	default:
		return fmt.Errorf("LOG_LEVEL must be one of debug, info, warn, error")
	}

	if c.OfferTTL <= 0 {
		return fmt.Errorf("OFFER_TTL must be positive")
	}
	if c.CommissionPercent < 0 {
		return fmt.Errorf("COMMISSION_PERCENT must be non-negative")
	}
	if c.ServiceFees.Adult < 0 || c.ServiceFees.Child < 0 || c.ServiceFees.Infant < 0 {
		return fmt.Errorf("service fees must be non-negative")
	}

	return nil
}

func envString(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func envInt(key string, fallback int) (int, error) {
	value := envString(key, strconv.Itoa(fallback))
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", key, err)
	}
	return parsed, nil
}

func envInt64(key string, fallback int64) (int64, error) {
	value := envString(key, strconv.FormatInt(fallback, 10))
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", key, err)
	}
	return parsed, nil
}

func envDuration(key string, fallback time.Duration) (time.Duration, error) {
	value := envString(key, fallback.String())
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be a duration: %w", key, err)
	}
	return parsed, nil
}

func validatePort(key string, port int) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("%s must be between 1 and 65535", key)
	}
	return nil
}
