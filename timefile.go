package main

import (
	"database/sql"
	"log"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var (
	db *sql.DB
)

type timefile struct {
	path      string
	startTime time.Time
	pause     int
	offset    int
}

func newTimefile(path string) *timefile {
	var err error
	db, err = sql.Open("sqlite3", path)
	if err != nil {
		log.Fatal(err)
	}

	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		stmt, err := db.Prepare(`
		CREATE TABLE IF NOT EXISTS times (
			ID INT,
			STARTTIME TEXT, 
			PAUSE INT, 
			OFFSET INT)
		`)
		if err != nil {
			log.Fatal(err)
		}

		_, err = stmt.Exec()
		if err != nil {
			log.Fatal(err)
		}

		stmt, err = db.Prepare("INSERT INTO times (ID) VALUES (1)")
		if err != nil {
			log.Fatal(err)
		}

		_, err = stmt.Exec()
		if err != nil {
			log.Fatal(err)
		}
	}

	timeFile := new(timefile)
	timeFile.path = path
	return timeFile
}

func (t *timefile) setStartTime(setTime string) {
	st, err := time.Parse("15:04", setTime)
	if err != nil {
		log.Fatalf("could not parse supplied start time: %s", setTime)
	}

	t.startTime = insertInToday(st)

	stmt, err := db.Prepare("UPDATE times SET STARTTIME = ? WHERE ID=1")
	if err != nil {
		log.Fatal(err)
	}

	_, err = stmt.Exec(setTime)
	if err != nil {
		log.Fatal(err)
	}
}

func (t *timefile) setPause(pause int) {
	t.pause = pause

	stmt, err := db.Prepare("UPDATE times SET PAUSE = ? WHERE ID=1")
	if err != nil {
		log.Fatal(err)
	}

	_, err = stmt.Exec(pause)
	if err != nil {
		log.Fatal(err)
	}
}

func (t *timefile) setOffset(offset int) {
	t.offset = offset

	stmt, err := db.Prepare("UPDATE times SET OFFSET = ? WHERE ID=1")
	if err != nil {
		log.Fatal(err)
	}

	_, err = stmt.Exec(offset)
	if err != nil {
		log.Fatal(err)
	}
}

func (t *timefile) read() {
	row, err := db.Query("SELECT STARTTIME,PAUSE,OFFSET FROM times")
	if err != nil {
		log.Fatal(err)
	}
	defer row.Close()

	var s string
	for row.Next() {
		err := row.Scan(&s, &t.pause, &t.offset)
		if err != nil {
			log.Fatal(err)
		}
	}

	foobar, _ := time.Parse("15:04", s)
	t.startTime = insertInToday(foobar)
	err = row.Err()
	if err != nil {
		log.Fatal(err)
	}
}

func (t *timefile) isOfToday() bool {
	stat, err := os.Stat(t.path)
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

func (t *timefile) remove() {
	err := os.Remove(t.path)

	if err != nil && !os.IsNotExist(err) {
		log.Fatal(err)
	}

	os.Exit(0)
}

func insertInToday(hourAndMin time.Time) time.Time {
	now := time.Now()
	return time.Date(
		now.Year(), now.Month(), now.Day(),
		hourAndMin.Hour(), hourAndMin.Minute(),
		0, 0, time.Local,
	)
}

func (t *timefile) buildTimeStruct() *timestruct {
	target := new(timestruct)

	target.goHomeAt = t.startTime.Add(8 * time.Hour).Add(time.Duration(t.pause) * time.Minute)
	target.goHomeLatest = t.startTime.Add(10 * time.Hour).Add(longer(45, t.pause) * time.Minute)

	target.goHomeIn = time.Until(target.goHomeAt)
	target.goLatestIn = time.Until(target.goHomeLatest)

	return target
}
