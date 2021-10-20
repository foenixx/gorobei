package main

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

func testInit(t *testing.T) (*Db, func()) {
	initLog(false)
	path := "./gorobei_test_db"
	d, err := OpenDb(path)
	require.NoError(t, err)
	return d, func() {
		d.Close()
		os.RemoveAll(path)
	}
}

func TestDb_Set(t *testing.T) {
	d, deferFunc := testInit(t)
	defer deferFunc()
	err := d.ReadUrlProcessed("test", 1)
	require.NoError(t, err)
	v, err := d.StoreUrlProcessed("test")
	require.NoError(t, err)
	assert.Equal(t, byte(1), v)
	v, err = d.StoreUrlProcessed("test1")
	assert.ErrorIs(t, err, ErrNotFound)

}

func TestDb_SetGetUserID(t *testing.T) {
	d, deferFunc := testInit(t)
	defer deferFunc()

	err := d.StoreUserId("someuser", 100)
	require.NoError(t, err)

	v, err := d.ReadUserId("someuser")
	require.NoError(t, err)
	assert.Equal(t, int64(100), v)

	err = d.StoreUserId("someuser", -100)
	require.NoError(t, err)

	v, err = d.ReadUserId("someuser")
	require.NoError(t, err)
	assert.Equal(t, int64(-100), v)

	v, err = d.ReadUserId("nosuchuser")
	assert.ErrorIs(t, err, ErrNotFound)

}

func TestDb_DailyReport(t *testing.T) {
	d, deferFunc := testInit(t)
	defer deferFunc()
	tm, _ := time.Parse(time.Stamp, "Jan 2 15:04:05")
	r := DailyReport{
		Posted: 100,
		Errors: 1,
		SentAt: tm,
		Date:   tm,
	}
	_, err := d.ReadDailyReport()
	require.ErrorIs(t, ErrNotFound, err)
	err = d.StoreDailyReport(&r)
	require.NoError(t, err)
	stored, err := d.ReadDailyReport()
	require.NoError(t, err)
	require.Equal(t, r, *stored)
}