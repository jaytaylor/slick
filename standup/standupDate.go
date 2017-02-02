package standup

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type standupDate struct {
	year  int
	month time.Month
	day   int
}

func standupDateFromString(src string) (standupDate, error) {
	pieces := strings.Split(src, "-")
	if len(pieces) != 3 {
		return standupDate{}, fmt.Errorf("invalid standupDate string=%q", src)
	}
	year, err := strconv.Atoi(pieces[0])
	if err != nil {
		return standupDate{}, fmt.Errorf("invalid year=%q", pieces[0])
	}
	month, err := strconv.Atoi(pieces[1])
	if err != nil {
		return standupDate{}, fmt.Errorf("invalid month=%q", pieces[1])
	}
	day, err := strconv.Atoi(pieces[2])
	if err != nil {
		return standupDate{}, fmt.Errorf("invalid day=%q", pieces[2])
	}
	sd := standupDate{
		year:  year,
		month: time.Month(month),
		day:   day,
	}
	return sd, nil
}

func getStandupDate(daysFromToday int) standupDate {
	d := time.Now().Add(time.Duration(daysFromToday) * 24 * time.Hour)
	return standupDate{
		year:  d.Year(),
		month: d.Month(),
		day:   d.Day(),
	}
}

func unixToStandupDate(unix int64) standupDate {
	d := time.Unix(unix, 0).UTC()
	return standupDate{
		year:  d.Year(),
		month: d.Month(),
		day:   d.Day(),
	}
}

func (sd standupDate) next() standupDate {
	current := time.Date(sd.year, sd.month, sd.day, 0, 0, 0, 0, time.Local)
	next := current.Add(24 * time.Hour)
	return standupDate{
		year:  next.Year(),
		month: next.Month(),
		day:   next.Day(),
	}
}

func (sd standupDate) String() string {
	s := fmt.Sprintf("%04v-%02v-%02v", sd.year, int(sd.month), sd.day)
	return s
}

func (sd standupDate) Unix() int64 {
	return time.Date(sd.year, sd.month, sd.day, 0, 0, 0, 0, time.Local).Unix()
}

func (sd standupDate) toUnixUTCString() string {
	return strconv.FormatInt(sd.Unix(), 10)
}

type standupDates []standupDate

func (slice standupDates) Len() int {
	return len(slice)
}

func (slice standupDates) Less(i, j int) bool {
	return slice[i].Unix() < slice[j].Unix()
}

func (slice standupDates) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}
