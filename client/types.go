package telegram

type ResponseUpdates struct {
	Ok     bool     `json:"ok"`
	Result []Update `json:"result"`
}

type Update struct {
	UpdateId      int            `json:"update_id"`
	Message       *Message       `json:"message"`
	CallbackQuery *CallbackQuery `json:"callback_query"`
}

type Message struct {
	MessageId   int         `json:"message_id"`
	Chat        Chat        `json:"chat"`
	From        From        `json:"from"`
	Text        string      `json:"text"`
	ReplyMarkup ReplyMarkup `json:"reply_markup"`
}

type CallbackQuery struct {
	Id      string  `json:"id"`
	From    From    `json:"from"`
	Message Message `json:"message"`
	Data    string  `json:"data"`
}

type ReplyMarkup struct {
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard"`
}

type InlineKeyboardButton struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data"`
}

type Chat struct {
	Id int `json:"id"`
}

type From struct {
	Username string `json:"username"`
}

type keyButton struct {
}

type ReplyKeyboardMarkup struct {
	keaboard keyButton `json:"keyboard"`
}
