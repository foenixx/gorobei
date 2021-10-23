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

type fetcher struct {}
var _ HttpFetcher = (*fetcher)(nil)

func (f *fetcher) FetchHtml(url string) (string, error) {
	panic("implement me")
}

func (f *fetcher) FetchImage(url string) (string, error) {
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
		fetcher: &fetcher{},
		clock: &clock.TestClock{Tm: time.Now()},
		chat:         "test_chat",
		admin:        "test_admin",
		adminId: 776}

	return g, closeFunc
}

func TestGorobei_UpdateAndSendDailyReport(t *testing.T) {
	g, f := newTestGorobei(t)
	defer f()
	clk := g.clock.(*clock.TestClock)
	tg := g.tg.(*telega)

	time1, _ := time.Parse(time.Stamp, "Jan  1 14:01:02")
	clk.Tm = time1
	err := g.UpdateAndSendDailyReport(20, 19, 1, "last error 1")
	require.NoError(t, err)
	require.Empty(t, tg.msg)

	time2, _ := time.Parse(time.Stamp, "Jan  1 23:01:00")
	clk.Tm = time2
	err = g.UpdateAndSendDailyReport(21, 18, 3, "last error 2")
	require.NoError(t, err)
	require.NotEmpty(t, tg.msg)
	require.Equal(t, tg.msg, g.FormatDailyReport(&DailyReport{1,1,1,20,"last error 1", time2}))

	tg.msg = ""
	time3, _ := time.Parse(time.Stamp, "Jan  1 23:15:00")
	clk.Tm = time3
	err = g.UpdateAndSendDailyReport(21, 18, 3, "last error 3")
	require.NoError(t, err)
	require.Empty(t, tg.msg)
}

func TestGorobei_UpdateAndSendDailyReport2(t *testing.T) {
	g, f := newTestGorobei(t)
	defer f()
	clk := g.clock.(*clock.TestClock)

	time1, _ := time.Parse(time.Stamp, "Jan  1 14:01:02")
	clk.Tm = time1
	err := g.UpdateAndSendDailyReport(20, 19, 1, "last error 1")
	require.NoError(t, err)
	r, err := g.d.ReadDailyReport()
	require.NoError(t, err)
	require.Equal(t, &DailyReport{1,1,1,20,"last error 1",time1}, r)

	time2, _ := time.Parse(time.Stamp, "Jan  1 23:01:00")
	clk.Tm = time2
	err = g.UpdateAndSendDailyReport(21, 18, 3, "last error 2")
	require.NoError(t, err)
	r, err = g.d.ReadDailyReport()
	require.NoError(t, err)
	require.Equal(t, &DailyReport{1,3,3,21,"last error 2",time2}, r)
}