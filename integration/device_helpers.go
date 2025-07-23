package integration

import (
	"dmt/pkg/device"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

var (
	deviceCounter int
	counterMutex  sync.Mutex
	randSource    = rand.New(rand.NewSource(time.Now().UnixNano()))
	randMutex     sync.Mutex
)

func getNextDeviceNumber() int {
	counterMutex.Lock()
	defer counterMutex.Unlock()
	deviceCounter++
	return deviceCounter
}

func generateRandomIP() string {
	randMutex.Lock()
	defer randMutex.Unlock()

	first := randSource.Intn(254) + 1
	second := randSource.Intn(254) + 1
	thrid := randSource.Intn(254) + 1
	return fmt.Sprintf("192.%d.%d.%d", first, second, thrid)
}

func generateRandomMAC() string {
	randMutex.Lock()
	defer randMutex.Unlock()

	return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x",
		0x02|(randSource.Intn(253)),
		randSource.Intn(256),
		randSource.Intn(256),
		randSource.Intn(256),
		randSource.Intn(256),
		randSource.Intn(256))
}

type DeviceOption func(*deviceConfig)

type deviceConfig struct {
	name        string
	deviceType  string
	ip          string
	mac         string
	description string
	employee    string
}

func withName(name string) DeviceOption {
	return func(c *deviceConfig) { c.name = name }
}

func withType(deviceType string) DeviceOption {
	return func(c *deviceConfig) { c.deviceType = deviceType }
}

func withIP(ip string) DeviceOption {
	return func(c *deviceConfig) { c.ip = ip }
}

func withMAC(mac string) DeviceOption {
	return func(c *deviceConfig) { c.mac = mac }
}

func withDescription(description string) DeviceOption {
	return func(c *deviceConfig) { c.description = description }
}

func withEmployee(employee string) DeviceOption {
	return func(c *deviceConfig) { c.employee = employee }
}

func createTestDeviceData(opts ...DeviceOption) map[string]interface{} {
	config := applyDeviceOptions(opts...)

	data := map[string]interface{}{
		"name": config.name,
		"type": config.deviceType,
		"ip":   config.ip,
		"mac":  config.mac,
	}

	if config.description != "" {
		data["description"] = config.description
	}

	if config.employee != "" {
		data["employee"] = config.employee
	}

	return data
}

func createTestDevice(opts ...DeviceOption) *device.Device {
	config := applyDeviceOptions(opts...)

	var description *string
	if config.description != "" {
		description = &config.description
	}

	var employee *string
	if config.employee != "" {
		employee = &config.employee
	}

	return &device.Device{
		Name:        config.name,
		Type:        config.deviceType,
		IP:          config.ip,
		MAC:         config.mac,
		Description: description,
		Employee:    employee,
	}
}

func applyDeviceOptions(opts ...DeviceOption) *deviceConfig {
	deviceNum := getNextDeviceNumber()

	config := &deviceConfig{
		name:       fmt.Sprintf("Test Device %d", deviceNum),
		deviceType: "laptop",
		ip:         generateRandomIP(),
		mac:        generateRandomMAC(),
	}

	for _, opt := range opts {
		opt(config)
	}

	return config
}

func createTestDevicesForEmployee(count int, employee string, baseOpts ...DeviceOption) []*device.Device {
	var devices []*device.Device

	for i := 0; i < count; i++ {
		opts := []DeviceOption{
			withEmployee(employee),
			withDescription(fmt.Sprintf("Test device %d for %s", i+1, employee)),
		}

		allOpts := append(opts, baseOpts...)
		devices = append(devices, createTestDevice(allOpts...))
	}

	return devices
}

func stringPtr(s string) *string {
	return &s
}
