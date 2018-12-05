package main

import (
	"log"
	"os"
	"time"
)

type Timefile struct {
	Path string
	Stat os.FileInfo
}

func NewTimefile(path string) *Timefile {
	timeFile := new(Timefile)
	timeFile.Path = path
	timeFile.Stat, _ = os.Stat(path)
	return timeFile
}

func (t Timefile) Set(setTime time.Time) {
	os.Create(t.Path)
	os.Chtimes(t.Path, setTime, setTime)
}

func (t Timefile) Get() time.Time {
	stat, _ := os.Stat(t.Path)
	return stat.ModTime()
}

func (t Timefile) IsOfToday() bool {
	mtime := t.Stat.ModTime()
	now := time.Now()

	return mtime.Year() == now.Year() &&
		mtime.Month() == now.Month() &&
		mtime.Day() == now.Day()
}

func (t Timefile) Remove() {
	if err := os.Remove(t.Path); err != nil {
		log.Fatalf("could not remove %s", t.Path)
	}

	os.Exit(0)
}
