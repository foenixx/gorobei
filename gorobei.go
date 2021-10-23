package main

import (
	"errors"
	"fmt"
	"github.com/phuslu/log"
	"gorobei/clock"
	"gorobei/utils"
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
	tg      Telegram
	clock   clock.Clock
	fetcher HttpFetcher
}

var ErrImageAlreadyProcessed = errors.New("image has been processed already")

func (g *Gorobei) Init() error {
	chat, err := g.tg.ChatInfo(g.chat)
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
	body, err := g.fetcher.FetchHtml(url)

	if err != nil {
		log.Error().Err(err).Msg("cannot fetch page content")
		er2 := g.UpdateAndSendDailyReport(0, 0, 0, err.Error())
		if er2 != nil {
			// return http error, possible UpdateAndSendDailyReport errors are just logged
			log.Error().Err(er2).Msg("cannot update daily report")
		}
		return err
	}

	re := regexp.MustCompile(`(?si)<div class="singlePost.*?<div class="postInner">\s*?<div class="paragraph">[^<]*<div[^<]*<img src=["']{1}(.*?)["']{1}`)
	ma := re.FindAllStringSubmatch(string(body), -1)

	var (
		total, skipped, errc int
		lastError            string
	)
	for _, m := range ma {
		if len(m) == 2 && m[1] != `https://i.imgur.com/sMhpFyR.jpg` {
			src := m[1]
			log.Info().Str("src", src).Msg("image found")
			err = g.processImage(src)
			if err != nil {
				if errors.Is(err, ErrImageAlreadyProcessed) {
					skipped += 1
				} else {
					lastError = err.Error()
					errc += 1
					log.Error().Err(err).Str("src", src).Msg("cannot process image")
					msg := fmt.Sprintf("Error occured during processing the image!\n[image](%s)\n\n__error__:\n```\n%s\n```", src, err.Error())
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
			log.Error().Str("src", m[0]).Msg("unexpected match")
		}
	}
	// update and send daily report, errors are just logged
	err = g.UpdateAndSendDailyReport(total, skipped, errc, lastError)
	if err != nil {
		log.Error().Err(err).Msg("cannot update daily report")
	}
	// notify admin about errors or new images posted
	if errc > 0 || (total-skipped) > 0 {
		msg := fmt.Sprintf("Fetching images from the [page](%s) completed.\nTotal: %v\nNew: %v\nSkipped: %v\nErrors: %v", url, total, total-skipped, skipped, errc)
		err = g.SendAdminMessage(msg)
		return err
	}
	return nil
}

func (g *Gorobei) FormatDailyReport(r *DailyReport) string {
	msg := utils.Bt(
		`*Daily report.*

The bot has run *%v* times.
New images posted: *%v*
Images found (during last run): *%v*
Errors (during last run): *%v*
Last error:
³³³
%v
³³³`)
	return fmt.Sprintf(msg, r.Run, r.Posted, r.Total, r.Errors, r.LastError)
}

func (g *Gorobei) ReadOrCreateDailyReport() (*DailyReport, error) {
	r, err := g.d.ReadDailyReport()
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			// no report yet
			return &DailyReport{SentAt: g.clock.Now()}, nil
		}
		return nil, err
	}
	return r, nil
}

func (g *Gorobei) UpdateAndSendDailyReport(total, skipped, errc int, lastError string) error {
	r, err := g.ReadOrCreateDailyReport()
	if err != nil {
		return err
	}
	// daily reports are sent after 23:00
	y, m, d := r.SentAt.Date()
	planned := time.Date(y, m, d, 23, 0, 0, 0, r.SentAt.Location())
	if r.SentAt.Sub(planned) > 0 {
		// planned < r.SentAt
		planned = planned.Add(24 * time.Hour)
	}
	// if Now() > planned send date?
	if g.clock.Now().Sub(planned) > 0 {
		// send the report
		err = g.SendAdminMessage(g.FormatDailyReport(r))
		if err != nil {
			return err
		}
		// store new values
		r.Run = 1
		r.Total = total
		r.Posted = total - skipped
		r.Errors = errc
		r.LastError = lastError
		r.SentAt = g.clock.Now()
		return g.d.StoreDailyReport(r)
	}

	// just update values
	r.Posted += total - skipped
	r.Errors = errc
	r.Total = total
	r.LastError = lastError
	r.Run += 1
	return g.d.StoreDailyReport(r)
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
	path, err := g.fetcher.FetchImage(src)
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
	return g.tg.SendMessageText("", g.chatId, message)
}

func (g *Gorobei) SendAdminMessage(message string) error {
	if g.adminId == 0 {
		return nil
	}

	return g.tg.SendMessageMarkdown("", g.adminId, message)
}

func (g *Gorobei) ForgetImg(url string) error {
	return g.d.ReadUrlProcessed(url, 0)
}
