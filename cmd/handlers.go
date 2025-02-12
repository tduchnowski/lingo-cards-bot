package main

func start(msg Message) SendMsgOpts {
	text := "Welcome"
	// TODO: saving a user in a database
	return SendMsgOpts{ChatId: msg.Chat.Id, Text: text}
}

func menu(msg Message) SendMsgOpts {
	text := "heres a menu:"
	return SendMsgOpts{ChatId: msg.Chat.Id, Text: text}
}

func about(msg Message) SendMsgOpts {
	text := "about this bot"
	return SendMsgOpts{ChatId: msg.Chat.Id, Text: text}
}

func help(msg Message) SendMsgOpts {
	text := "here are all the commands"
	return SendMsgOpts{ChatId: msg.Chat.Id, Text: text}
}
