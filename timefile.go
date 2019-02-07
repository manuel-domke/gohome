package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/danielb42/goat"
	c "github.com/logrusorgru/aurora"
	yaml "gopkg.in/yaml.v2"
)

var p *persistentData

type timestruct struct {
	TimefilePath           string        `yaml:"-"`
	Pause                  int           `yaml:"Pause"`
	StartTime              time.Time     `yaml:"StartTime"`
	GoHomeAt, GoHomeLatest time.Time     `yaml:"-"`
	GoHomeIn, GoLatestIn   time.Duration `yaml:"-"`
	AtJobID                int           `yaml:"AtJobID"`
}

type persistentData struct {
	DailyHours int `yaml:"DailyHours"`
}

func newTimestruct() *timestruct {
	p = new(persistentData)
	p.readPersistentFile()

	ts := new(timestruct)
	ts.TimefilePath = os.Getenv("HOME") + "/.gohome"
	return ts
}

func (t *timestruct) setStartTime(setTime string) {
	var err error
	t.StartTime, err = time.Parse("15:04", setTime)
	if err != nil {
		log.Fatalf("could not parse supplied start time: %s", setTime)
	}

	now := time.Now()
	t.StartTime = time.Date(
		now.Year(), now.Month(), now.Day(),
		t.StartTime.Hour(), t.StartTime.Minute(),
		0, 0, time.Local,
	)
}

func (t *timestruct) setPause(pause int) {
	if pause < 30 {
		t.Pause = 30
	} else {
		t.Pause = pause
	}
}

func (t *timestruct) store() {
	yaml, err := yaml.Marshal(&t)
	if err != nil {
		log.Fatal(err)
	}

	timefile, err := os.Create(t.TimefilePath)
	if err != nil {
		log.Fatal(err)
	}

	_, err = fmt.Fprintln(timefile, string(yaml))
	if err != nil {
		log.Fatal(err)
	}
}

func (t *timestruct) read() {
	timefileContent, err := ioutil.ReadFile(t.TimefilePath)
	if err != nil {
		log.Fatal(err)
	}

	if err = yaml.Unmarshal(timefileContent, &t); err != nil {
		log.Fatal(err)
	}
}

func (t *timestruct) timeFileisOfToday() bool {
	stat, err := os.Stat(t.TimefilePath)
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

func (t *timestruct) remove() {
	err := os.Remove(t.TimefilePath)

	if err != nil && !os.IsNotExist(err) {
		log.Fatal(err)
	}

	os.Exit(0)
}

func (t *timestruct) calculate() {
	t.GoHomeAt = t.StartTime.Add(time.Duration(p.DailyHours) * time.Hour).Add(time.Duration(t.Pause) * time.Minute)
	t.GoHomeLatest = t.StartTime.Add(10 * time.Hour).Add(longer(45, t.Pause) * time.Minute)

	if t.GoHomeLatest.Hour() >= 21 || t.GoHomeLatest.Day() != t.StartTime.Day() {
		t.GoHomeLatest = time.Date(
			t.StartTime.Year(), t.StartTime.Month(), t.StartTime.Day(),
			21, 0, 0, 0, time.Local,
		)
	}

	t.GoHomeIn = time.Until(t.GoHomeAt)
	t.GoLatestIn = time.Until(t.GoHomeLatest)

	goat.RemoveJob(t.AtJobID)
	t.AtJobID, _ = goat.AddJob("notify-send -i error 'Go home!'", t.GoHomeAt)
}

func (t *timestruct) print() {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.AlignRight)

	fmt.Fprintf(w, "started work at\t %s\n\n", c.Bold(c.Gray(t.StartTime.Format("15:04"))))

	if t.GoHomeAt.Hour() >= 21 || t.GoHomeAt.Day() != t.StartTime.Day() {
		fmt.Fprintf(w, "%d-hour day complete at\t %s (includes %d min. break) %s\n",
			p.DailyHours,
			c.Bold(c.Red(t.GoHomeAt.Format("15:04"))),
			c.Brown(t.Pause),
			c.Bold(c.Red("after cut off!")),
		)
	} else {
		fmt.Fprintf(w, "%d-hour day complete at\t %s (includes %d min. break)\n",
			p.DailyHours,
			c.Bold(c.Cyan(t.GoHomeAt.Format("15:04"))),
			c.Brown(t.Pause),
		)
	}

	if t.GoHomeIn.Minutes() >= 0 {
		fmt.Fprintf(w, "...that's in\t %s\n", c.Bold(c.Cyan(printDuration(t.GoHomeIn))))
	} else {
		fmt.Fprintf(w, "...that was\t %s ago\n", c.Bold(c.Green(printDuration(t.GoHomeIn))))
	}

	fmt.Fprintf(w, "\nleave latest at\t %s\n", c.Red(t.GoHomeLatest.Format("15:04")))

	if t.GoLatestIn.Minutes() >= 0 {
		fmt.Fprintf(w, "...that's in\t %s\n", c.Red(printDuration(t.GoLatestIn)))
	} else {
		fmt.Fprintf(w, "...that was\t %s ago\n", c.Bold(c.Red(printDuration(t.GoLatestIn))))
	}

	if err := w.Flush(); err != nil {
		log.Fatal(err)
	}
}

func (p *persistentData) readPersistentFile() {
	persistentFileContent, err := ioutil.ReadFile(os.Getenv("HOME") + "/.gohome_persistent")
	if err != nil {
		if os.IsNotExist(err) {
			p.DailyHours = 8
			return
		}

		log.Fatal(err)
	}

	if err = yaml.Unmarshal(persistentFileContent, &p); err != nil {
		log.Fatal(err)
	}
}

func longer(a int, b int) time.Duration {
	if a > b {
		return time.Duration(a)
	}

	return time.Duration(b)
}

func printDuration(dur time.Duration) string {
	h := int(dur.Hours())
	m := int(dur.Minutes()) - 60*h
	return strings.Replace(fmt.Sprintf("%dh%02dm", h, m), "-", "", -1)
}
