package controller

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sychonet/vdash-be/db/entity"
	request "github.com/sychonet/vdash-be/dto/request"
	response "github.com/sychonet/vdash-be/dto/response"
)

func (c *ServerController) AddPublicIP(w http.ResponseWriter, r *http.Request) {
	var req request.AddPublicIPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate the request
	if req.ServerID <= 0 {
		http.Error(w, "Invalid serverID", http.StatusBadRequest)
		return
	}

	if req.IP == "" {
		http.Error(w, "Invalid publicIP", http.StatusBadRequest)
		return
	}

	ipInfo := entity.IPInfo{
		ServerID:  req.ServerID,
		IP:        req.IP,
		Available: req.Available,
	}

	// Insert the public IP details
	err := c.dbService.AddIP(ipInfo)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to insert public IP: %v", err), http.StatusInternalServerError)
		return
	}

	// Prepare the response
	resp := response.AddIPResponse{
		IP: req.IP,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (c *ServerController) GetAvailablePublicIPs(w http.ResponseWriter, r *http.Request) {
	ips, err := c.dbService.GetAvailablePublicIPs()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get active public IPs: %v", err), http.StatusInternalServerError)
		return
	}

	var availableIPs []response.GetAvailableIPsResponse
	for _, ipInfo := range ips {
		var ip response.GetAvailableIPsResponse
		ip.IP = ipInfo.IP
		ip.ServerID = ipInfo.ServerID
		availableIPs = append(availableIPs, ip)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(availableIPs); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (c *ServerController) DeletePublicIP(w http.ResponseWriter, r *http.Request) {
	var req request.DeletePublicIPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate the request
	if req.IP == "" {
		http.Error(w, "Invalid Public IP", http.StatusBadRequest)
		return
	}

	err := c.dbService.DeletePublicIP(req.IP)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete public IP: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (c *ServerController) UpdatePublicIP(w http.ResponseWriter, r *http.Request) {
	var req request.UpdatePublicIPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate the request
	if req.IP == "" {
		http.Error(w, "Invalid Public IP", http.StatusBadRequest)
		return
	}

	if req.ServerID <= 0 {
		http.Error(w, "Invalid serverID", http.StatusBadRequest)
		return
	}

	err := c.dbService.UpdatePublicIP(req.IP, req.ServerID, req.Available)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update public IP: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
