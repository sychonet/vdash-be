package entity

// ServerInfo represents a server information.
type ServerInfo struct {
	ID         int    `bson:"_id"`
	Hostname   string `bson:"hostname"`
	PublicIP   string `bson:"publicIP"`
	LibvirtURI string `bson:"libvirtURI"`
}

// IPInfo represents public ip information associated with a server.
type IPInfo struct {
	IP        string `bson:"_id"`
	ServerID  int    `bson:"serverID"`
	Available bool   `bson:"available"`
}
