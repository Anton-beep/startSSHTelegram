package telegramLogic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"slices"
	"startSSHTelegram/telegramBot"
	"strings"
	"time"

	"go.uber.org/zap"
)

type Trigger func(update telegramBot.Update) bool

type Action func(update telegramBot.Update)

var (
	coammands = []telegramBot.BotCommand{
		{Command: "ping", Description: "ping pong, bot availability test"},
		{Command: "opentunnel", Description: "open tunnel for ssh connection using serveo"},
		{Command: "stopconnecting", Description: "stop connecting to tunnel"},
		{Command: "openngrok", Description: "open tunnel using ngrok"},
	}
)

type TelegramLogic struct {
	tgBot         *telegramBot.TelegramBot
	goodChats     []int
	pathToNgrok   string
	triggers      []Trigger
	actions       []Action
	tunnelContext context.Context
	cancelFunc    context.CancelFunc
}

func New(tgBot *telegramBot.TelegramBot, goodChats []int, ngrokPath string) (*TelegramLogic, error) {
	for _, el := range goodChats {
		scope := &telegramBot.BotCommandScopeChat{
			Type:   "chat",
			ChatId: el,
		}
		res, err := tgBot.SetMyCommands(coammands, scope)
		if err != nil {
			zap.S().Warn("error sending commands: ", err)
		}
		if !res {
			zap.S().Warn("recieved false from telegram, while sending commands")
		}
	}

	t := &TelegramLogic{tgBot, nil, "", nil, nil, nil, nil}

	// each trigger correspons to action by index, so if isPing has index = 0 in slice triggers, then action with index = 0 from slice actions will be executed
	triggers := []Trigger{
		t.isUnwantedUser,

		t.isPing,
		t.isOpenTunnel,
		t.isStopConnecting,
		t.isOpenNgrok,

		t.isNothingToSend,
	}

	actions := []Action{
		t.actIgnore,

		t.actPong,
		t.actOpenTunnel,
		t.actStopConnecting,
		t.actOpenNgrok,

		t.actNothingToSend,
	}

	if len(triggers) != len(actions) {
		return nil, errors.New("triggers and trigger_action must have the same len")
	}

	t.pathToNgrok = ngrokPath
	t.triggers = triggers
	t.actions = actions
	t.goodChats = goodChats
	t.tunnelContext, t.cancelFunc = context.WithCancel(context.Background())
	t.cancelFunc()

	return t, nil
}

func (t *TelegramLogic) isPing(update telegramBot.Update) bool {
	return update.Message.Text == "/ping"
}

func (t *TelegramLogic) actPong(update telegramBot.Update) {
	_, err := t.tgBot.SendTextMessage("pong", update.Message.Chat.Id, "")
	if err != nil {
		zap.S().Warn(err)
	}
}

func (t *TelegramLogic) isUnwantedUser(update telegramBot.Update) bool {
	return !slices.Contains(t.goodChats, update.Message.Chat.Id)
}

func (t *TelegramLogic) actIgnore(update telegramBot.Update) {
	zap.S().Warn("ignored: ", update)
}

func (t *TelegramLogic) isNothingToSend(update telegramBot.Update) bool {
	return true
}

func (t *TelegramLogic) actNothingToSend(update telegramBot.Update) {
	_, err := t.tgBot.SendTextMessage("unknown command", update.Message.Chat.Id, "")
	if err != nil {
		zap.S().Warn(err)
	}
}

func (t *TelegramLogic) isOpenTunnel(update telegramBot.Update) bool {
	return update.Message.Text == "/opentunnel"
}

func (t *TelegramLogic) actOpenTunnel(update telegramBot.Update) {
	_, err := t.tgBot.SendTextMessage("trying to open tunnel...", update.Message.Chat.Id, "")
	if err != nil {
		zap.S().Warn(err)
	}
	t.cancelFunc()

	t.tunnelContext, t.cancelFunc = context.WithCancel(context.Background())

	go func() {
		cmd := exec.CommandContext(t.tunnelContext, "ssh", "-R", "antdesktop:22:localhost:22", "serveo.net")
		stdout, err := cmd.Output()

		if err != nil {
			_, err := t.tgBot.SendTextMessage("err "+err.Error(), update.Message.Chat.Id, "")
			if err != nil {
				zap.S().Warn(err)
			}
		}

		_, err = t.tgBot.SendTextMessage(string(stdout), update.Message.Chat.Id, "")
		if err != nil {
			zap.S().Warn(err)
		}
	}()

	_, err = t.tgBot.SendTextMessage("use\n\n`ssh -J serveo.net usr@antdesktop`\n\nto connect", update.Message.Chat.Id, "MarkdownV2")
	if err != nil {
		zap.S().Warn(err)
	}
}

func (t *TelegramLogic) isStopConnecting(update telegramBot.Update) bool {
	return update.Message.Text == "/stopconnecting"
}

func (t *TelegramLogic) actStopConnecting(update telegramBot.Update) {
	if errors.Is(t.tunnelContext.Err(), context.Canceled) {
		_, err := t.tgBot.SendTextMessage("already stopped", update.Message.Chat.Id, "")
		if err != nil {
			zap.S().Warn(err)
		}
	}
	_, err := t.tgBot.SendTextMessage("stopping...", update.Message.Chat.Id, "")
	if err != nil {
		zap.S().Warn(err)
	}

	t.cancelFunc()
}

func (t *TelegramLogic) isOpenNgrok(update telegramBot.Update) bool {
	return update.Message.Text == "/openngrok"
}

type tunnel struct {
	PublicUrl string `json:"public_url"`
}

type tunnelsData struct {
	Tunnels []tunnel `json:"tunnels"`
}

func (t *TelegramLogic) actOpenNgrok(update telegramBot.Update) {
	_, err := t.tgBot.SendTextMessage("openning ngrok...", update.Message.Chat.Id, "")
	if err != nil {
		zap.S().Warn(err)
	}

	t.tunnelContext, t.cancelFunc = context.WithCancel(context.Background())

	go func() {
		cmd := exec.CommandContext(t.tunnelContext, t.pathToNgrok, "tcp", "22")
		err := cmd.Start()
		if err != nil {
			_, err = t.tgBot.SendTextMessage(err.Error(), update.Message.Chat.Id, "")
			if err != nil {
				zap.S().Warn(err)
			}
		}
	}()

	time.Sleep(3 * time.Second)

	resp, err := http.Get("http://localhost:4040/api/tunnels/")
	if err != nil {
		zap.S().Warn(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		_, err := t.tgBot.SendTextMessage(err.Error(), update.Message.Chat.Id, "")
		if err != nil {
			zap.S().Warn(err)
		}
	}

	var out tunnelsData
	if err := json.Unmarshal(body, &out); err != nil {
		_, err := t.tgBot.SendTextMessage(err.Error(), update.Message.Chat.Id, "")
		if err != nil {
			zap.S().Warn(err)
		}
	}

	ipAndPort := strings.Split(out.Tunnels[0].PublicUrl, ":")
	ip, port := ipAndPort[1][2:], ipAndPort[2]
	_, err = t.tgBot.SendTextMessage(fmt.Sprintf("use\n\n`ssh usr@%v -p %v`\n\nto connect", ip, port), update.Message.Chat.Id, "MarkdownV2")
	if err != nil {
		zap.S().Warn(err)
	}
}

func (t *TelegramLogic) Run() {
	offset := 0
	for {
		updates, err := t.tgBot.GetUpdates(offset, 3)
		if err != nil {
			zap.S().Warn("error occured while getting updates: ", err)
			time.Sleep(2 * time.Second)
		}

		for _, update := range updates {
			offset = update.UpdateId + 1

			for ind, trigger := range t.triggers {
				if trigger(update) {
					go t.actions[ind](update)
					break
				}
			}
		}
	}
}
