package controller

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	request "github.com/sychonet/vdash-be/dto/request"
	response "github.com/sychonet/vdash-be/dto/response"
)

// CreateStoragePool creates a new storage pool using the provided request.
func (c *ServerController) CreateStoragePool(w http.ResponseWriter, r *http.Request) {
	var req request.StoragePoolRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get libvirt URI from database
	serverDetail, err := c.dbService.GetServer(req.ServerID)
	if err != nil {
		slog.Error("Failed to get libvirt uri from database: " + err.Error())
		http.Error(w, fmt.Sprintf("Failed to get server details: %v", err), http.StatusInternalServerError)
		return
	}

	c.libvirtService.URI = serverDetail.LibvirtURI

	// Create storage pool
	err = c.libvirtService.CreateStoragePool(req.Name, req.Path)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create storage pool: %v", err), http.StatusInternalServerError)
		return
	}

	// Prepare the response
	resp := response.StoragePoolResponse{
		Name: req.Name,
		Path: req.Path,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// GetStoragePools returns a list of all storage pools on a node.
func (c *ServerController) GetStoragePools(w http.ResponseWriter, r *http.Request) {
	// Fetch query parameter for server id
	serverIDStr := r.URL.Query().Get("serverID")
	if serverIDStr == "" {
		http.Error(w, "Missing serverID query parameter", http.StatusBadRequest)
		return
	}

	// Convert query parameter to integer for server id
	serverID, err := strconv.Atoi(serverIDStr)
	if err != nil {
		http.Error(w, "Invalid serverID query parameter", http.StatusBadRequest)
		return
	}

	// Get libvirt URI from database
	serverDetail, err := c.dbService.GetServer(serverID)
	if err != nil {
		slog.Error("Failed to get libvirt uri from database: " + err.Error())
		http.Error(w, fmt.Sprintf("Failed to get server details: %v", err), http.StatusInternalServerError)
		return
	}

	c.libvirtService.URI = serverDetail.LibvirtURI

	pools, err := c.libvirtService.GetStoragePools()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch pools details on node: %v", err), http.StatusInternalServerError)
		return
	}

	// Prepare the response
	var resp []response.StoragePoolListResponse
	for _, pool := range pools {
		name, err := pool.GetName()
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to get pool name: %v", err), http.StatusInternalServerError)
			return
		}

		info, err := pool.GetInfo()
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to get pool info: %v", err), http.StatusInternalServerError)
			return
		}

		resp = append(resp, response.StoragePoolListResponse{
			Name:      name,
			Available: fmt.Sprintf("%dG", info.Available),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// DeleteStoragePool deletes the specified storage pool on a node.
func (c *ServerController) DeleteStoragePool(w http.ResponseWriter, r *http.Request) {
	var req request.DeleteStoragePoolRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get libvirt uri from database
	serverDetail, err := c.dbService.GetServer(req.ServerID)
	if err != nil {
		slog.Error("Failed to get libvirt uri from database: " + err.Error())
		http.Error(w, fmt.Sprintf("Failed to get server details: %v", err), http.StatusInternalServerError)
		return
	}

	c.libvirtService.URI = serverDetail.LibvirtURI

	// Delete pool from node
	err = c.libvirtService.DeleteStoragePool(req.Name)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete storage pool on server: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
