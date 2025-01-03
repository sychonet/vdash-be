package service

import (
	"fmt"
	"log/slog"
	"sync"

	"libvirt.org/go/libvirt"
)

type LibvirtService struct {
	URI string
}

type PoolCheckResult struct {
	LibvirtURI string
	PoolName   string
	HasSpace   bool
	Error      error
}

type ResourceCheckResult struct {
	LibvirtURI   string
	HasResources bool
	Error        error
}

func NewLibvirtService(uri string) *LibvirtService {
	return &LibvirtService{URI: uri}
}

// CreateStoragePool creates a new storage pool on a libvirt host.
func (l *LibvirtService) CreateStoragePool(name, path string) error {
	// Connect to libvirtd
	conn, err := libvirt.NewConnect(l.URI)
	if err != nil {
		slog.Error("Failed to connect to libvirt: " + err.Error())
		return err
	}
	defer conn.Close()

	// Define the storage pool XML
	poolXML := fmt.Sprintf(`
	<pool type='dir'>
		<name>%s</name>
		<target>
			<path>%s</path>
		</target>
	</pool>`, name, path)

	// Create the storage pool
	pool, err := conn.StoragePoolDefineXML(poolXML, 0)
	if err != nil {
		slog.Error("Failed to create storage pool: " + err.Error())
		return err
	}
	defer pool.Free()

	// Start the storage pool
	if err := pool.Create(0); err != nil {
		slog.Error("Failed to start storage pool: " + err.Error())
		return err
	}

	return nil
}

// GetStoragePools returns a list of all storage pools on a libvirt host.
func (l *LibvirtService) GetStoragePools() ([]libvirt.StoragePool, error) {
	// Connect to libvirtd
	conn, err := libvirt.NewConnect(l.URI)
	if err != nil {
		slog.Error("Failed to connect to libvirt: " + err.Error())
		return nil, err
	}
	defer conn.Close()

	// List all storage pools
	pools, err := conn.ListAllStoragePools(0)
	if err != nil {
		slog.Error("Failed to list storage pools: " + err.Error())
		return nil, err
	}
	defer func() {
		for _, pool := range pools {
			pool.Free()
		}
	}()

	return pools, nil
}

func (l *LibvirtService) DeleteStoragePool(name string) error {
	// Connect to libvirtd
	conn, err := libvirt.NewConnect(l.URI)
	if err != nil {
		slog.Error("Failed to connect to libvirt: " + err.Error())
		return err
	}
	defer conn.Close()

	// Lookup the storage pool
	pool, err := conn.LookupStoragePoolByName(name)
	if err != nil {
		slog.Error("Failed to find storage pool: " + err.Error())
		return err
	}
	defer pool.Free()

	// Destroy the storage pool
	if err := pool.Destroy(); err != nil {
		slog.Error("Failed to destroy storage pool: " + err.Error())
		return err
	}

	// Undefine the storage pool
	if err := pool.Undefine(); err != nil {
		slog.Error("Failed to undefine storage pool: " + err.Error())
		return err
	}

	return nil
}

func (l *LibvirtService) CreateStorageVolume(poolName, format, name string, size int) error {
	// Connect to libvirtd
	conn, err := libvirt.NewConnect(l.URI)
	if err != nil {
		slog.Error("Failed to connect to libvirt: " + err.Error())
		return err
	}
	defer conn.Close()

	// Lookup the storage pool
	pool, err := conn.LookupStoragePoolByName(poolName)
	if err != nil {
		slog.Error("Failed to find storage pool: " + err.Error())
		return err
	}
	defer pool.Free()

	// Define the storage volume XML
	volumeXML := fmt.Sprintf(`
	<volume>
		<name>%s</name>
		<capacity unit="G">%d</capacity>
		<target>
			<format type='%s'/>
		</target>
	</volume>`, name, size, format)

	// Create the storage volume
	volume, err := pool.StorageVolCreateXML(volumeXML, 0)
	if err != nil {
		slog.Error("Failed to create storage volume: " + err.Error())
		return err
	}
	defer volume.Free()

	return nil
}

func (l *LibvirtService) CheckPoolsForSpace(libvirtURIs []string, poolName string, requiredSpace uint64) []PoolCheckResult {
	var wg sync.WaitGroup
	results := make(chan PoolCheckResult, len(libvirtURIs))

	for _, uri := range libvirtURIs {
		wg.Add(1)
		go checkPoolForSpace(uri, poolName, requiredSpace, &wg, results)
	}

	wg.Wait()
	close(results)

	var checkResults []PoolCheckResult
	for result := range results {
		checkResults = append(checkResults, result)
	}

	return checkResults
}

func checkPoolForSpace(libvirtURI, poolName string, requiredSpace uint64, wg *sync.WaitGroup, results chan<- PoolCheckResult) {
	defer wg.Done()

	conn, err := libvirt.NewConnect(libvirtURI)
	if err != nil {
		slog.Error("Failed to connect to libvirt: " + err.Error())
		results <- PoolCheckResult{LibvirtURI: libvirtURI, PoolName: poolName, HasSpace: false, Error: err}
		return
	}
	defer conn.Close()

	pool, err := conn.LookupStoragePoolByName(poolName)
	if err != nil {
		slog.Error("Failed to find storage pool: " + err.Error())
		results <- PoolCheckResult{LibvirtURI: libvirtURI, PoolName: poolName, HasSpace: false, Error: err}
		return
	}
	defer pool.Free()

	info, err := pool.GetInfo()
	if err != nil {
		slog.Error("Failed to get storage pool info: " + err.Error())
		results <- PoolCheckResult{LibvirtURI: libvirtURI, PoolName: poolName, HasSpace: false, Error: err}
		return
	}

	availableSpace := info.Available
	hasSpace := availableSpace >= requiredSpace

	results <- PoolCheckResult{LibvirtURI: libvirtURI, PoolName: poolName, HasSpace: hasSpace, Error: nil}
}

func (l *LibvirtService) GetStorageVolumesOnServer(poolName string) ([]libvirt.StorageVol, error) {
	// Connect to libvirtd
	conn, err := libvirt.NewConnect(l.URI)
	if err != nil {
		slog.Error("Failed to connect to libvirt: " + err.Error())
		return nil, err
	}
	defer conn.Close()

	// Lookup the storage pool
	pool, err := conn.LookupStoragePoolByName(poolName)
	if err != nil {
		slog.Error("Failed to find storage pool: " + err.Error())
		return nil, err
	}
	defer pool.Free()

	// List all storage volumes
	volumes, err := pool.ListAllStorageVolumes(0)
	if err != nil {
		slog.Error("Failed to list storage volumes: " + err.Error())
		return nil, err
	}

	return volumes, nil
}

func (l *LibvirtService) DeleteStorageVolume(poolName, volumeName string) error {
	// Connect to libvirtd
	conn, err := libvirt.NewConnect(l.URI)
	if err != nil {
		slog.Error("Failed to connect to libvirt: " + err.Error())
		return err
	}
	defer conn.Close()

	// Lookup the storage pool
	pool, err := conn.LookupStoragePoolByName(poolName)
	if err != nil {
		slog.Error("Failed to find storage pool: " + err.Error())
		return err
	}
	defer pool.Free()

	// Lookup the storage volume
	volume, err := pool.LookupStorageVolByName(volumeName)
	if err != nil {
		slog.Error("Failed to find storage volume: " + err.Error())
		return err
	}
	defer volume.Free()

	// Delete the storage volume
	if err := volume.Delete(0); err != nil {
		slog.Error("Failed to delete storage volume: " + err.Error())
		return err
	}

	return nil
}

func (l *LibvirtService) CreateNetwork(name, bridge string) error {
	// Connect to libvirtd
	conn, err := libvirt.NewConnect(l.URI)
	if err != nil {
		slog.Error("Failed to connect to libvirt: " + err.Error())
		return err
	}
	defer conn.Close()

	// Define the network XML
	networkXML := fmt.Sprintf(`
	<network>
		<name>%s</name>
		<bridge name='%s'/>
	</network>`, name, bridge)

	// Create the network
	network, err := conn.NetworkDefineXML(networkXML)
	if err != nil {
		slog.Error("Failed to create network: " + err.Error())
		return err
	}
	defer network.Free()

	// Start the network
	if err := network.Create(); err != nil {
		slog.Error("Failed to start network: " + err.Error())
		return err
	}

	return nil
}

func (l *LibvirtService) GetNetworks() ([]libvirt.Network, error) {
	// Connect to libvirtd
	conn, err := libvirt.NewConnect(l.URI)
	if err != nil {
		slog.Error("Failed to connect to libvirt: " + err.Error())
		return nil, err
	}
	defer conn.Close()

	// List all networks
	networks, err := conn.ListAllNetworks(0)
	if err != nil {
		slog.Error("Failed to list networks: " + err.Error())
		return nil, err
	}

	return networks, nil
}

func (l *LibvirtService) DeleteNetwork(name string) error {
	// Connect to libvirtd
	conn, err := libvirt.NewConnect(l.URI)
	if err != nil {
		slog.Error("Failed to connect to libvirt: " + err.Error())
		return err
	}
	defer conn.Close()

	// Lookup the network
	network, err := conn.LookupNetworkByName(name)
	if err != nil {
		slog.Error("Failed to find network: " + err.Error())
		return err
	}
	defer network.Free()

	// Destroy the network
	if err := network.Destroy(); err != nil {
		slog.Error("Failed to destroy network: " + err.Error())
		return err
	}

	// Undefine the network
	if err := network.Undefine(); err != nil {
		slog.Error("Failed to undefine network: " + err.Error())
		return err
	}

	return nil
}

func (l *LibvirtService) CreateDomain(name, xml string) error {
	// Connect to libvirtd
	conn, err := libvirt.NewConnect(l.URI)
	if err != nil {
		slog.Error("Failed to connect to libvirt: " + err.Error())
		return err
	}
	defer conn.Close()

	// Define the domain XML
	domain, err := conn.DomainDefineXML(xml)
	if err != nil {
		slog.Error("Failed to define domain: " + err.Error())
		return err
	}
	defer domain.Free()

	// Start the domain
	if err := domain.Create(); err != nil {
		slog.Error(("Failed to start domain: " + err.Error()))
		return err
	}

	return nil
}

func (l *LibvirtService) GetDomains() ([]libvirt.Domain, error) {
	// Connect to libvirtd
	conn, err := libvirt.NewConnect(l.URI)
	if err != nil {
		slog.Error("Failed to connect to libvirt: " + err.Error())
		return nil, err
	}
	defer conn.Close()

	// Retrieve the list of domains
	domains, err := conn.ListAllDomains(libvirt.CONNECT_LIST_DOMAINS_ACTIVE)
	if err != nil {
		slog.Error("Failed to list domains: " + err.Error())
		return nil, err
	}

	return domains, nil
}

func (l *LibvirtService) DeleteDomain(name string) error {
	// Connect to libvirtd
	conn, err := libvirt.NewConnect(l.URI)
	if err != nil {
		slog.Error("Failed to connect to libvirt: " + err.Error())
		return err
	}
	defer conn.Close()

	// Lookup the domain
	domain, err := conn.LookupDomainByName(name)
	if err != nil {
		slog.Error("Failed to find domain: " + err.Error())
		return err
	}
	defer domain.Free()

	// Destroy the domain
	if err := domain.Destroy(); err != nil {
		slog.Error("Failed to destroy domain: " + err.Error())
		return err
	}

	// Undefine the domain
	if err := domain.Undefine(); err != nil {
		slog.Error("Failed to undefine domain: " + err.Error())
		return err
	}

	return nil
}

func checkResources(libvirtURI string, requiredMemory uint64, requiredVCPU uint, wg *sync.WaitGroup, results chan<- ResourceCheckResult) {
	defer wg.Done()

	conn, err := libvirt.NewConnect(libvirtURI)
	if err != nil {
		slog.Error("Failed to connect to libvirt: " + err.Error())
		results <- ResourceCheckResult{LibvirtURI: libvirtURI, HasResources: false, Error: err}
		return
	}
	defer conn.Close()

	nodeInfo, err := conn.GetNodeInfo()
	if err != nil {
		slog.Error("Failed to get node info: " + err.Error())
		results <- ResourceCheckResult{LibvirtURI: libvirtURI, HasResources: false, Error: err}
		return
	}

	// Total resources
	totalMemory := nodeInfo.Memory * 1024 // Convert from KB to bytes
	totalVCPU := nodeInfo.Cpus

	// Calculate used resources
	var usedMemory uint64
	var usedVCPU uint

	domains, err := conn.ListAllDomains(0)
	if err != nil {
		slog.Error("Failed to list domains: " + err.Error())
		results <- ResourceCheckResult{LibvirtURI: libvirtURI, HasResources: false, Error: err}
		return
	}
	defer func() {
		for _, domain := range domains {
			domain.Free()
		}
	}()

	for _, domain := range domains {
		domainInfo, err := domain.GetInfo()
		if err != nil {
			slog.Error("Failed to get domain info: " + err.Error())
			results <- ResourceCheckResult{LibvirtURI: libvirtURI, HasResources: false, Error: err}
			return
		}
		usedMemory += domainInfo.Memory * 1024 // Convert from KB to bytes
		usedVCPU += domainInfo.NrVirtCpu
	}

	// Calculate available resources
	freeMemory := totalMemory - usedMemory
	freeVCPU := totalVCPU - usedVCPU

	hasResources := freeMemory >= requiredMemory && freeVCPU >= requiredVCPU

	results <- ResourceCheckResult{LibvirtURI: libvirtURI, HasResources: hasResources, Error: nil}
}

func (l *LibvirtService) CheckServersForResources(libvirtURIs []string, requiredMemory uint64, requiredVCPU uint) []ResourceCheckResult {
	var wg sync.WaitGroup
	results := make(chan ResourceCheckResult, len(libvirtURIs))

	for _, uri := range libvirtURIs {
		wg.Add(1)
		go checkResources(uri, requiredMemory, requiredVCPU, &wg, results)
	}

	wg.Wait()
	close(results)

	var checkResults []ResourceCheckResult
	for result := range results {
		checkResults = append(checkResults, result)
	}

	return checkResults
}
