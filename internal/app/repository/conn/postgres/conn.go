package rcpostgres

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/KDarenskii/order-service/internal/app/config/section"
)

type Client struct {
	db  *gorm.DB
	cfg section.RepositoryPostgres
}

func (c *Client) DB() *gorm.DB {
	return c.db
}

func (c *Client) Close() error {
	db, err := c.db.DB()
	if err != nil {
		return fmt.Errorf("failed to close PostgreSQL connection: %w", err)
	}

	if err := db.Close(); err != nil {
		return fmt.Errorf("failed to close PostgreSQL connection: %w", err)
	}

	return nil
}

func NewClient(ctx context.Context, cfg section.RepositoryPostgres) (*Client, error) {
	var u url.URL

	u.Scheme = "postgres"
	u.Host = cfg.Address
	u.User = url.UserPassword(cfg.Username, cfg.Password)
	u.Path = cfg.Name

	args := make(url.Values)
	args.Set("sslmode", "disable")
	u.RawQuery = args.Encode()

	dsn := u.String()

	log.Printf("Initializing PostgreSQL connection: add=%s", cfg.Address)

	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		return nil, fmt.Errorf("failed to open PostgreSQL connection: %w", err)
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to open PostgreSQL connection: %w", err)
	}

	sqlDB.SetMaxOpenConns(10)

	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)

	defer cancel()

	if err := sqlDB.PingContext(pingCtx); err != nil {
		_ = sqlDB.Close()

		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	log.Println("PostgreSQL connection established")

	return &Client{db: gormDB, cfg: cfg}, nil
}
