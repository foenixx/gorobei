package main

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestDb_Set(t *testing.T) {
	initLog(false)
	path := "./gorobei_test_db"
	d, err := OpenDb(path)
	require.NoError(t, err)
	defer os.RemoveAll(path)
	defer d.Close()

	err = d.Set("test", 1)
	require.NoError(t, err)
	v, err := d.Get("test")
	require.NoError(t, err)
	assert.Equal(t, byte(1), v)
	v, err = d.Get("test1")
	assert.ErrorIs(t, err, ErrNotFound)

}

func TestDb_SetGetUserID(t *testing.T) {
	initLog(false)
	path := "./gorobei_test_db"
	d, err := OpenDb(path)
	require.NoError(t, err)
	defer os.RemoveAll(path)
	defer d.Close()

	err = d.SetUserID("someuser", 100)
	require.NoError(t, err)

	v, err := d.GetUserID("someuser")
	require.NoError(t, err)
	assert.Equal(t, int64(100), v)

	err = d.SetUserID("someuser", -100)
	require.NoError(t, err)

	v, err = d.GetUserID("someuser")
	require.NoError(t, err)
	assert.Equal(t, int64(-100), v)

	v, err = d.GetUserID("nosuchuser")
	assert.ErrorIs(t, err, ErrNotFound)

}