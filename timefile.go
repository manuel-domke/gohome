package main

import (
	"log"
	"os"
	"time"
)

type timefile struct {
	path string
	stat os.FileInfo
}

func newTimefile(path string) *timefile {
	var err error

	timeFile := new(timefile)
	timeFile.path = path
	timeFile.stat, err = os.Stat(path)

	if err != nil && !os.IsNotExist(err) {
		log.Fatal(err)
	}

	return timeFile
}

func (t *timefile) set(setTime time.Time) {
	var err error

	_, err = os.Create(t.path)
	if err != nil {
		log.Fatal(err)
	}

	err = os.Chtimes(t.path, setTime, setTime)
	if err != nil {
		log.Fatal(err)
	}

	t.stat, err = os.Stat(t.path)
	if err != nil {
		log.Fatal(err)
	}
}

func (t *timefile) get() time.Time {
	return t.stat.ModTime()
}

func (t *timefile) isOfToday() bool {
	if t.stat == nil {
		return false
	}

	mtime := t.stat.ModTime()
	now := time.Now()

	return mtime.Year() == now.Year() &&
		mtime.Month() == now.Month() &&
		mtime.Day() == now.Day()
}

func (t *timefile) remove() {
	err := os.Remove(t.path)

	if err != nil && !os.IsNotExist(err) {
		log.Fatal(err)
	}

	os.Exit(0)
}
