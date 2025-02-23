package main

import (
	"context"
	"fmt"
	"math/rand"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Language struct {
	fullName string
	code     string
}

// this is used for translating lang_code from the database to
// a normal name that users see in Telegram
// using ISO 639-1 Code
var languageCodeToName map[string]string = map[string]string{
	"pl":    "Polish",
	"ru":    "Russian",
	"es":    "Spanish",
	"pt":    "Portuguese",
	"pt-br": "Portuguese (Brazilian)",
	"de":    "German",
	"hu":    "Hungarian",
	"ko":    "Korean",
	"ja":    "Japanese",
	"zh-cn": "Chinese",
	"it":    "Italian",
	"sr":    "Serbian",
	"et":    "Estonian",
}

var difficultyToString map[uint8]string = map[uint8]string{
	0: "Most frequent words",
	1: "Pretty common words",
	2: "Less common words"}

var languageTableBaseName string = "words_" // all view tables are of the form words_<language code>_<frequency bin>
var allLangsTableName string = "language_codes"
var privateChatsTableName string = "private_chats"

func getSupportedLanguages(dbPool *pgxpool.Pool) ([]Language, error) {
	query := fmt.Sprintf("SELECT DISTINCT lang_code FROM %s", allLangsTableName)
	rows, err := dbPool.Query(context.TODO(), query)
	if err != nil {
		return []Language{}, err
	}
	// this is clumsy
	type stringList struct {
		Lang_code string `db:"lang_code"`
	}
	langCodes, err := pgx.CollectRows(rows, pgx.RowToStructByName[stringList])
	if err != nil {
		fmt.Println(err)
		return []Language{}, err
	}
	supportedLangs := make([]Language, len(langCodes))
	for i, code := range langCodes {
		supportedLangs[i] = Language{languageCodeToName[code.Lang_code], code.Lang_code}
	}
	return supportedLangs, nil
}

func createOrUpdatePrivateChat(dbPool *pgxpool.Pool, username string, chatId int64) error {
	query := fmt.Sprintf("INSERT INTO %s VALUES ('%s', %d) ON CONFLICT (username) DO UPDATE SET chat_id=%d", privateChatsTableName, username, chatId, chatId)
	_, err := dbPool.Query(context.TODO(), query)
	if err != nil {
		return err
	}
	return nil
}

func shuffleSlice[T any](sl []T, n int) {
	// shuffles first n elements in a slice in-place
	// if n < 0 shuffles all slice
	slLen := len(sl)
	if n < 0 || n > slLen {
		n = slLen
	}
	for i := 0; i < n-1; i++ {
		// for each i it picks a random element in sl[i+1:]
		// and swaps them
		randomIndex := i + 1 + rand.Intn(slLen-i-1)
		sl[i], sl[randomIndex] = sl[randomIndex], sl[i]
	}
}

func excapeChars(s string) string {
	// Telegram won't send messages containing some special
	// characters and demands them to be escaped
	specialChars := "[]()~`>#+-={}.!"
	for _, char := range specialChars {
		s = strings.Replace(s, string(char), fmt.Sprintf("\\%c", char), -1)
	}
	return s
}
