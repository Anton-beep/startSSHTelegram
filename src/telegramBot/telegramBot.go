package telegramBot

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

var BASE_URL string = "https://api.telegram.org/bot"

type TelegramBot struct {
}

func New(token string) *TelegramBot {
	BASE_URL += token
	return &TelegramBot{}
}

type getUpdatesIn struct {
	Offset  int `json:"offset"`
	Timeout int `json:"timeout"`
}

type getUpdatesOut struct {
	Response
	Result []Update `json:"result"`
}

func (t *TelegramBot) GetUpdates(offset int, timeout int) ([]Update, error) {
	inValues := getUpdatesIn{
		offset,
		timeout,
	}

	data, err := json.Marshal(inValues)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(BASE_URL+"/getUpdates", "application/json", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var out getUpdatesOut

	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}

	return out.Result, nil
}

type setMyCommandsIn struct {
	Commands []BotCommand    `json:"commands"`
	Scope    BotCommandScope `json:"scope"`
}

type setMyCommandsOut struct {
	Response
	Result bool `json:"result"`
}

func (t *TelegramBot) SetMyCommands(commands []BotCommand, scope BotCommandScope) (bool, error) {
	inValues := setMyCommandsIn{
		commands,
		scope,
	}

	data, err := json.Marshal(inValues)
	if err != nil {
		return false, err
	}

	resp, err := http.Post(BASE_URL+"/setMyCommands", "application/json", bytes.NewReader(data))
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	var out setMyCommandsOut

	if err := json.Unmarshal(body, &out); err != nil {
		return false, err
	}

	return out.Result, nil
}

func (t *TelegramBot) SendTextMessage(message string, chatId int, parseMode string) (*Message, error) {
	inValues := struct {
		ChatId    int    `json:"chat_id"`
		Text      string `json:"text"`
		ParseMode string `json:"parse_mode"`
	}{
		chatId,
		message,
		parseMode,
	}

	data, err := json.Marshal(inValues)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(BASE_URL+"/sendMessage", "application/json", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var res Message

	if err := json.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}
