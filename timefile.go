package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
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

func (t *timefile) set(setTime time.Time, pause int) {
	var err error
	var writer *os.File

	writer, err = os.Create(t.path)
	if err != nil {
		log.Fatal(err)
	}

	_, err = fmt.Fprintf(writer, "%d\n", pause)
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

func (t *timefile) getPauseFromFile() (int, error) {
	file, err := os.Open(t.path)
	if err != nil {
		return 0, fmt.Errorf("could not open timefile")
	}

	scanner := bufio.NewScanner(file)
	scanner.Scan()

	pause, err := strconv.Atoi(scanner.Text())
	if err != nil {
		return 0, fmt.Errorf("could not read number from timefile")
	}

	return pause, nil
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
