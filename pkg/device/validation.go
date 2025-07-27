package device

import (
	"errors"
	"net"
	"regexp"
	"strings"
)

func validateName(name string) error {
	if name == "" {
		return errors.New("name is required")
	}
	if len(name) > 255 {
		return errors.New("name must be less than 255 characters")
	}
	return nil
}

func validateType(deviceType string) error {
	switch deviceType {
	case "desktop", "laptop", "phone", "tablet":
		return nil
	default:
		return errors.New("invalid device type")
	}
}

func validateIP(ip string) error {
	if net.ParseIP(ip) == nil {
		return errors.New("invalid IP address")
	}

	return nil
}

func validateMAC(mac string) error {
	mac = strings.ToLower(mac)
	macRegex := regexp.MustCompile(`^([0-9a-f]{2}[:-]){5}([0-9a-f]{2})$`)
	if !macRegex.MatchString(mac) {
		return errors.New("invalid MAC address format")
	}

	return nil
}

func validateDescription(description *string) error {
	if description != nil && len(*description) > 500 {
		return errors.New("description must be less than 500 characters")
	}
	return nil
}

func validateEmployee(employee *string) error {
	if employee != nil && len(*employee) != 3 {
		return errors.New("employee must be 3 characters")
	}
	return nil
}

func validateDevice(device *Device) []error {
	errors := make([]error, 0, 5)

	if err := validateName(device.Name); err != nil {
		errors = append(errors, err)
	}

	if err := validateType(device.Type); err != nil {
		errors = append(errors, err)
	}

	if err := validateMAC(device.MAC); err != nil {
		errors = append(errors, err)
	}

	if err := validateDescription(device.Description); err != nil {
		errors = append(errors, err)
	}

	if err := validateEmployee(device.Employee); err != nil {
		errors = append(errors, err)
	}

	return errors
}

func sanitizeDevice(device *Device) {
	device.Name = strings.TrimSpace(device.Name)
	device.Type = strings.TrimSpace(device.Type)

	mac := strings.TrimSpace(device.MAC)
	mac = strings.ToLower(mac)
	mac = strings.ReplaceAll(mac, "-", ":")
	device.MAC = mac

	if device.Description != nil {
		*device.Description = strings.TrimSpace(*device.Description)
		if *device.Description == "" {
			device.Description = nil
		}
	}

	if device.Employee != nil {
		*device.Employee = strings.TrimSpace(*device.Employee)
		if *device.Employee == "" {
			device.Employee = nil
		}
	}
}
