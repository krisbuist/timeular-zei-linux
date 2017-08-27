package main

import (
	"net/http"
	"encoding/json"
	"bytes"
	"io/ioutil"
	"time"
	"fmt"
	"errors"
	"log"
	"github.com/0xAX/notificator"
)

type ApiClient struct {
	BaseUrl  string
	Token    string
	Notifier *notificator.Notificator
	Timeular *Timeular
}

type AuthorizationRequest struct {
	ApiKey    string `json:"apiKey"`
	ApiSecret string `json:"apiSecret"`
}

type AuthorizationResponse struct {
	Token string `json:"token"`
}

type ActivitiesResponse struct {
	Activities []Activity `json:"activities"`
}

type CurrentTrackingResponse struct {
	CurrentTracking *CurrentTracking `json:"currentTracking"`
}

type StartActivityRequest struct {
	StartedAt TimeularTime `json:"startedAt"`
}

type StartActivityResponse struct {
	CurrentTracking CurrentTracking `json:"currentTracking"`
}

type StopActivityRequest struct {
	StoppedAt TimeularTime `json:"stoppedAt"`
}

type StopActivityResponse struct {
	CreatedTimeEntry TimeEntry `json:"createdTimeEntry"`
}

func (client *ApiClient) run(t *Timeular) {
	client.Notifier = notificator.New(notificator.Options{
		DefaultIcon: "/home/kris/Desktop/timeular.png",
		AppName:     "ZEI",
	})

	client.Timeular = t

	t.OnOrientationChanged = client.onActivityChange
}

func (client *ApiClient) authenticate() error {
	authString, _ := ioutil.ReadFile("config.json")
	request := &AuthorizationRequest{}
	json.Unmarshal(authString, request)

	response := &AuthorizationResponse{}

	if err := client.doPost("/developer/sign-in", request, response); err != nil {
		return err
	}

	client.Token = "Bearer " + response.Token

	return nil
}

func (client *ApiClient) loadCurrentTracking(t *Timeular) error {

	response := &CurrentTrackingResponse{}

	if err := client.doGet("/tracking", response); err != nil {
		return err
	}

	t.CurrentTracking = response.CurrentTracking

	return nil
}

func (client *ApiClient) loadActivities(t *Timeular) error {

	response := &ActivitiesResponse{}

	if err := client.doGet("/activities", response); err != nil {
		return err
	}

	t.Activities = response.Activities

	return nil
}

func (client *ApiClient) onActivityChange(sideID int) {
	log.Printf("Device side: %d", sideID)

	matched := client.Timeular.getActivity(sideID)
	client.loadCurrentTracking(client.Timeular)
	current := client.Timeular.CurrentTracking

	if current != nil && matched != nil && current.Activity.ID == matched.ID {
		return
	}

	if current != nil {
		client.stopActivity(current.Activity)
		client.Timeular.CurrentTracking = nil
	}

	if matched != nil {
		client.Timeular.CurrentTracking = client.startActivity(*matched)
	}
}

func (client *ApiClient) startActivity(a Activity) *CurrentTracking {
	client.notify(fmt.Sprintf("Starting activity %s", a.Name))

	requestBody := &StartActivityRequest{
		StartedAt: TimeularTime{time.Now()},
	}

	response := &StartActivityResponse{}

	if err := client.doPost(fmt.Sprintf("/tracking/%s/start", a.ID), requestBody, response); err != nil {
		log.Println("Error: ", err)
		return nil
	}

	return &response.CurrentTracking
}

func (client *ApiClient) stopActivity(a Activity) {
	client.notify(fmt.Sprintf("Stopping activity: %s", a.Name))

	requestBody := &StopActivityRequest{
		StoppedAt: TimeularTime{time.Now()},
	}

	response := &StopActivityResponse{}

	if err := client.doPost(fmt.Sprintf("/tracking/%s/stop", a.ID), requestBody, response); err != nil {
		log.Println("Error: ", err)
		return
	}
}

func (client *ApiClient) notify(message string) {
	log.Println(message)
	client.Notifier.Push("Stopping activity", message, "", notificator.UR_NORMAL)
}

func (client *ApiClient) doPost(path string, requestObject interface{}, responseObject interface{}) error {
	requestBody, _ := json.Marshal(requestObject)
	request, _ := http.NewRequest("POST", client.BaseUrl+path, bytes.NewBuffer(requestBody))
	request.Header.Set("Authorization", client.Token)
	request.Header.Set("Accept", "application/json;charset:UTF-8")
	request.Header.Set("Content-Type", "application/json;charset:UTF-8")

	res, err := (&http.Client{}).Do(request)
	defer res.Body.Close()

	body, _ := ioutil.ReadAll(res.Body)

	if err != nil {
		log.Println("Request failed. Err: ", err)
		return err
	}

	if res.StatusCode != 200 {
		message := fmt.Sprintf("API call failed. Status code: %d. Body: %s", res.StatusCode, body)
		return errors.New(message)
	}

	return json.Unmarshal(body, responseObject)
}

func (client *ApiClient) doGet(path string, response interface{}) error {
	req, _ := http.NewRequest("GET", client.BaseUrl+path, nil)
	req.Header.Set("Accept", "application/json;charset:UTF-8")
	req.Header.Set("Authorization", client.Token)

	res, err := (&http.Client{}).Do(req)
	defer res.Body.Close()

	if err != nil {
		return err
	}

	body, _ := ioutil.ReadAll(res.Body)

	if res.StatusCode != 200 {
		message := fmt.Sprintf("API call failed. Status code: %d. Body: %s", res.StatusCode, body)
		return errors.New(message)
	}

	return json.Unmarshal(body, &response)
}
