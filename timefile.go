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
	path      string
	startTime time.Time
	pause     int
}

func newTimefile(path string) *timefile {
	timeFile := new(timefile)
	timeFile.path = path
	return timeFile
}

func (t *timefile) setStartTime(setTime string) {
	st, err := time.Parse("15:04", setTime)
	if err != nil {
		log.Fatalf("could not parse supplied start time: %s", setTime)
	}

	t.startTime = insertInToday(st)
}

func (t *timefile) setPause(pause int) {
	t.pause = pause
}

func (t *timefile) store() {
	writer, err := os.Create(t.path)
	if err != nil {
		log.Fatal(err)
	}

	_, err = fmt.Fprintf(writer, "%s\n%d\n", t.startTime.Format("15:04"), t.pause)
	if err != nil {
		log.Fatal(err)
	}
}

func (t *timefile) read() {
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
		log.Fatal("could not read pause value from timefile")
	}

	t.startTime = startTime
	t.pause = pause
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

func (t *timefile) buildTimeStruct() *timestruct {
	target := new(timestruct)

	target.goHomeAt = t.startTime.Add(8 * time.Hour).Add(time.Duration(prmPause) * time.Minute)
	target.goHomeLatest = t.startTime.Add(10 * time.Hour).Add(longer(45, prmPause) * time.Minute)

	target.goHomeIn = time.Until(target.goHomeAt)
	target.goLatestIn = time.Until(target.goHomeLatest)

	return target
}
