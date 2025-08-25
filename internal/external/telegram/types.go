package telegram

type InlineKeyboardButton struct {
	Text string `json:"text"`
	URL  string `json:"url"`
}

type InlineKeyboardMarkup struct {
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard"`
}

type ReplyParameters struct {
	MessageID                int    `json:"message_id"`
	ChatID                   any    `json:"chat_id,omitempty"`
	AllowSendingWithoutReply bool   `json:"allow_sending_without_reply,omitempty"`
	Quote                    string `json:"quote,omitempty"`
	QuotePosition            int    `json:"quote_position,omitempty"`
	ChecklistTaskID          int    `json:"checklist_task_id,omitempty"`
}

type TelegramMessagePayload struct {
	ChatID          string               `json:"chat_id"`
	Text            string               `json:"text"`
	ReplyMarkup     InlineKeyboardMarkup `json:"reply_markup"`
	ReplyParameters ReplyParameters      `json:"reply_parameters"`
}
