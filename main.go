package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/alecthomas/kingpin"
)

var (
	prmStartTime string
	prmPause     int
)

func getFirstSyslogEntry() (time.Time, error) {
	var startTime time.Time

	syslog, err := os.Open("/var/log/syslog")
	if err != nil {
		return time.Time{}, fmt.Errorf("could not read /var/log/syslog")
	}

	scanner := bufio.NewScanner(syslog)

	for scanner.Scan() {
		if strings.Contains(scanner.Text(), time.Now().Format("Jan 2")) {
			dateTokens := strings.SplitN(scanner.Text(), ":", 3)
			dateString := strings.Join(dateTokens[:2], " ")
			startTime, err = time.Parse("Jan 2 15 04", dateString)
			if err != nil {
				break
			}

			startTime = time.Date(
				time.Now().Year(), startTime.Month(), startTime.Day(),
				startTime.Hour(), startTime.Minute(), 0, 0, time.Local)

			return startTime, nil
		}
	}

	return time.Time{}, fmt.Errorf("could not parse time in syslog")
}

func parseGivenTime() (time.Time, error) {
	givenTime, err := time.Parse("15:04", prmStartTime)
	if err != nil {
		return time.Time{}, fmt.Errorf("given time not in format hh:mm")
	}

	y, m, d := time.Now().Date()
	return time.Date(y, m, d, givenTime.Hour(), givenTime.Minute(), 0, 0, time.Local), nil
}

func getStartTime() time.Time {
	var startTime time.Time
	var err error

	if len(prmStartTime) == 0 {
		startTime, err = getFirstSyslogEntry()
	} else {
		startTime, err = parseGivenTime()
	}

	if err != nil {
		log.Fatalf("unable to get start time: %s", err)
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
