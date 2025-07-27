package integration

import (
	"dmt/pkg/device"
	"fmt"
	"math/rand"
	"net"
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

func generateRandomIP() net.IP {
	randMutex.Lock()
	defer randMutex.Unlock()

	first := randSource.Intn(254) + 1
	second := randSource.Intn(254) + 1
	thrid := randSource.Intn(254) + 1
	return net.IPv4(byte(192), byte(first), byte(second), byte(thrid))
}

func generateRandomMAC() net.HardwareAddr {
	randMutex.Lock()
	defer randMutex.Unlock()

	addrString := fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x",
		0x02|(randSource.Intn(253)),
		randSource.Intn(256),
		randSource.Intn(256),
		randSource.Intn(256),
		randSource.Intn(256),
		randSource.Intn(256))

	mac, err := net.ParseMAC(addrString)
	if err != nil {
		panic(err)
	}

	return mac
}

type DeviceOption func(*device.Device)

func withName(name string) DeviceOption {
	return func(d *device.Device) { d.Name = name }
}

func withType(deviceType string) DeviceOption {
	return func(d *device.Device) { d.Type = deviceType }
}

func withIP(ip string) DeviceOption {
	return func(d *device.Device) { d.IP = net.ParseIP(ip) }
}

func withMAC(mac string) DeviceOption {
	addr, err := net.ParseMAC(mac)
	if err != nil {
		panic(err)
	}
	return func(d *device.Device) { d.MAC = addr }
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
