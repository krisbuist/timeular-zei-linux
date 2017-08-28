package main

import (
	"fmt"
	"log"
	"timeular-zei-linux/API"
	"timeular-zei-linux/BlueTooth"
	"timeular-zei-linux/Notification"
)

func main() {
	client := &API.Client{
		BaseUrl: "https://api.timeular.com/api/v1/",
	}

	state := &API.Timeular{}

	if err := client.Authenticate(); err != nil {
		panic(fmt.Sprintf("Could not authenticate. Err: %s\n", err))
	}
	log.Println("API client authenticated")

	activities, err := client.GetActivities()
	if err != nil {
		panic(fmt.Sprintf("Loading activities failed. Err: %s\n", err))
	}
	state.Activities = activities
	log.Println("Activities loaded")

	manager := BlueTooth.ZeiManager{}

	notification := Notification.NewDesktop()

	manager.OnOrientationChanged = func(sideID int) {
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

		if current != nil {
			go notification.Notify(fmt.Sprintf("Stopping activity: %s", current.Activity.Name))
			client.StopActivity(current.Activity)
		}

		if activity != nil {
			go notification.Notify(fmt.Sprintf("Starting activity: %s", activity.Name))
			client.StartActivity(*activity)
		}
	}

	manager.Run()

	<-make(chan struct{})
}
