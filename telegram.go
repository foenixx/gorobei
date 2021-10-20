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

type Telegram struct {
	Bot   *telego.Bot
	store UsersStore
	limiters map[string]*limiter.RateLimiter // key = channel_name or string(chat_id)
}

func NewTelegram(token string, store UsersStore) (*Telegram, error) {
	bot, err := telego.NewBot(token)
	if err != nil {
		return nil, err
	}
	var tg = Telegram{
		Bot:      bot,
		store:    store,
		limiters: make(map[string]*limiter.RateLimiter),
	}
	return &tg, nil
}

func (tg *Telegram) SendImage(user string, userId int64, image string, caption string) error {
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

func (tg *Telegram) getLimiter(id *telego.ChatID) *limiter.RateLimiter {
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

func (tg *Telegram) SendMessage(user string, userId int64, message string, mode string) error {
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

func (tg *Telegram) constructChatId(user string, userId int64) (*telego.ChatID, error) {
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
func (tg *Telegram) DoUpdates() error {
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

var (
	rePre = regexp.MustCompile(`(?ms)^\x60{3}\S*\n(.*?)\n\x60{3}`)
	reInlineCode= regexp.MustCompile(`\x60(.*?)\x60`)
	reLink= regexp.MustCompile(`\[.+?]\((.+?)\)`)
	reOthers = regexp.MustCompile(`[_*[\]()~\x60>#+-=|{}.!]`)
)

func replOthers(i []byte) []byte {
	return []byte(reOthers.ReplaceAllString(string(i), `\$0`))
}

// EscapeMarkdown implements required message text escaping: https://core.telegram.org/bots/api#formatting-options.
// Inside pre and code entities, all '`' and '\' characters must be escaped with a preceding '\' character.
// Inside (...) part of inline link definition, all ')' and '\' must be escaped with a preceding '\' character.
// In all other places characters '_', '*', '[', ']', '(', ')', '~', '`', '>', '#', '+', '-', '=', '|', '{', '}', '.', '!' must be escaped with the preceding character '\'.
func EscapeMarkdown(msg string) string {
	replPre := func(i [][]byte) [][]byte{
		s := strings.ReplaceAll(string(i[0]), `\`, `\\`)
		s = strings.ReplaceAll(s, "`", "\\`")
		i[0] = []byte(s)
		return i
	}
	replNotPre := func(s []byte) []byte{
		res := utils.ReplaceAllSubmatchFunc(reInlineCode, s, func(i [][]byte) [][]byte {
			i[0] = []byte(strings.ReplaceAll(string(i[0]), `\`, `\\`))
			return i
		}, -1)
		res = utils.ReplaceAllSubmatchFunc(reLink, res, func(i [][]byte) [][]byte {
			i[0] = []byte(strings.ReplaceAll(string(i[0]), `\`, `\\`))
			return i
		}, -1)
		return res
	}

	res := utils.ReplaceAllSubmatchFunc2(rePre, []byte(msg), replPre, replNotPre, -1)
	return string(res)
}