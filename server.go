package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"path"
	"time"
	"strconv"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

const (
	defaultHealthCheckIntervalSeconds = time.Duration(60)
)

// GPEmuDevicePlugin implements the Kubernetes device plugin API
type GPEmuDevicePluginConfig struct {
	ResourceName               string        `json:"resourceName"`
	SocketName                 string        `json:"socketName"`
	EGPUs                []*EGPU `json:"eGPUs"`
	HealthCheckIntervalSeconds time.Duration `json:"healthCheckIntervalSeconds"`
}

type GPEmuDevicePlugin struct {
	resourceName               string
	socket                     string
	healthCheckIntervalSeconds time.Duration
	devs                       []*pluginapi.Device

	stop   chan interface{}
	health chan string

	// this device files will be mounted to container
	eGPUs []*ExpandedEGPU

	server *grpc.Server
}

var (
	_ pluginapi.DevicePluginServer = &GPEmuDevicePlugin{}
)

// NewGPEmuDevicePlugin returns an initialized GPEmuDevicePlugin
func NewGPEmuDevicePlugin(config GPEmuDevicePluginConfig) (*GPEmuDevicePlugin, error) {
	expandedEGPUs := []*ExpandedEGPU{}
	for _, hd := range config.EGPUs {
		expanded, err := hd.Expand()
		if err != nil {
			return nil, err
		}
		expandedEGPUs = append(expandedEGPUs, expanded...)
	}

	log.Println("In NewGPEmuDevicePlugin, expandedEGPUs: ", expandedEGPUs)

	log.Println("len(expandedEGPUs): ", len(expandedEGPUs))

	var devs = make([]*pluginapi.Device, len(expandedEGPUs))

	for i, _ := range devs {
		devs[i] = &pluginapi.Device{
			ID:     fmt.Sprint(i),
			Health: pluginapi.Healthy,
		}
	}

	log.Println("devs: ", devs)


	healthCheckIntervalSeconds := defaultHealthCheckIntervalSeconds
	if config.HealthCheckIntervalSeconds > 0 {
		healthCheckIntervalSeconds = config.HealthCheckIntervalSeconds
	}

	return &GPEmuDevicePlugin{
		resourceName:               config.ResourceName,
		socket:                     pluginapi.DevicePluginPath + config.SocketName,
		healthCheckIntervalSeconds: healthCheckIntervalSeconds,

		devs:        devs,
		eGPUs: expandedEGPUs,

		stop:   make(chan interface{}),
		health: make(chan string),
	}, nil
}

// dial establishes the gRPC communication with the registered device plugin.
func dial(unixSocketPath string, timeout time.Duration) (*grpc.ClientConn, error) {
	c, err := grpc.Dial(unixSocketPath, grpc.WithInsecure(), grpc.WithBlock(),
		grpc.WithTimeout(timeout),
		grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) {
			return net.DialTimeout("unix", addr, timeout)
		}),
	)

	if err != nil {
		return nil, err
	}

	return c, nil
}


// Start starts the gRPC server of the device plugin
func (m *GPEmuDevicePlugin) Start() error {
	err := m.cleanup()
	if err != nil {
		return err
	}

	sock, err := net.Listen("unix", m.socket)
	if err != nil {
		return err
	}
	log.Println("socket: ", m.socket)

	m.server = grpc.NewServer([]grpc.ServerOption{}...)
	pluginapi.RegisterDevicePluginServer(m.server, m)

	go m.server.Serve(sock)

	// Wait for server to start by launching a blocking connexion
	conn, err := dial(m.socket, 5*time.Second)
	if err != nil {
		return err
	}
	conn.Close()

	go m.healthCheck()

	return nil
}

// Stop stops the gRPC server
func (m *GPEmuDevicePlugin) Stop() error {
	if m.server == nil {
		return nil
	}

	m.server.Stop()
	m.server = nil
	close(m.stop)

	return m.cleanup()
}

// Register registers the device plugin for the given resourceName with Kubelet.
func (m *GPEmuDevicePlugin) Register(kubeletEndpoint, resourceName string) error {
	conn, err := dial(kubeletEndpoint, 5*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()

	client := pluginapi.NewRegistrationClient(conn)
	reqt := &pluginapi.RegisterRequest{
		Version:      pluginapi.Version,
		Endpoint:     path.Base(m.socket),
		ResourceName: resourceName,
	}

	_, err = client.Register(context.Background(), reqt)
	if err != nil {
		return err
	}
	return nil
}

// ListAndWatch lists devices and update that list according to the health status
func (m *GPEmuDevicePlugin) ListAndWatch(e *pluginapi.Empty, s pluginapi.DevicePlugin_ListAndWatchServer) error {
	fmt.Println("exposing devices: ", m.devs)
	s.Send(&pluginapi.ListAndWatchResponse{Devices: m.devs})

	for {
		select {
		case <-m.stop:
			return nil
		case health := <-m.health:
			// Update health of devices only in this thread.
			for _, dev := range m.devs {
				dev.Health = health
			}
			s.Send(&pluginapi.ListAndWatchResponse{Devices: m.devs})
		}
	}
}

func (m *GPEmuDevicePlugin) healthCheck() {
	log.Printf("Starting health check every %d seconds", m.healthCheckIntervalSeconds)
	ticker := time.NewTicker(m.healthCheckIntervalSeconds * time.Second)
	lastHealth := pluginapi.Healthy
	m.health <- lastHealth
	for {
		select {
		case <-ticker.C:
			health := pluginapi.Healthy
			if lastHealth != health {
				log.Printf("Health is changed: %s -> %s", lastHealth, health)
				m.health <- health
			}
			lastHealth = health
		case <-m.stop:
			ticker.Stop()
			return
		}
	}
}

// Allocate which return list of devices.
func (m *GPEmuDevicePlugin) Allocate(ctx context.Context, r *pluginapi.AllocateRequest) (*pluginapi.AllocateResponse, error) {
	log.Println("allocate request:", r)

	ress := make([]*pluginapi.ContainerAllocateResponse, len(r.GetContainerRequests()))

	for i, request := range r.GetContainerRequests() {
		log.Println("request.DevicesIDs: ", request.DevicesIDs)
		ds := make([]*pluginapi.Mount, len(request.DevicesIDs))
		log.Println("m.eGPUs: len:", len(m.eGPUs), "value: ", m.eGPUs)
		for j, deviceID := range request.DevicesIDs {
			deviceIDInt, _ := strconv.Atoi(deviceID)
			log.Println(deviceIDInt)
			// ds[j] = &pluginapi.DeviceSpec{
			// 	HostPath:      m.eGPUs[deviceIDInt].HostPath,
			// 	ContainerPath: m.eGPUs[deviceIDInt].ContainerPath,
			// 	Permissions:   m.eGPUs[deviceIDInt].Permission,
			// }
			ds[j] = &pluginapi.Mount{
				HostPath:      m.eGPUs[deviceIDInt].HostPath,
				ContainerPath: m.eGPUs[deviceIDInt].ContainerPath,
				ReadOnly:      false,
			}
		}
		ress[i] = &pluginapi.ContainerAllocateResponse{
			// Devices: ds,
			Mounts: ds,
		}
		log.Println("ds: ", ds)
	}

	response := pluginapi.AllocateResponse{
		ContainerResponses: ress,
	}

	log.Println("allocate response: ", response)
	return &response, nil
}

func (m *GPEmuDevicePlugin) GetDevicePluginOptions(context.Context, *pluginapi.Empty) (*pluginapi.DevicePluginOptions, error) {
	return &pluginapi.DevicePluginOptions{
		PreStartRequired: false,
	}, nil
}

func (m *GPEmuDevicePlugin) PreStartContainer(context.Context, *pluginapi.PreStartContainerRequest) (*pluginapi.PreStartContainerResponse, error) {
	return &pluginapi.PreStartContainerResponse{}, nil
}

func (m *GPEmuDevicePlugin) GetPreferredAllocation(context.Context, *pluginapi.PreferredAllocationRequest) (*pluginapi.PreferredAllocationResponse, error) {
	return &pluginapi.PreferredAllocationResponse{}, nil
}

func (m *GPEmuDevicePlugin) cleanup() error {
	if err := os.Remove(m.socket); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

// Serve starts the gRPC server and register the device plugin to Kubelet
func (m *GPEmuDevicePlugin) Serve() error {
	err := m.Start()
	if err != nil {
		log.Printf("Could not start device plugin: %s", err)
		return err
	}
	log.Println("Starting to serve on", m.socket)

	err = m.Register(pluginapi.KubeletSocket, m.resourceName)
	if err != nil {
		log.Printf("Could not register device plugin: %s", err)
		m.Stop()
		return err
	}
	log.Println("Registered device plugin with Kubelet")

	return nil
}
