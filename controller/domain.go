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

// CreateDomain creates a new domain using the provided request.
func (c *ServerController) CreateDomain(w http.ResponseWriter, r *http.Request) {
	var req request.CreateDomainRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate the request
	if req.Name == "" {
		http.Error(w, "Invalid name", http.StatusBadRequest)
		return
	}

	if req.Memory <= 0 {
		http.Error(w, "Invalid memory", http.StatusBadRequest)
		return
	}

	if req.VCPU <= 0 {
		http.Error(w, "Invalid VCPU", http.StatusBadRequest)
		return
	}

	// TODO: Validate the disk and network

	serverID := req.ServerID

	if serverID < 0 {
		http.Error(w, "Invalid serverID", http.StatusBadRequest)
		return
	}

	if serverID == 0 {
		// Get availableServerID from domain scheduler
		availableServerID, err := c.schedulerService.GetServerIDForCreatingDomain(req.Memory, req.VCPU, req.PublicIP)
		serverID = availableServerID
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to get serverID: %v", err), http.StatusInternalServerError)
			return
		}
	}

	if serverID > 0 {
		if req.PublicIP {
			// Check if the server has a public IP
			ip, err := c.dbService.CheckPublicIPAvailable(serverID)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to check public IP: %v", err), http.StatusNotFound)
				return
			}
			slog.Info("Public IP available is: " + ip)
			// TODO: Assign the public IP to the domain
		}

		// Get the libvirt URI from the database
		serverDetail, err := c.dbService.GetServer(serverID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to get server details: %v", err), http.StatusInternalServerError)
			return
		}

		c.libvirtService.URI = serverDetail.LibvirtURI
	}

	// Define the domain XML
	domainXML := fmt.Sprintf(`
	<domain type='kvm'>
	<name>%s</name>
	<memory unit='KiB'>%d</memory>
	<vcpu>%d</vcpu>
	<os>
		<type arch='x86_64' machine='pc-i440fx-2.9'>hvm</type>
		<boot dev='hd'/>
	</os>
	<devices>`, req.Name, req.Memory*1024, req.VCPU)

	for _, disk := range req.Disks {
		domainXML += fmt.Sprintf(`
    <disk type='file' device='disk'>
      <driver name='qemu' type='qcow2'/>
      <source file='%s'/>
      <target dev='vda' bus='virtio'/>
    </disk>`, disk)
	}

	for _, network := range req.Networks {
		domainXML += fmt.Sprintf(`
    <interface type='network'>
      <source network='%s'/>
      <model type='virtio'/>
    </interface>`, network)
	}

	domainXML += `
	</devices>
	</domain>`

	err := c.libvirtService.CreateDomain(req.Name, domainXML)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create domain: %v", err), http.StatusInternalServerError)
		return
	}

	// Prepare the response
	resp := response.CreateDomainResponse{
		Name:     req.Name,
		Memory:   req.Memory,
		VCPU:     req.VCPU,
		Disks:    req.Disks,
		Networks: req.Networks,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// GetDomains retrieves a list of virtual machines on a given server.
func (c *ServerController) GetDomains(w http.ResponseWriter, r *http.Request) {
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

	domains, err := c.libvirtService.GetDomains()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list domains: %v", err), http.StatusInternalServerError)
		return
	}

	defer func() {
		for _, domain := range domains {
			domain.Free()
		}
	}()

	// Prepare the response
	var resp []response.GetDomainResponse
	for _, domain := range domains {
		name, err := domain.GetName()
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to get domain name: %v", err), http.StatusInternalServerError)
			return
		}

		info, err := domain.GetInfo()
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to get domain info: %v", err), http.StatusInternalServerError)
			return
		}

		resp = append(resp, response.GetDomainResponse{
			Name:   name,
			Memory: info.Memory,
			VCPU:   info.NrVirtCpu,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// DeleteDomain deletes a domain on a given server.
func (c *ServerController) DeleteDomain(w http.ResponseWriter, r *http.Request) {
	var req request.DeleteDomainRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get the libvirt URI from the database
	serverDetail, err := c.dbService.GetServer(req.ServerID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get server details: %v", err), http.StatusInternalServerError)
		return
	}

	c.libvirtService.URI = serverDetail.LibvirtURI

	err = c.libvirtService.DeleteDomain(req.Name)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete domain: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}
