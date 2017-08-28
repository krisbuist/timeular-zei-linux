package main

import (
	"fmt"
	"log"
	"github.com/krisbuist/timeular-zei-linux/API"
	"github.com/krisbuist/timeular-zei-linux/BlueTooth"
	"github.com/krisbuist/timeular-zei-linux/Notification"
)

func main() {
	client := &API.Client{
		BaseUrl: "https://api.timeular.com/api/v1/",
	}

	state := &API.Timeular{}

	if err := client.Authenticate(); err != nil {
		log.Fatalf("Could not authenticate. Err: %s\n", err)
	}
	log.Println("API client authenticated")

	activities, err := client.GetActivities()
	if err != nil {
		log.Fatalf("Loading activities failed. Err: %s\n", err)
	}
	state.Activities = activities
	log.Println("Activities loaded")

	notification := Notification.NewDesktop()

	manager := BlueTooth.ZeiManager{
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

			if current != nil && activity == nil {
				go notification.Notify("Stopping activity", current.Activity.Name)
				go client.StopActivity(current.Activity)
			}

			if activity != nil && current == nil {
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
		},
	}

	manager.Run()

	<-make(chan struct{})
}
