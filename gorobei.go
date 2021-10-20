package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/mymmrac/telego"
	"github.com/phuslu/log"
	"gorobei/clock"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"regexp"
	"time"
)

type Gorobei struct {
	d       *Db
	chat    string
	chatId  int64
	admin   string
	adminId int64
	tg *Telegram
	clock clock.Clock
}

var ErrImageAlreadyProcessed = errors.New("image has been processed already")

func (g *Gorobei) Init() error {
	params := &telego.GetChatParams{ChatID: telego.ChatID{Username: g.chat}}
	chat, err := g.tg.Bot.GetChat(params)
	if err != nil {
		return err
	}
	g.chatId = chat.ID
	err = g.tg.DoUpdates()
	if err != nil {
		return err
	}

	if g.admin != "" {
		g.adminId, err = g.d.ReadUserId(g.admin)
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
			log.Error().Str("src", m[0]).Msg("unexpected utils match")
		}
	}
	if errc > 0 || (total - skipped) > 0 {
		msg := fmt.Sprintf("Fetching images from the [page](%s) completed\\.\nTotal: %v\nNew: %v\nSkipped: %v\nErrors: %v", url, total, total-skipped, skipped, errc)
		err = g.SendAdminMessage(msg)
		return err
	}
	return nil
}

func (g *Gorobei) updateDailyReport(total, skipped, errc int) error {
	r, err := g.d.ReadDailyReport()
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			// no report yet
			r = &DailyReport{SentAt: g.clock.Now()}
		} else {
			return err
		}
	}
	r.Posted += total - skipped
	r.Errors += errc
	d, m, h := r.SentAt.Date()
	// daily reports are sent after 23:00
	planned := time.Date(d,m,h, 23,0,0,0, time.Local)
	if g.clock.Now().Sub(planned) > 0 {
		// send report

	}
	return nil
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
	done, err := g.d.StoreUrlProcessed(src)
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
	err = g.SendChatImage(path, "")
	if err != nil {
		log.Error().Err(err).Msg("cannot send fetched image to the chat")
		return err
	}
	err = g.d.ReadUrlProcessed(src, 1)
	return err

}

func (g *Gorobei) SendChatImage(image string, caption string) error {
	return g.tg.SendImage("", g.chatId, image, caption)
}

func (g *Gorobei) SendChatMessage(message string) error {
	return g.tg.SendMessage("", g.chatId, message, "")
}

func (g *Gorobei) SendAdminMessage(message string) error {
	if g.adminId == 0 {
		return nil
	}

	return g.tg.SendMessage("", g.adminId, message, "MarkdownV2")
}

func (g *Gorobei) ForgetImg(url string) error {
	return g.d.ReadUrlProcessed(url, 0)
}
