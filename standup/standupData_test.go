package standup

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/nlopes/slack"
)

func TestStandupDataString(t *testing.T) {

	datastr := standupData{
		Yesterday: "a",
		Today:     "b",
		Blocking:  "c",
	}.String()

	str := `Yesterday: a
Today: b
Blocking: c
`

	if datastr != str {
		t.Error("expected '" + datastr + "'" + " to be '" + str + "'")
	}
}

func getTestStandupMap() standupMap {

	sm := make(standupMap)

	uA := standupUser{
		&slack.User{
			Name:    "A",
			Profile: slack.UserProfile{Email: "A@test.ly"},
		},
		standupData{},
	}

	uB := standupUser{
		&slack.User{
			Name:    "B",
			Profile: slack.UserProfile{Email: "B@test.ly"},
		},
		standupData{},
	}

	var unixDate int64 = 1431921600 // 2015-05-18
	sds := []standupDate{
		unixToStandupDate(unixDate),
		unixToStandupDate(unixDate).next(),
	}

	for i := 0; i < 2; i += 1 {
		uA.Data.Yesterday = strconv.Itoa(i)
		uA.Data.Today = strconv.Itoa(i)
		uA.Data.Blocking = strconv.Itoa(i)
		uB.Data.Yesterday = strconv.Itoa(i)
		uB.Data.Today = strconv.Itoa(i)
		uB.Data.Blocking = strconv.Itoa(i)

		sm[sds[i].String()] = standupUsers{uA, uB}
	}

	return sm
}

func TestSingleUserMapString(t *testing.T) {

	sm := getTestStandupMap().filterByEmail("B@test.ly")

	str, err := sm.String()
	if err != nil {
		t.Fatal(err)
	}
	singleReport := fmt.Sprintf("%s", str)
	expectedReport := `Standup Report for B
2015-05-18
==========
Yesterday: 0
Today: 0
Blocking: 0

2015-05-19
==========
Yesterday: 1
Today: 1
Blocking: 1
`

	if singleReport != expectedReport {
		t.Error("Expected '" + singleReport + "' to be '" + expectedReport + "'")
	}

}

func TestMultipleUserMapString(t *testing.T) {

	sm := getTestStandupMap()

	str, err := sm.String()
	if err != nil {
		t.Fatal(err)
	}
	multiReport := fmt.Sprintf("%s", str)
	expectedReport := `Standup Report
2015-05-18
==========
A
Yesterday: 0
Today: 0
Blocking: 0

B
Yesterday: 0
Today: 0
Blocking: 0

2015-05-19
==========
A
Yesterday: 1
Today: 1
Blocking: 1

B
Yesterday: 1
Today: 1
Blocking: 1
`

	if multiReport != expectedReport {
		t.Error("Expected '" + multiReport + "' to be '" + expectedReport + "'")
	}

}
