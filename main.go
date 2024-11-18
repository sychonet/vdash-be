package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"libvirt.org/go/libvirt"
)

type DiskRequest struct {
	Format string `json:"format"`
	Name   string `json:"name"`
	Size   string `json:"size"`
}

type DiskResponse struct {
	Name   string `json:"name"`
	Format string `json:"format"`
	Size   string `json:"size"`
}

func createDisk(w http.ResponseWriter, r *http.Request) {
	var req DiskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Connect to libvirtd
	conn, err := libvirt.NewConnect("qemu:///system")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to connect to libvirt: %v", err), http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	// Define the storage pool and volume XML
	poolName := "default"
	volumeXML := fmt.Sprintf(`
    <volume>
        <name>%s</name>
        <capacity unit="G">%s</capacity>
        <target>
            <format type='%s'/>
        </target>
    </volume>`, req.Name, req.Size, req.Format)

	// Lookup the storage pool
	pool, err := conn.LookupStoragePoolByName(poolName)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to find storage pool: %v", err), http.StatusInternalServerError)
		return
	}
	defer pool.Free()

	// Create the storage volume
	volume, err := pool.StorageVolCreateXML(volumeXML, 0)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create storage volume: %v", err), http.StatusInternalServerError)
		return
	}
	defer volume.Free()

	// Prepare the response
	resp := DiskResponse{
		Name:   req.Name,
		Format: req.Format,
		Size:   req.Size,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Post("/v1/disks", createDisk)

	http.ListenAndServe(":7201", r)
}
