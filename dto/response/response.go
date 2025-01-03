package response

// CreateServerResponse represents a response to a server creation request.
type CreateServerResponse struct {
	ID         int    `json:"id"`
	Hostname   string `json:"hostname"`
	PublicIP   string `json:"publicIP"`
	LibvirtURI string `json:"libvirtURI"`
}

// GetServersResponse represents a response to a server list request.
type GetServersResponse struct {
	ID         int    `json:"id"`
	Hostname   string `json:"hostname"`
	PublicIP   string `json:"publicIP"`
	LibvirtURI string `json:"libvirtURI"`
}

// CreateVolumeResponse represents a response to a storage volume creation request.
type CreateVolumeResponse struct {
	ServerID int    `json:"serverID"`
	Name     string `json:"name"`
	Format   string `json:"format"`
	Size     uint64 `json:"size"`
}

// NetworkResponse represents a response to a network creation request.
type NetworkResponse struct {
	Name   string `json:"name"`
	Bridge string `json:"bridge"`
}

// StoragePoolResponse represents a response to a storage pool creation request.
type StoragePoolResponse struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// StoragePoolListResponse represents a response to a storage pool list request.
type StoragePoolListResponse struct {
	Name      string `json:"name"`
	Available string `json:"available"`
}

// CreateDomainResponse represents a response to a domain creation request.
type CreateDomainResponse struct {
	Name     string   `json:"name"`
	Memory   uint64   `json:"memory"`
	VCPU     uint     `json:"vcpu"`
	Disks    []string `json:"disks"`
	Networks []string `json:"networks"`
}

// ScalewayServerResponse represents a response from the Scaleway API GET https://api.online.net/api/v1/server/{server_id}.
type ScalewayServerResponse struct {
	ID       int                        `json:"id"`
	Hostname string                     `json:"hostname"`
	IP       []ScalewayServerResponseIP `json:"ip"`
}

// ScalewayServerResponseIP represents the ip object in the ScalewayServerResponse.
type ScalewayServerResponseIP struct {
	Address string `json:"address"`
	Type    string `json:"type"`
}

// AddIPResponse represents a response to an IP addition request.
type AddIPResponse struct {
	IP string `json:"publicIP"`
}

// GetAvailableIPsResponse represents a response to an IP list request.
type GetAvailableIPsResponse struct {
	IP       string `json:"publicIP"`
	ServerID int    `json:"serverID"`
}

// GetDomainResponse represents a response to a domain list request.
type GetDomainResponse struct {
	Name   string `json:"name"`
	Memory uint64 `json:"memory"`
	VCPU   uint   `json:"vcpu"`
}
