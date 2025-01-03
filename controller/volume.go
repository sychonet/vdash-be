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

// CreateVolume creates a new storage volume using the provided request.
func (c *ServerController) CreateVolume(w http.ResponseWriter, r *http.Request) {
	var req request.CreateVolumeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	serverID := req.ServerID

	if serverID < 0 {
		http.Error(w, "Invalid serverID", http.StatusBadRequest)
		return
	}

	if req.Size <= 0 {
		http.Error(w, "Invalid size", http.StatusBadRequest)
		return
	}

	if serverID == 0 {
		// Get serverID from volume scheduler with given storage pool name
		serverID, err := c.schedulerService.GetServerIDForVolume(req.PoolName, req.Size)

		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to get serverID: %v", err), http.StatusInternalServerError)
			return
		}

		if serverID == 0 {
			http.Error(w, "No server available", http.StatusNotFound)
			return
		}
	}

	// Get the libvirt uri from database
	serverDetail, err := c.dbService.GetServer(serverID)
	if err != nil {
		slog.Error("Failed to get libvirt uri from database: " + err.Error())
		http.Error(w, fmt.Sprintf("Failed to get server details: %v", err), http.StatusInternalServerError)
		return
	}

	c.libvirtService.URI = serverDetail.LibvirtURI

	// Create the storage volume
	err = c.libvirtService.CreateStorageVolume(req.PoolName, req.Name, req.Format, req.Size)

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create storage volume: %v", err), http.StatusInternalServerError)
		return
	}

	// Prepare the response
	resp := response.CreateVolumeResponse{
		ServerID: serverID,
		Name:     req.Name,
		Format:   req.Format,
		Size:     uint64(req.Size),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// GetVolumes returns a list of all storage volumes on a given server in a specified storage pool.
func (c *ServerController) GetVolumes(w http.ResponseWriter, r *http.Request) {
	// Get the storage pool name from query parameter
	poolName := r.URL.Query().Get("poolName")
	serverIDParam := r.URL.Query().Get("serverID")

	if poolName == "" {
		http.Error(w, "Missing poolName query parameter", http.StatusBadRequest)
		return
	}

	if serverIDParam == "" {
		http.Error(w, "Missing serverID query parameter", http.StatusBadRequest)
		return
	}

	serverID, err := strconv.Atoi(serverIDParam)
	if err != nil {
		http.Error(w, "Invalid serverID query parameter", http.StatusBadRequest)
		return
	}

	// Get the libvirt uri from database
	serverDetail, err := c.dbService.GetServer(serverID)
	if err != nil {
		slog.Error("Failed to get libvirt uri from database: " + err.Error())
		http.Error(w, fmt.Sprintf("Failed to get server details: %v", err), http.StatusInternalServerError)
		return
	}

	c.libvirtService.URI = serverDetail.LibvirtURI

	// List all storage volumes on a given server under a given storage pool
	volumes, err := c.libvirtService.GetStorageVolumesOnServer(poolName)

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list storage volumes: %v", err), http.StatusInternalServerError)
		return
	}
	defer func() {
		for _, volume := range volumes {
			volume.Free()
		}
	}()

	// Prepare the response
	var resp []response.CreateVolumeResponse
	for _, volume := range volumes {
		name, err := volume.GetName()
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to get volume name: %v", err), http.StatusInternalServerError)
			return
		}

		info, err := volume.GetInfo()
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to get volume info: %v", err), http.StatusInternalServerError)
			return
		}

		resp = append(resp, response.CreateVolumeResponse{
			Name:   name,
			Format: fmt.Sprintf("%d", info.Type),
			Size:   info.Capacity,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// DeleteVolume deletes the specified disk image.
func (c *ServerController) DeleteVolume(w http.ResponseWriter, r *http.Request) {
	var req request.DeleteVolumeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get the libvirt uri from database
	serverDetail, err := c.dbService.GetServer(req.ServerID)
	if err != nil {
		slog.Error("Failed to get libvirt uri from database: " + err.Error())
		http.Error(w, fmt.Sprintf("Failed to get server details: %v", err), http.StatusInternalServerError)
		return
	}

	c.libvirtService.URI = serverDetail.LibvirtURI

	err = c.libvirtService.DeleteStorageVolume(req.PoolName, req.Name)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete storage volume: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
