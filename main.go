package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/alecthomas/kingpin"
	c "github.com/logrusorgru/aurora"
)

var (
	prmStartTime string
	prmPause     int
	prmOffset    int
	prmReset     bool
	timefilePath = os.Getenv("HOME") + "/.gohome"
)

func getEarliestSyslogToday(syslogPath string) time.Time {
	var startTime time.Time

	syslog, err := os.Open(syslogPath)
	if err != nil {
		log.Fatalf("could not read %s", syslogPath)
	}

	scanner := bufio.NewScanner(syslog)

	for scanner.Scan() {
		if strings.Contains(scanner.Text(), time.Now().Format("Jan 2")) {
			dateTokens := strings.SplitN(scanner.Text(), ":", 3)
			dateString := strings.Join(dateTokens[:2], " ")
			startTime, err = time.Parse("Jan 2 15 04", dateString)
			if err != nil {
				log.Fatal("could not parse time in syslog")
			}

			startTime = time.Date(
				time.Now().Year(), startTime.Month(), startTime.Day(),
				startTime.Hour(), startTime.Minute(), 0, 0, time.Local,
			)

			break
		}
	}

	return startTime
}

func parseGivenTime() time.Time {
	givenTime, err := time.Parse("15:04", prmStartTime)
	if err != nil {
		log.Fatal("given time is not in format hh:mm")
	}

	y, m, d := time.Now().Date()
	return time.Date(
		y, m, d,
		givenTime.Hour(), givenTime.Minute(), 0,
		0, time.Local,
	)
}

func getStartTime() time.Time {
	stat, err := os.Stat(timefilePath)
	if !os.IsNotExist(err) && checkIfTimefileIsOfToday(stat) {
		return stat.ModTime()
	}

	var startTime time.Time

	if len(prmStartTime) == 0 {
		startTime = earliest(
			getEarliestSyslogToday("/var/log/syslog"),
			getEarliestSyslogToday("/var/log/syslog.1"),
		)
	} else {
		startTime = parseGivenTime()
	}

	startTime = startTime.Add(time.Duration(prmOffset*-1) * time.Minute)
	touchTimefile(startTime)
	return startTime
}

func checkIfTimefileIsOfToday(stat os.FileInfo) bool {
	mtime := stat.ModTime()
	now := time.Now()

	return mtime.Year() == now.Year() &&
		mtime.Month() == now.Month() &&
		mtime.Day() == now.Day()
}

func touchTimefile(startTime time.Time) {
	_, err := os.Create(timefilePath)
	if err != nil {
		log.Fatal("could not edit timefile")
	}

	err = os.Chtimes(timefilePath, startTime, startTime)
	if err != nil {
		log.Fatal("could not set times on timefile")
	}
}

func resetTimefile() {
	if err := os.Remove(timefilePath); err != nil {
		log.Fatalf("could not remove %s", timefilePath)
	}

	os.Exit(0)
}

func main() {
	log.SetFlags(0)
	kingpin.Flag("start", "start time (hh:mm)").
		Short('s').StringVar(&prmStartTime)
	kingpin.Flag("pause", "duration of break(s) in min.").
		Short('p').Default("60").IntVar(&prmPause)
	kingpin.Flag("offset", "time you need from door to booting your pc in min.").
		Short('o').Default("3").IntVar(&prmOffset)
	kingpin.Flag("reset", "reset the timefile").
		Short('r').BoolVar(&prmReset)
	kingpin.Parse()

	if prmReset {
		resetTimefile()
	}

	if prmPause < 30 {
		prmPause = 30
	}

	startTime := getStartTime()

	goHomeAt := startTime.Add(8 * time.Hour).Add(time.Duration(prmPause) * time.Minute)
	goHomeIn := time.Until(goHomeAt)
	goHomeLatest := startTime.Add(10 * time.Hour).Add(longer(45, prmPause) * time.Minute)
	goLatestIn := time.Until(goHomeLatest)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.AlignRight)
	defer w.Flush()

	fmt.Fprintf(w, "started work at\t %s\n\n", c.Gray(startTime.Format("15:04")))
	fmt.Fprintf(w, "day complete at\t %s (includes %d min. break)\n",
		c.Bold(c.Cyan(goHomeAt.Format("15:04"))),
		c.Brown(prmPause),
	)

	if goHomeIn.Minutes() >= 0 {
		fmt.Fprintf(w, "...that's in\t %.f min\n", c.Bold(c.Green(goHomeIn.Minutes())))
	} else {
		fmt.Fprintf(w, "...that was\t %.f min ago\n", c.Bold(c.Green(goHomeIn.Minutes()*-1)))
	}

	fmt.Fprintf(w, "\nleave latest at\t %s\n", c.Red(goHomeLatest.Format("15:04")))
	fmt.Fprintf(w, "...that's in\t %.f min\n", c.Red(goLatestIn.Minutes()))
}

func longer(a, b int) time.Duration {
	if a > b {
		return time.Duration(a)
	}

	return time.Duration(b)
}

func earliest(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}

	return b
}
