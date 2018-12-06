package main

import (
	"bufio"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"
)

func getResumeTimeFromJournal() time.Time {
	var startTime time.Time
	var err error
	success := false

	stdout, err := exec.Command("/bin/journalctl", "--since=today", "--no-pager", "-o", "short-iso").Output()
	if err != nil {
		log.Fatal(err)
	}

	journal := fmt.Sprintf("%s", stdout)

	scanner := bufio.NewScanner(strings.NewReader(journal))
	for scanner.Scan() {
		startTime, err = time.Parse("2006-01-02T15:04:05-0700", scanner.Text()[:24])
		if err == nil && isAfter0630(startTime) {
			success = true
			break
		}
	}

	if !success {
		log.Fatal("could not find timestamps in journalctl")
	}

	return startTime
}

func isAfter0630(t time.Time) bool {
	if t.Hour() > 6 ||
		t.Hour() == 6 && t.Minute() >= 30 {
		return true

	}
	return false
}
