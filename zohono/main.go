package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"
)

const (
	flagDateFormat = "2006-01-02"
	csvDateFormat  = "02-Jan-2006"
	csvTimeFormat  = "03:04 pm"
)

var (
	start = flag.String("start", "", "start date in YYYY-MM-DD format, defaults to last Monday")
	end   = flag.String("end", "", "start date in YYYY-MM-DD format, defaults to today")
	hours = flag.Int("hours", 8, "number of hours per day")
	job   = flag.String("job", "Work Time", "job name")

	startDay time.Time
	endDay   time.Time
)

type daily struct {
	date      string
	jobName   string
	startTime string
	endTime   string
}

func (d daily) toStringSlice() []string {
	return []string{d.date, d.jobName, d.startTime, d.endTime, strconv.Itoa(*hours)}
}

func main() {
	flag.Parse()

	startDay, endDay, err := parseFlags()
	if err != nil {
		panic(err)
	}

	fmt.Fprintf(os.Stderr, "Start: %s\n", startDay.Format(flagDateFormat))
	fmt.Fprintf(os.Stderr, "End:   %s\n", endDay.Format(flagDateFormat))

	log, err := buildLog(startDay, endDay)
	if err != nil {
		panic(err)
	}

	filename := fmt.Sprintf("%s.%s.csv", startDay.Format(flagDateFormat), endDay.Format(flagDateFormat))
	f, err := os.Create(filename)
	if err != nil {
		panic(fmt.Errorf("could not create filename %q: %v", filename, err))
	}
	defer f.Close()

	if err := writeCSV(f, log); err != nil {
		panic(err)
	}

	fmt.Fprintf(os.Stderr, "Wrote %d days to %s\n", len(log), filename)
}

func parseFlags() (time.Time, time.Time, error) {
	var endDay, startDay time.Time
	var err error

	if *end == "" {
		endDay = time.Now()
	} else {
		endDay, err = time.Parse(flagDateFormat, *end)
		if err != nil {
			return startDay, endDay, fmt.Errorf("invalid --end %q, expected format %q", *end, flagDateFormat)
		}
	}

	if *start == "" {
		// Pick the Monday before (or equal to) endDay.
		// time.Weekday has Sunday at 0.
		daysDiff := (endDay.Weekday() + 6) % 7
		startDay = endDay.AddDate(0, 0, -int(daysDiff))
	} else {
		startDay, err = time.Parse(flagDateFormat, *start)
		if err != nil {
			return startDay, endDay, fmt.Errorf("invalid --start %q, expected format %q", *start, flagDateFormat)
		}
	}

	if startDay.After(endDay) {
		return startDay, endDay, fmt.Errorf("start day %s is after end day %s, that's probably not going to work for you", startDay, endDay)
	}

	midnightStart := time.Date(startDay.Year(), startDay.Month(), startDay.Day(), 0, 0, 0, 0, time.Local)
	midnightEnd := time.Date(endDay.Year(), endDay.Month(), endDay.Day(), 0, 0, 0, 0, time.Local)
	return midnightStart, midnightEnd, nil
}

func buildLog(startDay, endDay time.Time) ([]daily, error) {
	log := []daily{}

	startHour := time.Date(0, 0, 0, 8, 0, 0, 0, time.Local).Format(csvTimeFormat)
	endHour := time.Date(0, 0, 0, 8+(*hours), 0, 0, 0, time.Local).Format(csvTimeFormat)

	// Loop until startDay > endDay (include equal)
	for !startDay.After(endDay) {
		if weekDay := startDay.Weekday(); weekDay != time.Saturday && weekDay != time.Sunday {
			day := daily{
				date:      startDay.Format(csvDateFormat),
				jobName:   *job,
				startTime: startHour,
				endTime:   endHour,
			}
			log = append(log, day)
		}
		startDay = startDay.AddDate(0, 0, 1)
	}

	return log, nil
}

func writeCSV(f io.Writer, log []daily) error {
	w := csv.NewWriter(f)
	err := w.Write([]string{"Date", "Job Name", "From time", "To time", "Hours"})
	if err != nil {
		return fmt.Errorf("could not write header: %v", err)
	}

	records := make([][]string, len(log), len(log))
	for i := 0; i < len(log); i++ {
		records[i] = log[i].toStringSlice()
	}

	err = w.WriteAll(records)
	if err != nil {
		return fmt.Errorf("could not write records: %v", err)
	}

	return nil
}
