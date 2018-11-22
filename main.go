package main

import (
	"fmt"
	"log"
	"syscall"
	"time"

	"github.com/alecthomas/kingpin"
)

var (
	prmStartTime string
	prmPause     int
)

func getStartTime() time.Time {
	if len(prmStartTime) == 0 {
		si := &syscall.Sysinfo_t{}
		syscall.Sysinfo(si)
		return time.Now().Add(time.Duration(-si.Uptime) * time.Second)
	}

	givenTime, err := time.Parse("15:04", prmStartTime)
	if err != nil {
		log.Fatal("could not parse start time, use format hh:mm")
	}

	y, m, d := time.Now().Date()
	startTime := time.Date(y, m, d, givenTime.Hour(), givenTime.Minute(), 0, 0, time.Local)

	wasYesterday := false

	if givenTime.Hour() > time.Now().Hour() {
		wasYesterday = true
	}

	if givenTime.Hour() == time.Now().Hour() &&
		givenTime.Minute() > time.Now().Minute() {
		wasYesterday = true
	}

	if wasYesterday {
		startTime = startTime.Add(time.Duration(-24) * time.Hour)
	}

	return startTime
}

func main() {
	log.SetFlags(0)
	kingpin.Flag("start", "start time (hh:mm)").Short('s').StringVar(&prmStartTime)
	kingpin.Flag("pause", "duration of break(s)").Short('p').Default("60").IntVar(&prmPause)
	kingpin.CommandLine.HelpFlag.Hidden()
	kingpin.Parse()

	startTime := getStartTime()

	if prmPause < 30 {
		prmPause = 30
	}

	goHomeAt := startTime.Add(time.Duration(8) * time.Hour).Add(time.Duration(prmPause) * time.Minute)
	goHomeIn := time.Until(goHomeAt)

	fmt.Printf("started at %s\n", startTime.Format("15:04"))
	fmt.Printf("day complete at %s (with %d min. break)\n", goHomeAt.Format("15:04"), prmPause)

	if goHomeIn.Minutes() >= 0 {
		fmt.Printf("that is in %.f minutes\n", goHomeIn.Minutes())
	} else {
		fmt.Printf("that was %.f minutes ago\n", goHomeIn.Minutes()*-1)
	}

	if prmPause < 45 {
		prmPause = 45
	}

	goHomeLatest := startTime.Add(time.Duration(10) * time.Hour).Add(time.Duration(prmPause) * time.Minute)
	fmt.Printf("leave latest at %s\n", goHomeLatest.Format("15:04"))
}
