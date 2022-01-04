package main

import (
	"testing"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"gotest.tools/assert"
)

func Test_db(t *testing.T) {
	for _, dsn := range []string{
		"file::memory:?cache=shared",
	} {
		db, err := openDB(dsn)
		assert.NilError(t, err)
		assert.NilError(t, db.Ping())
	}

	for _, dsn := range []string{
		":::::",
	} {
		_, err := openDB(dsn)
		assert.Assert(t, err != nil)
	}
	for _, dsn := range []string{
		"file:///no/such/file",
	} {
		_, err := openDB(dsn)
		assert.Assert(t, err != nil)
	}
	for _, dsn := range []string{
		"postgres://bad:user@no.such.host/nonsense",
	} {
		_, err := openDB(dsn)
		assert.Assert(t, err != nil)
	}
}
