package main

import (
	"log"
	"fmt"
)

func main() {
	client := &ApiClient{
		BaseUrl: "https://api.timeular.com/api/v1/",
	}

	state := &Timeular{}

	done := make(chan bool, 1)

	if err := client.authenticate(); err != nil {
		panic(fmt.Sprintf("Could not authenticate. Err: %s\n", err))
	}
	log.Println("API client authenticated")

	if err := client.loadActivities(state); err != nil {
		panic(fmt.Sprintf("Loading activities failed. Err: %s\n", err))
	}
	log.Println("Activities loaded")

	client.run(state)

	manager := ZeiManager{}
	manager.OnOrientationChanged = state.OnOrientationChanged
	manager.run()

	<-done
}
