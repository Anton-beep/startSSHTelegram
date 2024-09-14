package telegramBot

type Message struct {
	MessageId int    `json:"message_id"`
	From      User   `json:"from"`
	Date      int    `json:"date"`
	Chat      Chat   `json:"chat"`
	Text      string `json:"text"`
}
