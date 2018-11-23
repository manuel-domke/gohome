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
	kingpin.Flag("pause", "duration of break(s) in min.").Short('p').Default("60").IntVar(&prmPause)
	kingpin.Flag("offset", "time you need from door to booting your pc in min.").Short('o').Default("5").IntVar(&prmOffset)
	kingpin.CommandLine.HelpFlag.Hidden()
	kingpin.Parse()

	if prmPause < 30 {
		prmPause = 30
	}

	startTime := getStartTime().Add(time.Duration(prmOffset*-1) * time.Minute)
	goHomeAt := startTime.Add(8 * time.Hour).Add(time.Duration(prmPause) * time.Minute)
	goHomeIn := time.Until(goHomeAt)
	goHomeLatest := startTime.Add(10 * time.Hour).Add(max(45, prmPause) * time.Minute)
	goLatestIn := time.Until(goHomeLatest)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.AlignRight)
	defer w.Flush()

	fmt.Fprintf(w, "started work at\t %s\n", c.Gray(startTime.Format("15:04")))
	fmt.Fprintf(w, "day complete at\t %s (includes %d min. break)\n", c.Bold(c.Cyan(goHomeAt.Format("15:04"))), c.Bold(c.Brown(prmPause)))

	if goHomeIn.Minutes() >= 0 {
		fmt.Fprintf(w, "...that's in\t %.f min\n", c.Bold(c.Green(goHomeIn.Minutes())))
	} else {
		fmt.Fprintf(w, "...that was\t %.f min ago\n", c.Bold(c.Green(goHomeIn.Minutes()*-1)))
	}

	fmt.Fprintf(w, "leave latest at\t %s\n", c.Red(goHomeLatest.Format("15:04")))
	fmt.Fprintf(w, "...that's in\t %.f min\n", c.Red(goLatestIn.Minutes()))
}

func max(a, b int) time.Duration {
	if a > b {
		return time.Duration(a)
	}

	return time.Duration(b)
}
