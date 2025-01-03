package request

// CreateServerRequest represents a request to create a new server.
type CreateServerRequest struct {
	ID       int    `json:"id"`
	Hostname string `json:"hostname"`
	PublicIP string `json:"publicIP"`
}

// DeleteServerRequest represents a request to delete a server.
type DeleteServerRequest struct {
	ID int `json:"id"`
}

// CreateVolumeRequest represents a request to create a new volume in a storage pool.
type CreateVolumeRequest struct {
	ServerID int    `json:"serverID"`
	PoolName string `json:"poolName"`
	Format   string `json:"format"`
	Name     string `json:"name"`
	Size     int    `json:"size"`
}

// DeleteVolumeRequest represents a request to delete a storage volume in a storage pool.
type DeleteVolumeRequest struct {
	ServerID int    `json:"serverID"`
	PoolName string `json:"poolName"`
	Name     string `json:"name"`
}

// NetworkRequest represents a request to create a new network.
type NetworkRequest struct {
	ServerID int    `json:"serverID"`
	Name     string `json:"name"`
	Bridge   string `json:"bridge"`
}

// StoragePoolRequest represents a request to create a new storage pool.
type StoragePoolRequest struct {
	ServerID int    `json:"serverID"`
	Name     string `json:"name"`
	Path     string `json:"path"`
}

// DeleteStoragePoolRequest represents a request to delete a storage pool.
type DeleteStoragePoolRequest struct {
	ServerID int    `json:"serverID"`
	Name     string `json:"name"`
}

// CreateDomainRequest represents a request to create a new domain.
type CreateDomainRequest struct {
	ServerID int      `json:"serverID"`
	Name     string   `json:"name"`
	Memory   uint64   `json:"memory"`
	VCPU     uint     `json:"vcpu"`
	Disks    []string `json:"disks"`
	Networks []string `json:"networks"`
	PublicIP bool     `json:"publicIP"`
}

// AddPublicIPRequest represents a request to add a public IP for a server.
type AddPublicIPRequest struct {
	ServerID  int    `json:"serverID"`
	IP        string `json:"ip"`
	Available bool   `json:"available"`
}

// DeletePublicIPRequest represents a request to delete a public IP for a server.
type DeletePublicIPRequest struct {
	IP string `json:"ip"`
}

// UpdatePublicIPRequest represents a request to update a public IP for a server.
type UpdatePublicIPRequest struct {
	ServerID  int    `json:"serverID"`
	IP        string `json:"ip"`
	Available bool   `json:"available"`
}

// DeleteDomainRequest represents a request to delete a domain.
type DeleteDomainRequest struct {
	ServerID int    `json:"serverID"`
	Name     string `json:"name"`
}
