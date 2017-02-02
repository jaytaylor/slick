package standup

import (
	"fmt"
	"sort"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
)

type standupData struct {
	Yesterday  string
	Today      string
	Blocking   string
	LastUpdate time.Time
}

func (sd *standupData) Update(section sectionMatch) error {
	log.Infof("sd=%+v", sd)
	switch section.name[0] {
	case 'y':
		sd.Yesterday = section.text
	case 't':
		sd.Today = section.text
	case 'b':
		sd.Blocking = section.text
	default:
		return fmt.Errorf("unrecognized section.name=%q", section.name)
	}
	return nil
}

func (sd standupData) String() string {
	str := fmt.Sprintf("Yesterday: %s\n", sd.Yesterday)
	str += fmt.Sprintf("Today: %s\n", sd.Today)
	str += fmt.Sprintf("Blocking: %s\n", sd.Blocking)
	return str
}

type standupMap map[string]standupUsers

func (sm standupMap) Keys() (standupDates, error) {
	keys := make(standupDates, 0, len(sm))
	for k := range sm {
		sd, err := standupDateFromString(k)
		if err != nil {
			return nil, err
		}
		keys = append(keys, sd)
	}
	return keys, nil
}

// Filter returns a copy of standupMap filtered by user fields [Name || Email]
func (sm standupMap) filterByEmail(email string) standupMap {
	fsm := make(standupMap)
	for date, users := range sm {
		fsm[date] = users.filterByEmail(email)
	}
	return fsm
}

func lineBreak(nchars int) string {
	line := ""
	for i := 0; i < nchars; i++ {
		line += "="
	}
	return line
}

/* String stringifies the map such that it will print:
Standup Report
 DATE
=====
username1
Yesterday: blah
Today: blah blah
Blocking: blah blah blah

username2
Yesterday: blah
Today: blah blah
Blocking: blah blah blah

DATE
====

...

Unless there is only 1 user in the map, in which case it does not repeat
the username and will look like

Standup Report for username
 DATE
=====
Yesterday: blah
Today: blah blah
Blocking: blah blah blah

DATE
====
Yesterday: blah
Today: blah blah
Blocking: blah blah blah

DATE
====
....

*/
func (sm standupMap) String() (string, error) {
	sorted, err := sm.Keys()
	if err != nil {
		return "", err
	}
	sort.Sort(sorted)

	var str string

	// first pass detects if there are multiple users or a single user (email is used as unique ID)
	seenUsers := make(map[string]standupUser)
	var lastUser standupUser
	singleUserReport := false

	for _, sdate := range sorted {
		users := sm[sdate.String()]
		for _, user := range users {
			seenUsers[user.Profile.Email] = user
			lastUser = user
		}
	}

	// write header depending on single or multiple user case
	if len(seenUsers) == 1 {
		singleUserReport = true
		str += fmt.Sprintf("Standup Report for %s\n", lastUser.Name)
	} else {
		str += "Standup Report\n"
	}

	// second pass stringifies the body and only prints user name if multiple users exist
	for _, sdate := range sorted {
		users := sm[sdate.String()]
		str += fmt.Sprintf("%s\n", sdate.String())
		str += fmt.Sprintf("%s\n", lineBreak(len(sdate.String())))
		for _, user := range users {
			// if only single user don't repeatedly write name
			if !singleUserReport {
				str += fmt.Sprintf("%s\n", user.Name)
			}
			str += fmt.Sprintf("%s\n", user.Data.String())
		}
	}

	// replace multiple newlines with a single newline at the end.
	str = strings.TrimRight(str, "\n") + "\n"
	return str, nil
}
