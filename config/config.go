package config

import (
	"encoding/json"
	"log"

	_ "embed"
)

//go:embed config.json
var configData []byte

type Config struct {
	Application ApplicationConfig `json:"application"`
	Scaleway    ScalewayConfig    `json:"scaleway"`
	Database    DatabaseConfig    `json:"database"`
}

type ApplicationConfig struct {
	Port        string `json:"port"`
	Environment string `json:"env"`
}

type ScalewayConfig struct {
	BaseURL string `json:"baseurl"`
	Token   string `json:"token"`
}

type DatabaseConfig struct {
	Host              string `json:"host"`
	Port              string `json:"port"`
	Username          string `json:"username"`
	Password          string `json:"password"`
	Name              string `json:"name"`
	ServersCollection string `json:"serversCollection"`
	IPsCollection     string `json:"ipsCollection"`
}

var AppConfig Config

func LoadConfig() {
	// data, err := os.ReadFile("config.json")
	// if err != nil {
	// 	log.Fatalf("Error reading config file, %s", err)
	// }

	err := json.Unmarshal(configData, &AppConfig)
	if err != nil {
		log.Fatalf("Error parsing config file, %s", err)
	}
}
