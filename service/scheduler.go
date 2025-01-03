package service

import (
	"log/slog"
	"slices"

	"github.com/sychonet/vdash-be/db/entity"
)

// SchedulerService is a service that schedules the creation of resources on servers. It uses the DatabaseService and LibvirtService to get the server details, connect to libvirtd and find the best possible server with available resources.
type SchedulerService struct {
	databaseService *DatabaseService
	libvirtService  *LibvirtService
}

func NewSchedulerService(databaseService *DatabaseService, libvirtService *LibvirtService) *SchedulerService {
	return &SchedulerService{
		databaseService: databaseService,
		libvirtService:  libvirtService,
	}
}

func (s *SchedulerService) GetServerIDForVolume(poolName string, size int) (int, error) {
	var libvirtURIs []string
	// Get all the server details from the database
	serverDetails, err := s.databaseService.GetServers()
	if err != nil {
		return 0, err
	}

	// Convert size from GB to bytes
	requiredSpace := uint64(size * 1024 * 1024 * 1024)

	for _, server := range serverDetails {
		libvirtURIs = append(libvirtURIs, server.LibvirtURI)
	}

	// Check all servers for pool with sufficient space
	results := s.libvirtService.CheckPoolsForSpace(libvirtURIs, poolName, requiredSpace)

	for _, result := range results {
		if result.HasSpace {
			index := slices.Index(libvirtURIs, result.LibvirtURI)
			// Return the server id where the pool has enough space left to create the volume
			return serverDetails[index].ID, nil
		}
	}

	return 0, nil
}

func (s *SchedulerService) GetServerIDForCreatingDomain(memory uint64, vcpu uint, publicIP bool) (int, error) {
	var serverIDs []int
	var serverDetails []entity.ServerInfo
	var libvirtURIs []string
	if publicIP {
		// Get all the servers with public IP available from the database
		availableServers, err := s.databaseService.GetAvailablePublicIPs()
		if err != nil {
			slog.Error("Failed to get serverIDs for available public IPs: " + err.Error())
			return 0, err
		}
		for _, server := range availableServers {
			serverIDs = append(serverIDs, server.ServerID)
		}

		// Get all the server details from the database for servers with public IP available
		servers, err := s.databaseService.GetServersWithIDs(serverIDs)
		serverDetails = servers
		if err != nil {
			slog.Error("Failed to get server details for serverIDs: " + err.Error())
			return 0, err
		}
	} else {
		// Get all the server details from the database
		servers, err := s.databaseService.GetServers()
		serverDetails = servers
		if err != nil {
			slog.Error("Failed to get server details: " + err.Error())
			return 0, err
		}
	}

	// Get list of eligible libvirtURIs
	for _, server := range serverDetails {
		libvirtURIs = append(libvirtURIs, server.LibvirtURI)
	}

	// Check all servers for available resources
	results := s.libvirtService.CheckServersForResources(libvirtURIs, memory, vcpu)

	for _, result := range results {
		if result.HasResources {
			index := slices.Index(libvirtURIs, result.LibvirtURI)
			// Return the server id where the domain can be created
			return serverDetails[index].ID, nil
		}
	}

	return 0, nil
}

// func (s *SchedulerService) GetServerIDForNetwork() (int, error) {
// 	return 0, nil
// }
