package database

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type WordEntry struct {
	Lemma        string  `db:"lemma"`
	LemmaMeaning string  `db:"lemma_meaning"`
	Sentences    *string `db:"sentences"`
}

type Example struct {
	Sentence    string `xml:"sentence"`
	Translation string `xml:"translation"`
}

type Examples struct {
	Sentences []Example `xml:"example"`
}

func CreateConnection(url string) (*pgxpool.Pool, error) {
	dbCfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, err
	}
	dbCfg.MaxConns = 100
	dbCfg.MinConns = 10
	dbCfg.MaxConnLifetime = time.Hour
	dbCfg.MaxConnIdleTime = time.Hour
	dbCfg.HealthCheckPeriod = time.Minute * 5
	return pgxpool.NewWithConfig(context.Background(), dbCfg)
}

// TODO:
func Ping(db *pgxpool.Pool) error {
	return nil
}
