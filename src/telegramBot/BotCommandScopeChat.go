package telegramBot

type BotCommandScopeChat struct {
	Type   string `json:"type"`
	ChatId int    `json:"chat_id"`
}

func (b *BotCommandScopeChat) BotCommandScope() {}
