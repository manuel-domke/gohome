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

var (
	prmStartTime string
	prmPause     int
	prmOffset    int
	prmReset     bool
)

type timestruct struct {
	goHomeAt, goHomeLatest time.Time
	goHomeIn, goLatestIn   time.Duration
}

func main() {
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

	if len(prmStartTime) == 0 && prmPause == 0 {
		if !timeFile.isOfToday() {
			timeFile.setStartTime(getResumeTimeFromJournal())
			timeFile.setPause(60)
			timeFile.store()
		}
	} else if len(prmStartTime) == 0 && prmPause > 0 {
		if !timeFile.isOfToday() {
			timeFile.setStartTime(getResumeTimeFromJournal())
			timeFile.setPause(max(prmPause, 30))
			timeFile.store()
		} else {
			timeFile.setPause(max(prmPause, 30))
			timeFile.store()
		}
	} else if len(prmStartTime) > 0 && prmPause == 0 {
		if !timeFile.isOfToday() {
			timeFile.setStartTime(prmStartTime)
			timeFile.setPause(60)
			timeFile.store()
		} else {
			timeFile.setStartTime(prmStartTime)
			timeFile.store()
		}
	} else {
		timeFile.setStartTime(prmStartTime)
		timeFile.setPause(max(prmPause, 30))
		timeFile.store()
	}

	timeFile.read()

	timeFile.startTime = timeFile.startTime.Add(time.Duration(prmOffset*-1) * time.Minute)
	timeStruct := timeFile.buildTimeStruct()

	output(timeFile, timeStruct)
}
func output(timeFile *timefile, timeStruct *timestruct) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.AlignRight)

	fmt.Fprintf(w, "started work at\t %s", c.Bold(c.Gray(timeFile.startTime.Format("15:04"))))
	fmt.Fprintf(w, " (includes %d min. offset)\n\n", c.Bold(prmOffset))
	fmt.Fprintf(w, "day complete at\t %s (includes %d min. break)\n",
		c.Bold(c.Cyan(timeStruct.goHomeAt.Format("15:04"))),
		c.Brown(prmPause),
	)

	if timeStruct.goHomeIn.Minutes() >= 0 {
		fmt.Fprintf(w, "...that's in\t %s\n", c.Bold(c.Cyan(printDuration(timeStruct.goHomeIn))))
	} else {
		fmt.Fprintf(w, "...that was\t %s ago\n", c.Bold(c.Green(printDuration(timeStruct.goHomeIn))))
	}

	fmt.Fprintf(w, "\nleave latest at\t %s\n", c.Red(timeStruct.goHomeLatest.Format("15:04")))

	if timeStruct.goLatestIn.Minutes() >= 0 {
		fmt.Fprintf(w, "...that's in\t %s\n", c.Red(printDuration(timeStruct.goLatestIn)))
	} else {
		fmt.Fprintf(w, "...that was\t %s ago\n", c.Bold(c.Red(printDuration(timeStruct.goLatestIn))))
	}

	if err := w.Flush(); err != nil {
		log.Fatal(err)
	}
}

func longer(a int, b int) time.Duration {
	if a > b {
		return time.Duration(a)
	}

	return time.Duration(b)
}

func max(a int, b int) int {
	if a > b {
		return a
	}

	return b
}

func printDuration(dur time.Duration) string {
	h := int(dur.Hours())
	m := int(dur.Minutes()) - 60*h
	return strings.Replace(fmt.Sprintf("%dh%dm", h, m), "-", "", -1)
}
