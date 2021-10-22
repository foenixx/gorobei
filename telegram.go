package main

import (
	"errors"
	"fmt"
	"github.com/mymmrac/telego"
	"github.com/phuslu/log"
	"gorobei/limiter"
	"gorobei/utils"
	"os"
	"regexp"
	"strings"
)

type (
	Telegram interface {
		SendImage(user string, userId int64, image string, caption string) error
		SendMessageMarkdown(user string, userId int64, message string) error
		SendMessageText(user string, userId int64, message string) error
		DoUpdates() error
		ChatInfo(string) (*telego.Chat, error)
	}
	telegramImpl struct {
		Bot      *telego.Bot
		store    UsersStore
		limiters map[string]*limiter.RateLimiter // key = channel_name or string(chat_id)
	}
)

var (
	_ Telegram = (*telegramImpl)(nil)

	testMessageText = utils.Bt(`
Message header!
_Line_ *with* __inline__ ~markup~.
{Line}+(with)-[special]=#symbols.

[image](https://i.imgur.com/sMhpFyR.jpg)

__error__:
³³³
not enough arguments, expected at least 3, got 0
main.parseArgs
        /home/dfc/src/github.com/pkg/errors/_examples/wrap/main.go:12
main.main
        /home/dfc/src/github.com/pkg/errors/_examples/wrap/main.go:18
runtime.main
        /home/dfc/go/src/runtime/proc.go:183
runtime.goexit
        /home/dfc/go/src/runtime/asm_amd64.s:2059
³³³`)
)

func NewTelegram(token string, store UsersStore) (Telegram, error) {
	bot, err := telego.NewBot(token)
	if err != nil {
		return nil, err
	}
	var tg = telegramImpl{
		Bot:      bot,
		store:    store,
		limiters: make(map[string]*limiter.RateLimiter),
	}
	return &tg, nil
}

func (tg *telegramImpl) SendImage(user string, userId int64, image string, caption string) error {
	chatId, err := tg.constructChatId(user, userId)
	if err != nil {
		return err
	}

	tg.getLimiter(chatId).TikTak()

	f, err := os.Open(image)
	if err != nil {
		return err
	}

	params := &telego.SendPhotoParams{
		ChatID:  *chatId,
		Photo:   telego.InputFile{File: f},
		Caption: caption,
	}
	_, err = tg.Bot.SendPhoto(params)
	return err
}

func (tg *telegramImpl) getLimiter(id *telego.ChatID) *limiter.RateLimiter {
	var key string
	if id.ID != 0 {
		key = fmt.Sprintf("%v", id.ID)
	} else {
		key = id.Username
	}
	if key == "" {
		log.Error().Msg("cannot get limiter: empty chat id")
		return nil
	}
	l, ok := tg.limiters[key]
	if !ok {
		l = limiter.NewLimiter()
		tg.limiters[key] = l
	}
	return l
}

func (tg *telegramImpl) SendMessageMarkdown(user string, userId int64, message string) error {
	message = EscapeMarkdown(message)
	return tg.sendMessage(user, userId, message, "MarkdownV2")
}

func (tg *telegramImpl) SendMessageText(user string, userId int64, message string) error {
	return tg.sendMessage(user, userId, message, "")
}

func (tg *telegramImpl) sendMessage(user string, userId int64, message string, mode string) error {
	chatId, err := tg.constructChatId(user, userId)
	if err != nil {
		return err
	}

	tg.getLimiter(chatId).TikTak()

	params := &telego.SendMessageParams{
		ChatID:    *chatId,
		Text:      message,
		ParseMode: mode,
	}
	_, err = tg.Bot.SendMessage(params)
	return err
}

func (tg *telegramImpl) constructChatId(user string, userId int64) (*telego.ChatID, error) {
	var chatId telego.ChatID
	user = strings.ToLower(user)

	switch {
	case user == "":
		chatId.ID = userId
	case user[0] == '@':
		//channel or group name
		chatId.Username = user
	default:
		//ordinary user
		id, err := tg.store.ReadUserId(user)
		if errors.Is(err, ErrNotFound) {
			return nil, fmt.Errorf("cannot find ID of the user '%s'. To enable Bot to send messages to a user, the user should send '/start' message to the Bot first", user)
		}
		if err != nil {
			return nil, err
		}
		chatId.ID = id
	}
	return &chatId, nil
}

// https://api.telegram.org/bot<TOKEN>/getUpdates
//
func (tg *telegramImpl) DoUpdates() error {
	var err error
	var updates []telego.Update
	params := &telego.GetUpdatesParams{
		Timeout: 0,
	}
	updates, err = tg.Bot.GetUpdates(params)
	if err != nil {
		return err
	}

	for _, u := range updates {
		if u.Message != nil {
			uname := strings.ToLower(u.Message.Chat.Username)
			_, err = tg.store.ReadUserId(uname)
			if errors.Is(err, ErrNotFound) {
				err = tg.store.StoreUserId(uname, u.Message.Chat.ID)
				if err != nil {
					return err
				}
				continue
			}
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (tg *telegramImpl) ChatInfo(name string) (*telego.Chat, error) {
	params := &telego.GetChatParams{ChatID: telego.ChatID{Username: name}}
	return tg.Bot.GetChat(params)
}

var (
	rePre        = regexp.MustCompile(`(?ms)^\x60{3}\S*\n(.*?)\n\x60{3}`)
	reInlineCode = regexp.MustCompile(`\x60(.*?)\x60`)
	reLink       = regexp.MustCompile(`\[.+?]\((.+?)\)`)
	reOthers     = regexp.MustCompile(`[[\]()\x60>#+\-=|{}.!\\]`)
)

func replacePreInside(g [][]byte) [][]byte {
	g[0] = replaceSlashBacktick(g[0])
	return g
}
func replaceSlashBacktick(s []byte) []byte {
	res := strings.ReplaceAll(string(s), `\`, `\\`)
	return []byte(strings.ReplaceAll(res, "`", "\\`"))
}

func replacePreOutside(s []byte) []byte {
	return utils.ReplaceAllSubmatchFunc2(reInlineCode, s, replaceInlineCodeInside, replaceInlineCodeOutside, -1)
}

func replaceInlineCodeInside(g [][]byte) [][]byte {
	g[0] = replaceSlashBacktick(g[0])
	return g
}

func replaceInlineCodeOutside(s []byte) []byte {
	return utils.ReplaceAllSubmatchFunc2(reLink, s, replaceLinkInside, replaceAllChars, -1)
}

func replaceLinkInside(g [][]byte) [][]byte {
	g[0] = replaceSlashBacktick(g[0])
	return g
}

func replaceAllChars(i []byte) []byte {
	return []byte(reOthers.ReplaceAllString(string(i), `\$0`))
}

// EscapeMarkdown implements required message text escaping: https://core.telegram.org/bots/api#formatting-options.
// Inside pre and code entities, all '`' and '\' characters must be escaped with a preceding '\' character.
// Inside (...) part of inline link definition, all ')' and '\' must be escaped with a preceding '\' character.
// In all other places characters '_', '*', '[', ']', '(', ')', '~', '`', '>', '#', '+', '-', '=', '|', '{', '}', '.', '!' must be escaped with the preceding character '\'.
func EscapeMarkdown(msg string) string {
	res := utils.ReplaceAllSubmatchFunc2(rePre, []byte(msg), replacePreInside, replacePreOutside, -1)
	return string(res)
}
