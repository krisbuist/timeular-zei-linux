package main

import (
	"time"
	"fmt"
	"strings"
)

type TimeularTime struct {
	time.Time
}

const FORMAT = "2006-01-02T15:04:05.999"

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