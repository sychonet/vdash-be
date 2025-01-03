package controller

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/sychonet/vdash-be/db/entity"
	request "github.com/sychonet/vdash-be/dto/request"
	response "github.com/sychonet/vdash-be/dto/response"
)

// CreateServer creates a new server using the provided request in mongodb database. We want to save id, hostname, publicIP, and libvirtURI for each server.
func (c *ServerController) CreateServer(w http.ResponseWriter, r *http.Request) {
	var req request.CreateServerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Insert the server
	server := entity.ServerInfo{
		ID:         req.ID,
		Hostname:   req.Hostname,
		PublicIP:   req.PublicIP,
		LibvirtURI: "qemu+libssh://root@" + req.PublicIP + "/system",
	}

	err := c.dbService.AddServer(server)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to insert server: %v", err), http.StatusInternalServerError)
		return
	}

	// Prepare the response
	resp := response.CreateServerResponse{
		ID:         req.ID,
		Hostname:   req.Hostname,
		PublicIP:   req.PublicIP,
		LibvirtURI: "qemu+libssh://root@" + req.PublicIP + "/system",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// GetServers returns all servers from mongodb database.
func (c *ServerController) GetServers(w http.ResponseWriter, r *http.Request) {
	serversDetails, err := c.dbService.GetServers()
	if err != nil {
		slog.Error(err.Error())
		http.Error(w, fmt.Sprintf("Failed to get servers: %v", err), http.StatusInternalServerError)
		return
	}

	// Prepare the response
	var servers []response.GetServersResponse
	for _, serverDetail := range serversDetails {
		var serverResponse response.GetServersResponse

		serverResponse.ID = serverDetail.ID
		serverResponse.Hostname = serverDetail.Hostname
		serverResponse.PublicIP = serverDetail.PublicIP
		serverResponse.LibvirtURI = serverDetail.LibvirtURI
		servers = append(servers, serverResponse)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(servers); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// DeleteServer deletes a server from mongodb database.
func (c *ServerController) DeleteServer(w http.ResponseWriter, r *http.Request) {
	var req request.DeleteServerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := c.dbService.DeleteServer(req.ID)

	if err != nil {
		slog.Error(err.Error())
		http.Error(w, fmt.Sprintf("Failed to delete server: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}
