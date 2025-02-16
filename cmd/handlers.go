package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5"
	"net/http"
)

// TODO: take it from a db
var supportedLanguages []Language = []Language{
	{"Polish", "pl"},
	{"Russian", "ru"},
	{"Italian", "it"}}

var difficultyToString map[uint8]string = map[uint8]string{
	0: "Most frequent words",
	1: "Pretty common words",
	2: "Less common words"}

type Language struct {
	fullName string
	code     string
}

type CallbackReply func(CallbackQuery) Responder
type MsgReply func(Message) Responder

type Responders []Responder

// this makes a list of Responders also a Responder
func (r Responders) Respond(baseUrl string) {
	for _, responder := range r {
		responder.Respond(baseUrl)
	}
}

type CommandHandler struct {
	commands map[string]MsgReply
	db       *pgx.Conn
}

func NewCommandHandler() CommandHandler {
	return CommandHandler{commands: make(map[string]MsgReply)}

}

func (cmdHandler CommandHandler) AddCommand(name string, msgHandler MsgReply) {
	cmdHandler.commands[name] = msgHandler
}

func (cmdHandler CommandHandler) GetResponder(msg Message) Responder {
	msgHandler, ok := cmdHandler.commands[msg.Text]
	fmt.Println(ok)
	if !ok {
		return SendMsg{}
	}
	return msgHandler(msg)
}

type CallbackHandler struct {
	db *pgx.Conn
}

func (callbackHandler CallbackHandler) GetResponder(cq CallbackQuery) Responder {
	if cq.Msg.Id != 0 {
		var callbackData MenuCallbackData
		err := json.Unmarshal([]byte(cq.Data), &callbackData)
		if err != nil {
			fmt.Println(err)
			return SendMsg{}
		}
		switch callbackData.Stage {
		case 0:
			return EditMsg{
				ChatId:      cq.Msg.Chat.Id,
				Text:        "Now choose the difficulty (how common or rare are the words)",
				ReplyMarkup: chooseLevelKeyboardMarkup(callbackData.Language),
				MsgId:       cq.Msg.Id}
		case 1:
			dm := DeleteMsg{ChatId: cq.Msg.Chat.Id, MsgId: cq.Msg.Id}
			sm := nextWord(cq, callbackData.Language, callbackData.Difficulty)
			return Responders{dm, sm}
		}
	}
	return SendMsg{}
}

type SendMsg struct {
	ChatId      int64                 `json:"chat_id"`
	Text        string                `json:"text"`
	ReplyMarkup *InlineKeyboardMarkup `json:"reply_markup,omitempty"` // has to be a pointer for json serializer to skip it if empty
}

// TODO: do it with more generic type instead of Responder
func postRequest(baseUrl string, method string, r Responder) {
	data, err := json.Marshal(r)
	if err != nil {
		fmt.Println("failed to marshal: ", r)
		fmt.Println(err)
	}
	url := fmt.Sprintf("%s/%s", baseUrl, method)
	client := &http.Client{}
	res, err := client.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		// TODO: do smth if Telegram gives an error back
		fmt.Printf("error on post request: %s\n", err)
	}
	defer res.Body.Close()
	// TODO: define a new type for getting Telegram errors
	var j interface{}
	err = json.NewDecoder(res.Body).Decode(&j)
	if err != nil {
		fmt.Println("serializing error: ", err)
	}
}

func (sm SendMsg) Respond(baseUrl string) {
	postRequest(baseUrl, "sendMessage", sm)
}

// TODO: there should be a way to embed SendMsg into this, its just one field more
type EditMsg struct {
	ChatId      int64                 `json:"chat_id"`
	Text        string                `json:"text"`
	ReplyMarkup *InlineKeyboardMarkup `json:"reply_markup,omitempty"` // has to be a pointer for json serializer to skip it if empty
	MsgId       int64                 `json:"message_id"`             // id of a message to edit
}

// TODO: refactor, coz this is a duplicated code except for two variables
func (em EditMsg) Respond(baseUrl string) {
	postRequest(baseUrl, "editMessageText", em)
}

type DeleteMsg struct {
	ChatId int64 `json:"chat_id"`
	MsgId  int64 `json:"message_id"`
}

func (dm DeleteMsg) Respond(baseUrl string) {
	postRequest(baseUrl, "deleteMessage", dm)
}

type MenuCallbackData struct {
	Stage      uint8  `json:"stage"`
	Language   string `json:"language"`
	Difficulty uint8  `json:"difficulty"`
}

func start(msg Message) Responder {
	text := "Welcome"
	// TODO: saving a user to the database
	return SendMsg{ChatId: msg.Chat.Id, Text: text}
}

func menu(msg Message) Responder {
	text := "Choose a language:"
	return SendMsg{ChatId: msg.Chat.Id, Text: text, ReplyMarkup: chooseLanguageKeyboardMarkup()}
}

func about(msg Message) Responder {
	text := "about this bot"
	return SendMsg{ChatId: msg.Chat.Id, Text: text}
}

func help(msg Message) Responder {
	text := "here are all the commands"
	return SendMsg{ChatId: msg.Chat.Id, Text: text}
}

// func menuCallback(callback CallbackQuery) Responder {
// 	if callback.Msg.Id != 0 {
// 		var callbackData MenuCallbackData
// 		err := json.Unmarshal([]byte(callback.Data), &callbackData)
// 		if err != nil {
// 			fmt.Println(err)
// 			return SendMsg{}
// 		}
// 		switch callbackData.Stage {
// 		case 0:
// 			return EditMsg{
// 				ChatId:      callback.Msg.Chat.Id,
// 				Text:        "Now choose the difficulty (how common or rare are the words)",
// 				ReplyMarkup: chooseLevelKeyboardMarkup(callbackData.Language),
// 				MsgId:       callback.Msg.Id}
// 		case 1:
// 			dm := DeleteMsg{ChatId: callback.Msg.Chat.Id, MsgId: callback.Msg.Id}
// 			sm := nextWord(callback, callbackData.Language, callbackData.Difficulty)
// 			return Responders{dm, sm}
// 		}
// 	}
// 	return SendMsg{}
// }

func nextWord(callbackQuery CallbackQuery, langCode string, difficulty uint8) SendMsg {
	// query db for a random word in a given difficulty and language
	return SendMsg{ChatId: callbackQuery.Msg.Chat.Id, Text: "Fetching data from my database"}
}

func (cq CallbackQuery) answer(baseUrl string) {
	// this function needs to be called to stop buttons from
	// blinking after they're pressed by the user
	url := fmt.Sprintf("%s/answerCallbackQuery", baseUrl)
	data := fmt.Sprintf("{\"callback_query_id\":\"%s\"}", cq.Id)
	client := &http.Client{}
	res, _ := client.Post(url, "application/json", bytes.NewBufferString(data)) //I dont even care about this error, no big deal
	res.Body.Close()
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
	buttonsNum := len(difficultyToString)
	keyboard := make([][]InlineKeyboardButton, buttonsNum)
	for i := 0; i < buttonsNum; i++ {
		data := MenuCallbackData{Stage: 1, Language: langCode, Difficulty: uint8(i)}
		callbackData, err := json.Marshal(data)
		if err != nil {
			fmt.Println(err)
		}
		keyboard[i] = []InlineKeyboardButton{{Text: difficultyToString[uint8(i)], CallbackData: string(callbackData)}}
	}
	return &InlineKeyboardMarkup{InlineKeyboard: keyboard}
}
