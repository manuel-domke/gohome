package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/alecthomas/kingpin"
	c "github.com/logrusorgru/aurora"
)

func parseGivenTime(prmStartTime string) time.Time {
	givenTime, err := time.Parse("15:04", prmStartTime)
	if err != nil {
		log.Fatal(err)
	}

	return time.Date(
		time.Now().Year(), time.Now().Month(), time.Now().Day(),
		givenTime.Hour(), givenTime.Minute(), 0,
		0, time.Local,
	)
}

func main() {
	var (
		startTime    time.Time
		prmStartTime string
		prmPause     int
		prmOffset    int
		prmReset     bool
	)

	timeFile := newTimefile(os.Getenv("HOME") + "/.gohome")

	log.SetFlags(0)
	kingpin.Flag("start", "start time (hh:mm)").
		Short('s').StringVar(&prmStartTime)
	kingpin.Flag("pause", "duration of break(s) in min.").
		Short('p').IntVar(&prmPause)
	kingpin.Flag("offset", "time you need from door to booting your pc in min.").
		Short('o').Default("3").IntVar(&prmOffset)
	kingpin.Flag("reset", "reset the timefile").
		Short('r').BoolVar(&prmReset)
	kingpin.Parse()

	if prmReset {
		timeFile.remove()
	}

	if prmPause == 0 {
		var err error
		prmPause, err = timeFile.getPauseFromFile()
		if err != nil {
			log.Println("warning:", err)
		}
	}

	if prmPause < 30 {
		prmPause = 30
	}

	if len(prmStartTime) == 0 {
		if !timeFile.isOfToday() {
			timeFile.set(getResumeTimeFromJournal(), prmPause)
		}

		startTime = timeFile.get()
	} else {
		startTime = parseGivenTime(prmStartTime)
		timeFile.set(startTime, prmPause)
	}

	startTime = startTime.Add(time.Duration(prmOffset*-1) * time.Minute)

	goHomeAt := startTime.Add(8 * time.Hour).Add(time.Duration(prmPause) * time.Minute)
	goHomeLatest := startTime.Add(10 * time.Hour).Add(longer(45, prmPause) * time.Minute)

	goHomeIn := time.Until(goHomeAt)
	goLatestIn := time.Until(goHomeLatest)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.AlignRight)

	fmt.Fprintf(w, "started work at\t %s", c.Bold(c.Gray(startTime.Format("15:04"))))
	fmt.Fprintf(w, " (includes %d min. offset)\n\n", c.Bold(prmOffset))
	fmt.Fprintf(w, "day complete at\t %s (includes %d min. break)\n",
		c.Bold(c.Cyan(goHomeAt.Format("15:04"))),
		c.Brown(prmPause),
	)

	if goHomeIn.Minutes() >= 0 {
		fmt.Fprintf(w, "...that's in\t %s\n", c.Bold(c.Cyan(printDuration(goHomeIn))))
	} else {
		fmt.Fprintf(w, "...that was\t %s ago\n", c.Bold(c.Green(printDuration(goHomeIn))))
	}

	fmt.Fprintf(w, "\nleave latest at\t %s\n", c.Red(goHomeLatest.Format("15:04")))

	if goLatestIn.Minutes() >= 0 {
		fmt.Fprintf(w, "...that's in\t %s\n", c.Red(printDuration(goLatestIn)))
	} else {
		fmt.Fprintf(w, "...that was\t %s ago\n", c.Bold(c.Red(printDuration(goLatestIn))))
	}

	if err := w.Flush(); err != nil {
		log.Fatal(err)
	}
}

func longer(a, b int) time.Duration {
	if a > b {
		return time.Duration(a)
	}

	return time.Duration(b)
}

func printDuration(dur time.Duration) string {
	h := int(dur.Hours())
	m := int(dur.Minutes()) - 60*h
	return strings.Replace(fmt.Sprintf("%dh%dm", h, m), "-", "", -1)
}
