package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	request "github.com/sychonet/vdash-be/dto/request"
	response "github.com/sychonet/vdash-be/dto/response"
)

// CreateNetwork creates a new network on a server using the provided request.
func (c *ServerController) CreateNetwork(w http.ResponseWriter, r *http.Request) {
	var req request.NetworkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate the request
	if req.Name == "" {
		http.Error(w, "Invalid name", http.StatusBadRequest)
		return
	}

	if req.Bridge == "" {
		http.Error(w, "Invalid bridge", http.StatusBadRequest)
		return
	}

	if req.ServerID <= 0 {
		http.Error(w, "Invalid serverID", http.StatusBadRequest)
		return
	}

	// Get the libvirt URI from the database
	serverDetail, err := c.dbService.GetServer(req.ServerID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get server details: %v", err), http.StatusInternalServerError)
		return
	}

	c.libvirtService.URI = serverDetail.LibvirtURI

	err = c.libvirtService.CreateNetwork(req.Name, req.Bridge)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create network: %v", err), http.StatusInternalServerError)
		return
	}

	// Prepare the response
	resp := response.NetworkResponse{
		Name:   req.Name,
		Bridge: req.Bridge,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// GetNetworks returns a list of all networks on a server.
func (c *ServerController) GetNetworks(w http.ResponseWriter, r *http.Request) {
	serverIDParam := r.URL.Query().Get("serverID")

	if serverIDParam == "" {
		http.Error(w, "Missing serverID query parameter", http.StatusBadRequest)
		return
	}

	serverID, err := strconv.Atoi(serverIDParam)
	if err != nil {
		http.Error(w, "Invalid serverID query parameter", http.StatusBadRequest)
		return
	}

	// Get the libvirt URI from the database
	serverDetail, err := c.dbService.GetServer(serverID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get server details: %v", err), http.StatusInternalServerError)
		return
	}

	c.libvirtService.URI = serverDetail.LibvirtURI

	networks, err := c.libvirtService.GetNetworks()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list networks: %v", err), http.StatusInternalServerError)
		return
	}

	defer func() {
		for _, network := range networks {
			network.Free()
		}
	}()

	// Prepare the response
	var resp []response.NetworkResponse
	for _, network := range networks {
		name, err := network.GetName()
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to get network name: %v", err), http.StatusInternalServerError)
			return
		}

		resp = append(resp, response.NetworkResponse{
			Name:   name,
			Bridge: name,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// DeleteNetwork deletes a network on a server using the provided request.
func (c *ServerController) DeleteNetwork(w http.ResponseWriter, r *http.Request) {
	var req request.NetworkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate the request
	if req.Name == "" {
		http.Error(w, "Invalid name", http.StatusBadRequest)
		return
	}

	if req.ServerID <= 0 {
		http.Error(w, "Invalid serverID", http.StatusBadRequest)
		return
	}

	// Get the libvirt URI from the database
	serverDetail, err := c.dbService.GetServer(req.ServerID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get server details: %v", err), http.StatusInternalServerError)
		return
	}

	c.libvirtService.URI = serverDetail.LibvirtURI

	err = c.libvirtService.DeleteNetwork(req.Name)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete network: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
