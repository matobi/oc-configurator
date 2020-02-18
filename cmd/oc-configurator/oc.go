package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

func (s *server) GetConfig(url string) (string, error) {
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", s.ocToken))
	request.Header.Add("Impersonate-User", "system:admin")
	//log.Printf("request:\n%+v\n", request)

	response, err := s.ocClient.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		errBody, _ := ioutil.ReadAll(response.Body)
		log.Printf("error body: %s\n", errBody)
		return "", fmt.Errorf("http status error=%d", response.StatusCode)
	}
	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func (s *server) PatchConfig(url string, body string) (string, error) {
	request, err := http.NewRequest(http.MethodPatch, url, strings.NewReader(body))
	if err != nil {
		return "", err
	}
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", s.ocToken))
	request.Header.Add("Impersonate-User", "system:admin")
	request.Header.Add("Content-Type", "application/merge-patch+json")

	response, err := s.ocClient.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		errBody, _ := ioutil.ReadAll(response.Body)
		log.Printf("error body: %s\n", errBody)
		return "", fmt.Errorf("http status error=%d", response.StatusCode)
	}
	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	return string(content), nil
}
