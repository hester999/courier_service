package db

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

func InitDataBaseConnection(maxAttempts, delay int) (*sqlx.DB, error) {
	log.Println("Connecting to database ...")
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	dbName := os.Getenv("POSTGRES_DB")
	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, dbName)

	var err error
	var db *sqlx.DB
	db, err = sqlx.Connect("pgx", dsn) // теперь использует pgx
	if err != nil {
		db, err = func() (*sqlx.DB, error) {
			for i := 1; i <= maxAttempts; i++ { // Ретраи через цикл. Количество попыток и их частоту можно установить в конфиге через .env
				if db, err = sqlx.Connect("pgx", dsn); err == nil {
					log.Println("Database connected successfully.")
					return db, nil
				}
				log.Printf("Database not ready (attempt %d/%d): %v", i, maxAttempts, err)
				time.Sleep(time.Second * time.Duration(delay))
			}
			return nil, fmt.Errorf("could not connect to postgres database after %d attempts: %w", maxAttempts, err)
		}()
		if err != nil {
			return nil, fmt.Errorf("could not connect to postgres database: %w", err)
		}
	}

	maxConn, err := strconv.Atoi(os.Getenv("MAX_CONNECTIONS"))
	if err != nil {
		maxConn = 1
	}

	db.SetMaxOpenConns(maxConn)
	return db, nil

}
