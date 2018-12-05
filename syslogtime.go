package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

func getEarliestSyslogToday(syslogPaths ...string) time.Time {
	var startTime, earliestHere time.Time
	found := false

	for _, syslogPath := range syslogPaths {
		syslog, err := os.Open(syslogPath)
		if err != nil {
			continue
		}

		scanner := bufio.NewScanner(syslog)

		for scanner.Scan() {
			if strings.Contains(scanner.Text(), time.Now().Format("Jan 2")) {
				found = true
				fmt.Printf("found %s\n", scanner.Text())
				dateTokens := strings.SplitN(scanner.Text(), ":", 3)
				dateString := strings.Join(dateTokens[:2], " ")
				earliestHere, err = time.Parse("Jan 2 15 04", dateString)
				if err != nil {
					log.Fatalf("could not parse timestamp: %s\n", dateString)
				}

				if earliestHere.Hour() < 6 {
					continue
				}

				earliestHere = time.Date(
					time.Now().Year(), earliestHere.Month(), earliestHere.Day(),
					earliestHere.Hour(), earliestHere.Minute(), 0, 0, time.Local,
				)

				break
			}
		}

		if startTime.Before(earliestHere) {
			startTime = earliestHere
		}
	}

	if !found {
		log.Fatal("did not find any syslog entries of today")
	}

	return startTime
}
