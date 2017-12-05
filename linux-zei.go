package main

import (
	"fmt"
	"log"
	"time"
)


func main() {
	hub := newHub()
	go hub.run()

	go RunWebserver(hub)

	client := &APIClient{
		BaseUrl: "https://api.timeular.com/api/v1/",
	}

	state := &Timeular{}

	if err := client.Authenticate(); err != nil {
		log.Fatalf("Could not authenticate. Err: %s\n", err)
	}
	log.Println("API server authenticated")

	activities, err := client.GetActivities()
	if err != nil {
		log.Fatalf("Loading activities failed. Err: %s\n", err)
	}
	state.Activities = activities
	log.Println("Activities loaded")

	notification := NewNotification()

	manager := BluetoothManager{
		OnOrientationChanged: func(sideID int) {
			log.Printf("Device side: %d", sideID)

			activity := state.GetActivity(sideID)
			current, err := client.GetCurrentTracking()

			if err != nil {
				log.Println("Error: ", err)
				return
			}

			if current != nil && activity != nil && current.Activity.ID == activity.ID {
				return
			}

			if current != nil && ( activity == nil || sideID == 0 ) {
				go notification.Notify("Stopping activity", current.Activity.Name)
				go client.StopActivity(current.Activity)
				return
			}

			if activity != nil && current == nil && sideID != 0 {
				go notification.Notify("Starting activity", activity.Name)
				go client.StartActivity(*activity)
			}

			if current != nil && activity != nil {
				go notification.Notify(
					"Switching activity",
					fmt.Sprintf("%s → %s", current.Activity.Name, activity.Name),
				)
				go func() {
					client.StopActivity(current.Activity)
					client.StartActivity(*activity)
				}()
			}

			state.CurrentSide = sideID
			state.Tracking = &CurrentTracking{
				Activity: *activity,
				StartedAt: TimeularTime{time.Now()},
				Note: "",
			}

			go func() {
				hub.broadcast <- state
			}()
		},
	}

	manager.Run()

	<-make(chan struct{})
}
