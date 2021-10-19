package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/mymmrac/telego"
	"github.com/phuslu/log"
	"os"
	"strings"

	"io/ioutil"
	"mime"
	"net/http"
	"regexp"
	"time"
)

type Gorobei struct {
	d       *Db
	bot     *telego.Bot
	chat    string
	chatId  int64
	admin   string
	adminId int64
	limiter *RateLimiter
	adminLimiter *RateLimiter
}

var ErrImageAlreadyProcessed = errors.New("image has been processed already")

func (g *Gorobei) Init() error {
	params := &telego.GetChatParams{ChatID: telego.ChatID{Username: g.chat}}
	chat, err := g.bot.GetChat(params)
	if err != nil {
		return err
	}
	g.chatId = chat.ID
	err = g.processUpdates()
	if err != nil {
		return err
	}

	if g.admin != "" {
		g.adminId, err = g.d.GetUserID(g.admin)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				log.Error().Err(err).Msg("cannot obtain admin user ID. Admin notifications will be disabled.")
			} else {
				return err
			}
		}
	}
	return nil
}

func (g *Gorobei) Close() error {
	if g.d != nil {
		return g.d.Close()
	}
	return nil
}

func (g *Gorobei) Fetch(url string, limit int) error {
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return err
	}
	//log.Info().Msg(string(body))

	re := regexp.MustCompile(`(?si)<div class="singlePost.*?<div class="postInner">\s*?<div class="paragraph">[^<]*<div[^<]*<img src=["']{1}(.*?)["']{1}`)
	ma := re.FindAllStringSubmatch(string(body), -1)
	var total,skipped, errc int
	for _, m := range ma {
		if len(m) == 2 && m[1] != `https://i.imgur.com/sMhpFyR.jpg` {
			src := m[1]
			log.Info().Str("src", src).Msg("image found")
			err = g.processImage(src)
			if err != nil {
				if errors.Is(err, ErrImageAlreadyProcessed) {
					skipped += 1
				} else {
					errc += 1
					log.Error().Err(err).Str("src", src).Msg("cannot process image")
					msg := fmt.Sprintf("Error occured during processing the image\\!\n[image](%s)\n\n__error__:\n```\n%s\n```", src, err.Error())
					err2 := g.SendAdminMessage(msg)
					if err2 != nil {
						log.Error().Err(err2).Msg("cannot send admin message")
					}
				}
			}

			total += 1
			if limit > 0 && total >= limit {
				break //for
			}
		} else {
			log.Error().Str("src", m[0]).Msg("unexpected regex match")
		}
	}

	msg := fmt.Sprintf("Fetching images from the [page](%s) completed\\.\nTotal: %v\nNew: %v\nSkipped: %v\nErrors: %v", url, total, total - skipped, skipped, errc)
	err = g.SendAdminMessage(msg)
	return err
}

var ImageExt = map[string]string{
	"image/bmp":     "bmp",
	"image/gif":     "gif",
	"image/jpeg":    "jpeg",
	"image/png":     "png",
	"image/svg+xml": "svg",
	"image/tiff":    "tiff",
	"image/webp":    "webp",
}

func (g *Gorobei) downloadImage(src string) (string, error) {
	client := http.Client{
		Timeout: 15 * time.Second,
	}
	resp, err := client.Get(src)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var mediatype string
	mediatype, _, err = mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if err != nil {
		return "", err
	}
	ext, ok := ImageExt[mediatype]
	if !ok {
		return "", fmt.Errorf("unsupported mediatype: %s", mediatype)
	}
	f, err := ioutil.TempFile("", "*."+ext)
	defer f.Close()
	log.Info().Str("file", f.Name()).Msg("image temp file name")
	r := bufio.NewReader(resp.Body)
	_, err = r.WriteTo(f)
	if err != nil {
		return "", err
	}
	return f.Name(), nil
}

func (g *Gorobei) processImage(src string) error {
	done, err := g.d.Get(src)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return err
	}
	if done == 1 {
		log.Info().Str("src", src).Msg("image has been processed already")
		return ErrImageAlreadyProcessed
	}
	path, err := g.downloadImage(src)
	if err != nil {
		return err
	}
	defer os.Remove(path)
	err = g.SendChatImage(path, src)
	if err != nil {
		log.Error().Err(err).Msg("cannot send fetched image to the chat")
		return err
	}
	err = g.d.Set(src, 1)
	return err

}

func (g *Gorobei) constructChatId(user string, userId int64) (*telego.ChatID, error) {
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
		id, err := g.d.GetUserID(user)
		if errors.Is(err, ErrNotFound) {
			return nil, fmt.Errorf("cannot find ID of the user '%s'. To enable bot to send messages to a user, the user should send '/start' message to the bot first", user)
		}
		if err != nil {
			return nil, err
		}
		chatId.ID = id
	}
	return &chatId, nil
}

func (g *Gorobei) SendChatImage(image string, caption string) error {
	return g.SendImage("", g.chatId, image, caption, nil)
}

func (g *Gorobei) SendImage(user string, userId int64, image string, caption string, limiter *RateLimiter) error {
	if limiter == nil {
		limiter = g.limiter
	}
	limiter.TikTak()
	chatId, err := g.constructChatId(user, userId)
	if err != nil {
		return err
	}
	// Send using file from disk
	//f, err := os.Open("/tmp/996584576.jpeg")
	//f, err := os.Open("c:/personal/1430353161-0a6e350da88a8d827765cc8fe6dccd70.jpg")
	f, err := os.Open(image)
	if err != nil {
		return err
	}

	params := &telego.SendPhotoParams{
		ChatID:  *chatId,
		Photo:   telego.InputFile{File: f},
		Caption: caption,
	}
	_, err = g.bot.SendPhoto(params)
	return err
}

func (g *Gorobei) SendMessage(user string, userID int64, message string, mode string, limiter *RateLimiter) error {
	if limiter == nil {
		limiter = g.limiter
	}
	limiter.TikTak()
	chatId, err := g.constructChatId(user, userID)
	if err != nil {
		return err
	}

	params := &telego.SendMessageParams{
		ChatID:    *chatId,
		Text:      message,
		ParseMode: mode,
	}
	_, err = g.bot.SendMessage(params)
	return err
}

func (g *Gorobei) SendChatMessage(message string) error {
	return g.SendMessage("", g.chatId, message, "", nil)
}

func (g *Gorobei) SendAdminMessage(message string) error {
	if g.adminId == 0 {
		return nil
	}

	return g.SendMessage("", g.adminId, message, "MarkdownV2", g.adminLimiter)
}
// https://api.telegram.org/bot<TOKEN>/getUpdates
//
func (g *Gorobei) processUpdates() error {
	var err error
	var updates []telego.Update
	params := &telego.GetUpdatesParams{
		Timeout: 0,
	}
	updates, err = g.bot.GetUpdates(params)
	if err != nil {
		return err
	}

	for _, u := range updates {
		if u.Message != nil {
			uname := strings.ToLower(u.Message.Chat.Username)
			_, err = g.d.GetUserID(uname)
			if errors.Is(err, ErrNotFound) {
				err = g.d.SetUserID(uname, u.Message.Chat.ID)
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

func (g *Gorobei) ForgetImg(url string) error {
	return g.d.Set(url, 0)
}
