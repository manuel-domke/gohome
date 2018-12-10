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
}

func newTimefile(path string) *timefile {
	timeFile := new(timefile)
	timeFile.path = path
	return timeFile
}

func (t *timefile) set(setTime string, pause int) {
	var err error
	var writer *os.File

	writer, err = os.Create(t.path)
	if err != nil {
		log.Fatal(err)
	}

	_, err = time.Parse("15:04", setTime)
	if err != nil {
		log.Fatal(err)
	}

	_, err = fmt.Fprintf(writer, "%s\n%d\n", setTime, pause)
	if err != nil {
		log.Fatal(err)
	}
}

func (t *timefile) read() (time.Time, int) {
	file, err := os.Open(t.path)
	if err != nil {
		log.Fatal("could not open timefile")
	}

	scanner := bufio.NewScanner(file)

	scanner.Scan()
	startTime, err := time.Parse("15:04", scanner.Text())
	if err != nil {
		log.Fatal("could not parse time in timefile")
	}

	scanner.Scan()
	pause, err := strconv.Atoi(scanner.Text())
	if err != nil {
		log.Fatal("could not pause value from timefile")
	}

	return insertInToday(startTime), pause
}

func (t *timefile) isOfToday() bool {
	stat, err := os.Stat(t.path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}

		log.Fatal(err)
	}

	mtime := stat.ModTime()
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

func insertInToday(hourAndMin time.Time) time.Time {
	now := time.Now()
	return time.Date(
		now.Year(), now.Month(), now.Day(),
		hourAndMin.Hour(), hourAndMin.Minute(),
		0, 0, time.Local,
	)
}
