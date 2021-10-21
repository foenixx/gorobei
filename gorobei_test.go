package main

import (
	"github.com/mymmrac/telego"
	"github.com/stretchr/testify/require"
	"gorobei/clock"
	"os"
	"testing"
	"time"
)

type telega struct {
	msg string
}

func (t *telega) SendImage(user string, userId int64, image string, caption string) error {
	panic("implement me")
}

func (t *telega) SendMessageMarkdown(user string, userId int64, message string) error {
	t.msg = message
	return nil
}

func (t *telega) SendMessageText(user string, userId int64, message string) error {
	panic("implement me")
}

func (t *telega) DoUpdates() error {
	panic("implement me")
}

func (t telega) ChatInfo(s string) (*telego.Chat, error) {
	panic("implement me")
}

func openTestDb(t *testing.T) (*Db, func()) {
	path := "./gorobei_test_db"
	d, err := OpenDb(path)
	require.NoError(t, err)
	return d, func() {
		d.Close()
		os.RemoveAll(path)
	}
}
func newTestGorobei(t* testing.T) (*Gorobei, func()) {
	db, closeFunc := openTestDb(t)

	g := &Gorobei{d: db,
		tg: &telega{},
		clock: &clock.TestClock{Tm: time.Now()},
		chat:         "test_chat",
		admin:        "test_admin",
		adminId: 776}

	return g, closeFunc
}

func Test_updateDailyReport(t *testing.T) {
	g, f := newTestGorobei(t)
	defer f()
	clk := g.clock.(*clock.TestClock)
	tg := g.tg.(*telega)
	clk.Tm, _ = time.Parse(time.Stamp, "Jan  1 14:01:02")
	err := g.updateDailyReport(20, 19, 1)
	require.NoError(t, err)
	require.Empty(t, tg.msg)
	clk.Tm, _ = time.Parse(time.Stamp, "Jan  1 23:01:00")
	err = g.updateDailyReport(21, 18, 3)
	require.NoError(t, err)
	require.NotEmpty(t, tg.msg)
	r, err := g.d.ReadDailyReport()
	require.NoError(t, err)
	rE := DailyReport{
		Run:    0,
		Posted: 0,
		Errors: 3,
		Total:  21,
	}
	rE.SentAt, _ = time.Parse(time.Stamp, "Jan  1 23:01:00")
	require.Equal(t, rE, r)
}