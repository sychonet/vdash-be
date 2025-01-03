package service

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/sychonet/vdash-be/dto/response"
)

type ScalewayService struct {
	BaseURL string
	Token   string
}

func NewScalewayService(baseURL, token string) *ScalewayService {
	return &ScalewayService{
		BaseURL: baseURL,
		Token:   token,
	}
}

// Call API endpoint GET https://api.online.net/api/v1/server to fetch all servers running under admin account on Scaleway
func (s *ScalewayService) GetServers() ([]string, error) {
	// create a new http client
	client := &http.Client{Timeout: 5 * time.Second}

	// create a new http request
	req, err := http.NewRequest("GET", s.BaseURL+"/api/v1/server", nil)
	if err != nil {
		return nil, err
	}

	// set the bearer token in Authorization header
	req.Header.Add("Authorization", s.Token)
	// send request to the api server for service provider Scaleway
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	// parse the response
	var servers []string
	if err := json.NewDecoder(resp.Body).Decode(&servers); err != nil {
		return nil, err
	}

	// return the list of servers
	return servers, nil
}

// Call API endpoint GET https://api.online.net/api/v1/server/{server_id} to fetch server details
func (s *ScalewayService) GetServerDetails(serverString string) (*response.ScalewayServerResponse, error) {
	// create a new http client
	client := &http.Client{Timeout: 5 * time.Second}

	// create a new http request
	req, err := http.NewRequest("GET", s.BaseURL+serverString, nil)
	if err != nil {
		return nil, err
	}

	// set the bearer token in Authorization header
	req.Header.Add("Authorization", s.Token)
	// send request to the api server for service provider Scaleway
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	// parse the response
	var server response.ScalewayServerResponse
	if err := json.NewDecoder(resp.Body).Decode(&server); err != nil {
		return nil, err
	}

	// return the server details
	return &server, nil
}
