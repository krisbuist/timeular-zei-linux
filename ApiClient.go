package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type APIClient struct {
	BaseUrl string
	Token   string
}

type AuthenticationRequest struct {
	ApiKey    string `json:"apiKey"`
	ApiSecret string `json:"apiSecret"`
}

type AuthenticationResponse struct {
	Token string `json:"token"`
}

type ActivateDeviceParam struct {
	Serial string `json:"deviceSerial"`
}

type ActivateDeviceResponse struct {
	Serial   string `json:"serial"`
	Name     string `json:"name"`
	Active   bool   `json:"active"`
	Disabled bool   `json:"disabled"`
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

type ErrorResponse struct {
	Timestamp int    `json:"timestamp"`
	Status    int    `json:"status"`
	Error     string `json:"error"`
	Exception string `json:"exception"`
	Message   string `json:"message"`
	Path      string `json:"path"`
}

func (client *APIClient) Authenticate() error {
	config, err := ioutil.ReadFile("./config.json")
	if err != nil {
		return err
	}
	request := &AuthenticationRequest{}
	json.Unmarshal(config, request)

	response := &AuthenticationResponse{}

	if err := client.doPost("/developer/sign-in", request, response); err != nil {
		return err
	}

	client.Token = "Bearer " + response.Token

	return nil
}

func (client *APIClient) ActivateDevice() error {
	config, err := ioutil.ReadFile("./config.json")
	if err != nil {
		return err
	}
	param := &ActivateDeviceParam{}
	json.Unmarshal(config, param)

	if param.Serial == "" {
		log.Println("Not activating any specific device")
		return nil
	}

	response := &ActivateDeviceResponse{}

	if err := client.doPost(fmt.Sprintf("/devices/%s/active", param.Serial), nil, response); err != nil {
		return err
	}

	log.Printf("Device %s activated", param.Serial)
	return nil
}

func (client *APIClient) GetCurrentTracking() (*CurrentTracking, error) {

	response := &CurrentTrackingResponse{}

	if err := client.doGet("/tracking", response); err != nil {
		return nil, err
	}

	return response.CurrentTracking, nil
}

func (client *APIClient) GetActivities() ([]Activity, error) {

	response := &ActivitiesResponse{}

	if err := client.doGet("/activities", response); err != nil {
		return []Activity{}, err
	}

	return response.Activities, nil
}

func (client *APIClient) StartActivity(a Activity) *CurrentTracking {

	requestBody := &StartActivityRequest{
		StartedAt: TimeularTime{time.Now()},
	}

	response := &StartActivityResponse{}

	if err := client.doPost(fmt.Sprintf("/tracking/%s/start", a.ID), requestBody, response); err != nil {
		return nil
	}

	return &response.CurrentTracking
}

func (client *APIClient) StopActivity(a Activity) {

	requestBody := &StopActivityRequest{
		StoppedAt: TimeularTime{time.Now()},
	}

	response := &StopActivityResponse{}

	if err := client.doPost(fmt.Sprintf("/tracking/%s/stop", a.ID), requestBody, response); err != nil {
		return
	}
}

func (client *APIClient) doPost(path string, requestObject interface{}, responseObject interface{}) error {
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
		response := ErrorResponse{}
		err = json.Unmarshal(body, &response)
		log.Println()
		log.Println("========== POST call failed ==========")
		log.Printf("Path:        %s\n", response.Path)
		log.Printf("Status Code: %d\n", response.Status)
		log.Printf("Error:       %s\n", response.Error)
		log.Printf("Message:     %s\n", response.Message)
		log.Println()
		return errors.New(response.Message)
	}

	return json.Unmarshal(body, responseObject)
}

func (client *APIClient) doGet(path string, response interface{}) error {
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
		response := ErrorResponse{}
		err = json.Unmarshal(body, &response)
		log.Println()
		log.Println("========== GET call failed ==========")
		log.Printf("Path:        %s\n", response.Path)
		log.Printf("Status Code: %d\n", response.Status)
		log.Printf("Error:       %s\n", response.Error)
		log.Printf("Message:     %s\n", response.Message)
		log.Println()
		return errors.New(response.Message)
	}

	return json.Unmarshal(body, &response)
}
