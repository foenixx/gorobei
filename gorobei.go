package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/mymmrac/telego"
	"github.com/phuslu/log"
	"os"

	"io/ioutil"
	"mime"
	"net/http"
	"regexp"
	"time"
)

type Gorobei struct {
	d *Db
	bot *telego.Bot
	chat string
	chatID int64
}

func (g *Gorobei) Init() error {
	params := &telego.GetChatParams{ChatID: telego.ChatID{ Username: g.chat }}
	chat, err := g.bot.GetChat(params)
	if err != nil {
		return err
	}
	g.chatID = chat.ID
	return nil
}

func (g *Gorobei) Close() error {
	if g.d != nil {
		return g.d.Close()
	}
	return nil
}


func (g *Gorobei) Fetch(url string) error {
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

	re := regexp.MustCompile(`(?si)<div class="singlePost.*?<div class="postInner">\s*?<div class="paragraph">\s*?<div.*?<img src="(.*?)"`)
	ma := re.FindAllStringSubmatch(string(body), -1)
	for _, m := range ma {
		if len(m) == 2 && m[1] != `https://i.imgur.com/sMhpFyR.jpg` {
			log.Info().Str("src", m[1]).Msg("image found")
			err = g.processImage(m[1])
			if err != nil {
				log.Error().Err(err).Msg("cannot process image")
			}
		} else {
			log.Error().Str("src", m[0]).Msg("unexpected regex match")
		}
	}
	return nil
}

var ImageExt = map[string]string {
	"image/bmp": "bmp",
	"image/gif": "gif",
	"image/jpeg": "jpeg",
	"image/png": "png",
	"image/svg+xml": "svg",
	"image/tiff": "tiff",
	"image/webp": "webp",
}

func (g *Gorobei) processImage(src string) error {
	done, err := g.d.Get(src)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return err
	}
	if done == 1 {
		log.Info().Str("src", src).Msg("image has been processed already")
		return nil
	}

	client := http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Get(src)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var mediatype string
	mediatype, _, err = mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if err != nil {
		return err
	}
	ext, ok := ImageExt[mediatype]
	if !ok {
		return fmt.Errorf("unsupported mediatype: %s", mediatype)
	}
	f, err := ioutil.TempFile("", "*." + ext)
	log.Info().Str("file", f.Name()).Msg("image temp file name")
	r := bufio.NewReader(resp.Body)
	_, err = r.WriteTo(f)
	if err != nil {
		return err
	}
	err = g.d.Set(src,1)
	return err

}

func (g *Gorobei) TestImage() error {
	// Send using file from disk
	//f, err := os.Open("/tmp/996584576.jpeg")
	f, err := os.Open("c:/personal/1430353161-0a6e350da88a8d827765cc8fe6dccd70.jpg")
	if err != nil {
		return err
	}
	log.Info().Str("chat", g.chat).Msg("sending test image...")
	params := &telego.SendPhotoParams{
		ChatID:                   telego.ChatID{ ID: g.chatID },
		Photo:                    telego.InputFile{ File: f},
		Caption:                  "https://test.ru/test_url",
	}
	_, err = g.bot.SendPhoto(params)
	return  err
}

func (g *Gorobei) TestMessage(chat string, message string) error {
	params := &telego.SendMessageParams{
		ChatID:                   telego.ChatID{
			ID: g.chatID,
		},
		Text:                     message,
	}
	_, err := g.bot.SendMessage(params)
	return err
}

