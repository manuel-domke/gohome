package main

import (
	"github.com/alecthomas/kingpin"
)

func main() {
	var (
		prmStartTime string
		prmPause     int
		prmOffset    int
		prmReset     bool
	)

	kingpin.Flag("start", "start time (hh:mm)").
		Short('s').StringVar(&prmStartTime)
	kingpin.Flag("pause", "duration of break(s) in min.").
		Short('p').IntVar(&prmPause)
	kingpin.Flag("offset", "time you need from door to booting your pc in min.").
		Short('o').IntVar(&prmOffset)
	kingpin.Flag("reset", "reset the timefile").
		Short('r').BoolVar(&prmReset)
	kingpin.Parse()

	timeStruct := newTimestruct()

	if prmReset {
		timeStruct.remove()
	}

	if timeStruct.timeFileisOfToday() {
		timeStruct.read()
	}

	if len(prmStartTime) > 0 {
		timeStruct.setStartTime(prmStartTime)
	} else if !timeStruct.timeFileisOfToday() {
		timeStruct.setStartTime(getResumeTimeFromJournal())
	}

	if prmPause > 0 {
		timeStruct.setPause(prmPause)
	} else if !timeStruct.timeFileisOfToday() {
		timeStruct.setPause(60)
	}

	if prmOffset > 0 {
		timeStruct.setOffset(prmOffset)
	} else if !timeStruct.timeFileisOfToday() {
		timeStruct.setOffset(3)
	}

	timeStruct.store()

	timeStruct.calculateDeadlines()
	timeStruct.print()
}
