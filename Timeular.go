package main

import (
	"fmt"
	"strings"
	"time"
)

type Activity struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Color       string `json:"color"`
	Integration string `json:"integration"`
	DeviceSide  *int   `json:"deviceSide"`
}

type CurrentTracking struct {
	Activity  *Activity    `json:"activity"`
	StartedAt TimeularTime `json:"startedAt"`
	Note      string       `json:"note"`
}

type TimeEntry struct {
	ID       string   `json:"id"`
	Activity Activity `json:"activity"`
	Note     string   `json:"note"`
}

type Timeular struct {
	CurrentSide int
	Tracking    *CurrentTracking
	Activities  []Activity
}

func (timeular *Timeular) GetActivity(deviceSide int) *Activity {
	for _, a := range timeular.Activities {
		if a.DeviceSide != nil && *a.DeviceSide == deviceSide {
			return &a
		}
	}

	return nil
}

type TimeularTime struct {
	time.Time
}

const FORMAT = "2006-01-02T15:04:05.000"

func (t *TimeularTime) MarshalJSON() ([]byte, error) {
	loc, _ := time.LoadLocation("UTC")
	stamp := fmt.Sprintf("\"%s\"", time.Time(t.Time).In(loc).Format(FORMAT))
	return []byte(stamp), nil
}

func (t *TimeularTime) UnmarshalJSON(data []byte) error {

	value := strings.Trim(string(data), "\"")

	if value == "null" {
		t.Time = time.Time{}
		return nil
	}

	t.Time, _ = time.Parse(FORMAT, value)

	return nil
}
