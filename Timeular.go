package main

type Activity struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Color       string `json:"color"`
	Integration string `json:"integration"`
	DeviceSide  int    `json:"deviceSide"`
}

type CurrentTracking struct {
	Activity  Activity     `json:"activity"`
	StartedAt TimeularTime `json:"startedAt"`
	Note      string       `json:"note"`
}

type TimeEntry struct {
	ID       string   `json:"id"`
	Activity Activity `json:"activity"`
	Note     string   `json:"note"`
}

type Timeular struct {
	Activities           []Activity
	CurrentTracking      *CurrentTracking
	TimeEntries          []TimeEntry
	OnOrientationChanged func(side int)
}

func (timeular *Timeular) getActivity(deviceSide int) *Activity {
	for _, a := range timeular.Activities {
		if a.DeviceSide == deviceSide {
			return &a
		}
	}

	return nil
}
