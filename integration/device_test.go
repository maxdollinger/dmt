package integration

import (
	"bytes"
	"context"
	"dmt/pkg/device"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeviceAPI(t *testing.T) {
	testDB := SetupTestDB(t)
	defer testDB.Cleanup(t)

	t.Run("Create and Get Device", func(t *testing.T) {
		defer testDB.ClearDB(t)
		app, db := testDB.CreateApp(t)

		deviceData := map[string]interface{}{
			"name": "Get Test Device",
			"type": "phone",
			"ip":   "192.168.1.101",
			"mac":  "bb:cc:dd:ee:ff:aa",
		}

		jsonData, err := json.Marshal(deviceData)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/devices", bytes.NewBuffer(jsonData))
		SetAuthHeader(req)
		req.Header.Set("Content-Type", "application/json")

		var createResponse map[string]interface{}
		makeRequest(t, app, req, http.StatusCreated, &createResponse)

		deviceMap := createResponse["device"].(map[string]interface{})
		deviceID := int(deviceMap["id"].(float64))

		dbDevice := &device.Device{ID: deviceID}
		err = device.GetDeviceByID(context.Background(), db, dbDevice)
		require.NoError(t, err)

		assert.Equal(t, deviceData["name"], dbDevice.Name)
		assert.Equal(t, deviceData["type"], dbDevice.Type)
		assert.Equal(t, deviceData["ip"], dbDevice.IP)
		assert.Equal(t, deviceData["mac"], dbDevice.MAC)
	})

	t.Run("Get Non-existent Device", func(t *testing.T) {
		defer testDB.ClearDB(t)
		app, _ := testDB.CreateApp(t)

		req := httptest.NewRequest("GET", "/devices/99999", nil)
		SetAuthHeader(req)

		makeRequest(t, app, req, http.StatusNotFound, nil)
	})

	t.Run("Unauthorized Request", func(t *testing.T) {
		defer testDB.ClearDB(t)
		app, _ := testDB.CreateApp(t)

		req := httptest.NewRequest("GET", "/devices/1", nil)

		makeRequest(t, app, req, http.StatusUnauthorized, nil)
	})

	t.Run("Delete Device", func(t *testing.T) {
		defer testDB.ClearDB(t)
		app, db := testDB.CreateApp(t)

		testDevice := &device.Device{
			Name: "Delete Test Device",
			Type: "tablet",
			IP:   "192.168.1.102",
			MAC:  "aa:bb:cc:dd:ee:ff",
		}

		err := device.InsertDevice(context.Background(), db, testDevice)
		require.NoError(t, err)
		require.NotZero(t, testDevice.ID)

		deleteReq := httptest.NewRequest("DELETE", fmt.Sprintf("/devices/%d", testDevice.ID), nil)
		SetAuthHeader(deleteReq)

		var deleteResponse map[string]interface{}
		makeRequest(t, app, deleteReq, http.StatusOK, &deleteResponse)

		verifyDevice := &device.Device{ID: testDevice.ID}
		err = device.GetDeviceByID(context.Background(), db, verifyDevice)
		require.Error(t, err)
	})

	t.Run("Delete Non-existent Device", func(t *testing.T) {
		defer testDB.ClearDB(t)
		app, _ := testDB.CreateApp(t)

		deleteReq := httptest.NewRequest("DELETE", "/devices/99999", nil)
		SetAuthHeader(deleteReq)

		var deleteResponse map[string]interface{}
		makeRequest(t, app, deleteReq, http.StatusOK, &deleteResponse)
	})

	t.Run("Update Device", func(t *testing.T) {
		defer testDB.ClearDB(t)
		app, db := testDB.CreateApp(t)

		testDevice := &device.Device{
			Name: "Update Test Device",
			Type: "laptop",
			IP:   "192.168.1.103",
			MAC:  "cc:dd:ee:ff:aa:bb",
		}

		err := device.InsertDevice(context.Background(), db, testDevice)
		require.NoError(t, err)
		require.NotZero(t, testDevice.ID)

		updatedDeviceData := map[string]interface{}{
			"name":        "Updated Test Device",
			"type":        "desktop",
			"ip":          "192.168.1.104",
			"mac":         "dd:ee:ff:aa:bb:cc",
			"description": "Updated device for testing",
			"employee":    "jdo",
		}

		updatedJsonData, err := json.Marshal(updatedDeviceData)
		require.NoError(t, err)

		updateReq := httptest.NewRequest("PUT", fmt.Sprintf("/devices/%d", testDevice.ID), bytes.NewBuffer(updatedJsonData))
		SetAuthHeader(updateReq)
		updateReq.Header.Set("Content-Type", "application/json")

		var updateResponse map[string]interface{}
		makeRequest(t, app, updateReq, http.StatusOK, &updateResponse)

		dbDevice := &device.Device{ID: testDevice.ID}
		err = device.GetDeviceByID(context.Background(), db, dbDevice)
		require.NoError(t, err)
		assert.Equal(t, updatedDeviceData["name"], dbDevice.Name)
		assert.Equal(t, updatedDeviceData["type"], dbDevice.Type)
		assert.Equal(t, updatedDeviceData["ip"], dbDevice.IP)
		assert.Equal(t, updatedDeviceData["mac"], dbDevice.MAC)
		assert.Equal(t, updatedDeviceData["description"], *dbDevice.Description)
		assert.Equal(t, updatedDeviceData["employee"], *dbDevice.Employee)
	})

	t.Run("Get Devices with Filters", func(t *testing.T) {
		defer testDB.ClearDB(t)
		app, db := testDB.CreateApp(t)

		testDevices := []*device.Device{
			{
				Name:        "Device 1",
				Type:        "laptop",
				IP:          "192.168.1.100",
				MAC:         "aa:bb:cc:dd:ee:ff",
				Description: stringPtr("Test laptop"),
				Employee:    stringPtr("jdo"),
			},
			{
				Name:        "Device 2",
				Type:        "phone",
				IP:          "192.168.2.100",
				MAC:         "bb:cc:dd:ee:ff:aa",
				Description: stringPtr("Test phone"),
				Employee:    stringPtr("jsm"),
			},
			{
				Name:        "Device 3",
				Type:        "laptop",
				IP:          "10.0.1.100",
				MAC:         "cc:dd:ee:ff:aa:bb",
				Description: stringPtr("Another laptop"),
				Employee:    stringPtr("jdo"),
			},
		}

		for _, testDevice := range testDevices {
			err := device.InsertDevice(context.Background(), db, testDevice)
			require.NoError(t, err)
			require.NotZero(t, testDevice.ID)
		}

		// Test 1: Get all devices (no filters)
		getAllReq := httptest.NewRequest("GET", "/devices", nil)
		SetAuthHeader(getAllReq)

		var getAllResponse map[string]interface{}
		makeRequest(t, app, getAllReq, http.StatusOK, &getAllResponse)
		assert.Equal(t, float64(3), getAllResponse["count"], "Expected 3 devices but got %0.f", getAllResponse["count"])

		// Test 2: Filter by employee (exact match)
		getByEmployeeReq := httptest.NewRequest("GET", "/devices?employee=jdo", nil)
		SetAuthHeader(getByEmployeeReq)

		var getByEmployeeResponse map[string]interface{}
		makeRequest(t, app, getByEmployeeReq, http.StatusOK, &getByEmployeeResponse)
		assert.Equal(t, float64(2), getByEmployeeResponse["count"], "Expected 2 devices with employee jdo but got %0.f", getByEmployeeResponse["count"])

		// Test 3: Filter by type (exact match)
		getByTypeReq := httptest.NewRequest("GET", "/devices?type=phone", nil)
		SetAuthHeader(getByTypeReq)

		var getByTypeResponse map[string]interface{}
		makeRequest(t, app, getByTypeReq, http.StatusOK, &getByTypeResponse)
		assert.Equal(t, float64(1), getByTypeResponse["count"], "Expected 1 device with type phone but got %0.f", getByTypeResponse["count"])

		// Test 4: Filter by IP (like search)
		getByIpReq := httptest.NewRequest("GET", "/devices?ip=192.168", nil)
		SetAuthHeader(getByIpReq)

		var getByIpResponse map[string]interface{}
		makeRequest(t, app, getByIpReq, http.StatusOK, &getByIpResponse)
		assert.Equal(t, float64(2), getByIpResponse["count"], "Expected 2 devices with IP 192.168 but got %0.f", getByIpResponse["count"])

		// Test 5: Filter by MAC (like search)
		getByMacReq := httptest.NewRequest("GET", "/devices?mac=aa:bb", nil)
		SetAuthHeader(getByMacReq)

		var getByMacResponse map[string]interface{}
		makeRequest(t, app, getByMacReq, http.StatusOK, &getByMacResponse)
		assert.Equal(t, float64(2), getByMacResponse["count"], "Expected 1 device with MAC aa:bb but got %0.f", getByMacResponse["count"])

		// Test 6: Combine multiple filters
		getCombinedReq := httptest.NewRequest("GET", "/devices?employee=jdo&type=laptop", nil)
		SetAuthHeader(getCombinedReq)

		var getCombinedResponse map[string]interface{}
		makeRequest(t, app, getCombinedReq, http.StatusOK, &getCombinedResponse)
		assert.Equal(t, float64(2), getCombinedResponse["count"], "Expected 2 devices with employee jdo and type laptop but got %0.f", getCombinedResponse["count"])
	})
}

func makeRequest(t *testing.T, app *fiber.App, req *http.Request, expectedStatus int, response interface{}) {
	resp, err := app.Test(req, 5000)
	require.NoError(t, err)
	assert.Equal(t, expectedStatus, resp.StatusCode)

	if response != nil && expectedStatus < 300 {
		err = json.NewDecoder(resp.Body).Decode(response)
		require.NoError(t, err)
	}
}

func stringPtr(s string) *string {
	return &s
}

func BenchmarkDeviceCreation(b *testing.B) {
	t := &testing.T{}
	testDB := SetupTestDB(t)
	defer testDB.Cleanup(t)
	defer testDB.ClearDB(t)

	app, _ := testDB.CreateApp(&testing.T{})

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		deviceData := map[string]interface{}{
			"name": "Benchmark Device " + strconv.Itoa(i),
			"type": "laptop",
			"ip":   "192.168.1." + strconv.Itoa(100+i%50),
			"mac":  fmt.Sprintf("BB:BB:CC:DD:EE:%02X", i%256),
		}

		jsonData, _ := json.Marshal(deviceData)
		req := httptest.NewRequest("POST", "/devices", bytes.NewBuffer(jsonData))
		SetAuthHeader(req)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, 1000)
		if err != nil {
			b.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != http.StatusCreated {
			b.Fatalf("Expected 201, got %d", resp.StatusCode)
		}
	}
}
