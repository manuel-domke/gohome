package main

import (
	wf "github.com/danielb42/whiteflag"
)

func main() {
	wf.Alias("s", "start", "start time (hh:mm)")
	wf.Alias("p", "pause", "duration of break(s) in min.")
	wf.Alias("o", "offset", "time you need from door to booting your pc in min.")
	wf.Alias("r", "reset", "reset the timefile")
	wf.ParseCommandLine()

	timeStruct := newTimestruct()

	if wf.CheckBool("reset") {
		timeStruct.remove()
	}

	if timeStruct.timeFileisOfToday() {
		timeStruct.read()
	}

	if wf.CheckString("start") {
		timeStruct.setStartTime(wf.GetString("start"))
	} else if !timeStruct.timeFileisOfToday() {
		timeStruct.setStartTime(getResumeTimeFromJournal())
	}

	if wf.CheckInt("pause") {
		timeStruct.setPause(wf.GetInt("pause"))
	} else if !timeStruct.timeFileisOfToday() {
		timeStruct.setPause(60)
	}

	if wf.CheckInt("offset") {
		timeStruct.setOffset(wf.GetInt("offset"))
	} else if !timeStruct.timeFileisOfToday() {
		timeStruct.setOffset(3)
	}

	timeStruct.store()

	timeStruct.calculateDeadlines()
	timeStruct.print()
}
