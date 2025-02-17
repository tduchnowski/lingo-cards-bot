package main

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type WordEntry struct {
	Id           int64   `db:"id"`
	Word         string  `db:"word"`
	LangCode     string  `db:"lang_code"`
	Language     string  `db:"language"`
	Meaning      *string `db:"meaning"`
	Lemma        *string `db:"lemma"`
	Usage        *string `db:"usage"`
	PartOfSpeech *string `db:"part_of_speech"`
	Frequency    int     `db:"frequency"`
}

func createConnection(url string) (*pgxpool.Pool, error) {
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
