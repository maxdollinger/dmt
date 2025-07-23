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

type DeviceOption func(*device.Device)

func withName(name string) DeviceOption {
	return func(d *device.Device) { d.Name = name }
}

func withType(deviceType string) DeviceOption {
	return func(d *device.Device) { d.Type = deviceType }
}

func withIP(ip string) DeviceOption {
	return func(d *device.Device) { d.IP = ip }
}

func withMAC(mac string) DeviceOption {
	return func(d *device.Device) { d.MAC = mac }
}

func withEmployee(employee string) DeviceOption {
	return func(d *device.Device) { d.Employee = &employee }
}

func createTestDevice(opts ...DeviceOption) *device.Device {
	device := &device.Device{
		Name: "Test Device",
		Type: "laptop",
		IP:   generateRandomIP(),
		MAC:  generateRandomMAC(),
	}
	for _, opt := range opts {
		opt(device)
	}

	return device
}

func createTestDevicesForEmployee(count int, employee string, baseOpts ...DeviceOption) []*device.Device {
	var devices []*device.Device

	for i := 0; i < count; i++ {
		opts := []DeviceOption{
			withEmployee(employee),
		}

		allOpts := append(opts, baseOpts...)
		devices = append(devices, createTestDevice(allOpts...))
	}

	return devices
}

func stringPtr(s string) *string {
	return &s
}
