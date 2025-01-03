package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sychonet/vdash-be/config"
	controller "github.com/sychonet/vdash-be/controller"
	entity "github.com/sychonet/vdash-be/db/entity"
	"github.com/sychonet/vdash-be/service"
)

// init is called before the main function. It initializes the application by fetching the servers from the cloud service provider Scaleway and inserting them into the database if the database has got no entry in it.
func init() {
	config.LoadConfig()

	scalewayService := service.NewScalewayService(config.AppConfig.Scaleway.BaseURL, config.AppConfig.Scaleway.Token)
	databaseService := service.NewDatabaseService(config.AppConfig.Database.Host, config.AppConfig.Database.Port, config.AppConfig.Database.Username, config.AppConfig.Database.Password, config.AppConfig.Database.Name, config.AppConfig.Database.ServersCollection, config.AppConfig.Database.IPsCollection)

	count, err := databaseService.CountServers()

	if count > 0 {
		slog.Info("Servers already present in the database")
		return
	}

	var wg sync.WaitGroup

	// Get servers from cloud service provider Scaleway
	servers, err := scalewayService.GetServers()
	if err != nil {
		panic(err)
	}

	serverDetailsChan := make(chan entity.ServerInfo, len(servers))

	// For each node save id, public ip, hostname (only with prefix "pr"). For this we need to make API requests in parallel using goroutines but we need to make sure that we don't breach API quota contract.
	for _, server := range servers {
		wg.Add(1)
		// Get server details
		go func(server string) {
			defer wg.Done()
			serverDetails, err := scalewayService.GetServerDetails(server)

			fmt.Println(serverDetails)
			if err != nil {
				panic(err)
			}

			if strings.HasPrefix(serverDetails.Hostname, "pr") {
				// Find public IP of the server and generate URI for libvirt connection from it
				var publicIP string
				var libvirtURI string

				for _, ip := range serverDetails.IP {
					if ip.Type == "public" {
						publicIP = ip.Address
						// URI for libvirt connection for each node will be "qemu+libssh://root@{public ip}/system"
						libvirtURI = "qemu+libssh://root@" + publicIP + "/system"
						break
					}
				}

				entity := entity.ServerInfo{
					ID:         serverDetails.ID,
					Hostname:   serverDetails.Hostname,
					PublicIP:   publicIP,
					LibvirtURI: libvirtURI,
				}

				serverDetailsChan <- entity
			}
		}(server)
	}

	wg.Wait()
	var serverDetails []interface{}

	for serverDetail := range serverDetailsChan {
		serverDetails = append(serverDetails, serverDetail)
	}
	close(serverDetailsChan)

	// Insert the servers in the database
	err = databaseService.AddServers(serverDetails)
	if err != nil {
		panic(err)
	}

	slog.Info("Servers inserted in the database")
}

// main is the entrypoint for the application.
func main() {
	config.LoadConfig()

	scalewayService := service.NewScalewayService(config.AppConfig.Scaleway.BaseURL, config.AppConfig.Scaleway.Token)
	databaseService := service.NewDatabaseService(config.AppConfig.Database.Host, config.AppConfig.Database.Port, config.AppConfig.Database.Username, config.AppConfig.Database.Password, config.AppConfig.Database.Name, config.AppConfig.Database.ServersCollection, config.AppConfig.Database.IPsCollection)
	libvirtService := service.NewLibvirtService("")
	schedulerService := service.NewSchedulerService(databaseService, libvirtService)

	serverController := controller.NewServerController(scalewayService, databaseService, libvirtService, schedulerService)

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// While creating a resource such as disk, network, storage pool, or virtual machine, the user may provide the node hostname. If the hostname is provided, the application will connect to the node using the URI and create the resource on that node. If the hostname is not provided, the application will use the scheduler to decide which node should be picked.
	// Scheduler -> CPU, Memory, Disks, Networks, Public IP (if required)
	// Public IP = False -> Scheduler will pick a node based on CPU, Memory, Disks
	// Public IP = True -> Scheduler will first fetch list of nodes where failover IPs are available pick a node based on CPU, Memory, Disks

	// Define routes
	r.Post("/v1/servers", serverController.CreateServer)
	r.Get("/v1/servers", serverController.GetServers)
	r.Delete("/v1/servers", serverController.DeleteServer)
	r.Post("/v1/ips", serverController.AddPublicIP)
	r.Get("/v1/ips", serverController.GetAvailablePublicIPs)
	r.Put("/v1/ips", serverController.UpdatePublicIP)
	r.Delete("/v1/ips", serverController.DeletePublicIP)
	r.Post("/v1/storage/pools", serverController.CreateStoragePool)
	r.Get("/v1/storage/pools", serverController.GetStoragePools)
	r.Delete("/v1/storage/pools", serverController.DeleteStoragePool)
	r.Post("/v1/storage/volumes", serverController.CreateVolume)
	r.Get("/v1/storage/volumes", serverController.GetVolumes)
	r.Delete("/v1/storage/volumes", serverController.DeleteVolume)
	r.Post("/v1/networks", serverController.CreateNetwork)
	r.Get("/v1/networks", serverController.GetNetworks)
	r.Delete("/v1/networks", serverController.DeleteNetwork)
	r.Post("/v1/domains", serverController.CreateDomain)
	r.Get("/v1/domains", serverController.GetDomains)
	r.Delete("/v1/domains", serverController.DeleteDomain)

	http.ListenAndServe(":"+config.AppConfig.Application.Port, r)
}
