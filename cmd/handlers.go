package main

import (
	"encoding/json"
	"fmt"
)

type Language struct {
	fullName string
	code     string
}

type MenuCallbackData struct {
	Stage      uint8  `json:"stage"`
	Language   string `json:"language"`
	Difficulty uint8  `json:"difficulty"`
}

var supportedLanguages []Language = []Language{
	{"Polish", "pl"},
	{"Russian", "ru"},
	{"Italian", "it"}}

func start(msg Message) SendMsgOpts {
	text := "Welcome"
	// TODO: saving a user to the database
	return SendMsgOpts{ChatId: msg.Chat.Id, Text: text}
}

func menu(msg Message) SendMsgOpts {
	text := "Choose a language:"
	return SendMsgOpts{ChatId: msg.Chat.Id, Text: text, ReplyMarkup: chooseLanguageKeyboardMarkup()}
}

func about(msg Message) SendMsgOpts {
	text := "about this bot"
	return SendMsgOpts{ChatId: msg.Chat.Id, Text: text}
}

func help(msg Message) SendMsgOpts {
	text := "here are all the commands"
	return SendMsgOpts{ChatId: msg.Chat.Id, Text: text}
}

func menuCallback(callback CallbackQuery) SendMsgOpts {
	if callback.Msg.Id != 0 {
		var callbackData MenuCallbackData
		err := json.Unmarshal([]byte(callback.Data), &callbackData)
		if err != nil {
			fmt.Println(err)
			return SendMsgOpts{}
		}
		switch callbackData.Stage {
		case 0:
			return SendMsgOpts{
				ChatId:      callback.Msg.Chat.Id,
				Text:        "Now choose the difficulty (how common or rare are the words)",
				ReplyMarkup: chooseLevelKeyboardMarkup(callbackData.Language)}
		case 1:
			return SendMsgOpts{
				ChatId: callback.Msg.Chat.Id,
				Text:   "Fetching data from my database"}
		}
	}
	return SendMsgOpts{}
}

func chooseLanguageKeyboardMarkup() *InlineKeyboardMarkup {
	buttonsNum := len(supportedLanguages)
	keyboard := make([][]InlineKeyboardButton, buttonsNum)
	for i, lang := range supportedLanguages {
		data := MenuCallbackData{Stage: 0, Language: lang.code}
		callbackData, err := json.Marshal(data)
		if err != nil {
			fmt.Println(err)
		}
		keyboard[i] = []InlineKeyboardButton{{Text: lang.fullName, CallbackData: string(callbackData)}}
	}
	return &InlineKeyboardMarkup{InlineKeyboard: keyboard}
}

func chooseLevelKeyboardMarkup(langCode string) *InlineKeyboardMarkup {
	dataEasy := MenuCallbackData{Stage: 1, Language: langCode, Difficulty: 0}
	callbackDataEasy, err := json.Marshal(dataEasy)
	if err != nil {
		fmt.Println(err)
	}
	dataMedium := MenuCallbackData{Stage: 1, Language: langCode, Difficulty: 1}
	callbackDataMedium, err := json.Marshal(dataMedium)
	if err != nil {
		fmt.Println(err)
	}
	dataHard := MenuCallbackData{Stage: 1, Language: langCode, Difficulty: 2}
	callbackDataHard, err := json.Marshal(dataHard)
	if err != nil {
		fmt.Println(err)
	}
	keyboard := [][]InlineKeyboardButton{
		{{Text: "Frequent words", CallbackData: string(callbackDataEasy)}},
		{{Text: "Frequent words", CallbackData: string(callbackDataMedium)}},
		{{Text: "Frequent words", CallbackData: string(callbackDataHard)}}}
	return &InlineKeyboardMarkup{InlineKeyboard: keyboard}
}
