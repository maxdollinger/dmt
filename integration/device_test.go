package integration

import (
	"bytes"
	"context"
	"dmt/internal"
	"dmt/pkg/device"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeviceAPI(t *testing.T) {
	ctx := t.Context()
	testDB, err := NewTestDB(ctx)
	require.NoError(t, err)
	defer testDB.Terminate()

	db, err := testDB.GetConnectionPool()
	require.NoError(t, err)

	app := internal.CreateHttpServer(db, testAPIKey)

	t.Run("Create and Get Device", func(t *testing.T) {
		defer testDB.ClearDB(t)
		ctx := t.Context()

		deviceData := createTestDevice()
		jsonData, err := json.Marshal(deviceData)
		require.NoError(t, err)

		req := JSONRequestWithApiKey("POST", "/api/v1/devices", jsonData)
		responseBody := make(map[string]interface{})
		makeRequest(t, app, req, http.StatusCreated, &responseBody)

		createdDevice := &device.Device{ID: int(responseBody["device"].(map[string]interface{})["id"].(float64))}
		err = device.GetDeviceByID(ctx, db, createdDevice)
		require.NoError(t, err)

		assert.Equal(t, deviceData.Name, createdDevice.Name)
		assert.Equal(t, deviceData.Type, createdDevice.Type)
		assert.Equal(t, deviceData.IP, createdDevice.IP)
		assert.Equal(t, deviceData.MAC, createdDevice.MAC, "Expected MAC to be %s but got %s", deviceData.MAC.String(), createdDevice.MAC.String())
	})

	t.Run("Get Non-existent Device", func(t *testing.T) {
		req := JSONRequestWithApiKey("GET", "/api/v1/devices/9999999", nil)
		makeRequest(t, app, req, http.StatusNotFound, nil)
	})

	t.Run("Unauthorized Request - missing apikey", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/devices/1", nil)
		makeRequest(t, app, req, http.StatusUnauthorized, nil)
	})

	t.Run("Unauthorized Request - invalid apikey", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/devices/1", nil)
		req.Header.Set("Authorization", "Bearer invalidKey")
		makeRequest(t, app, req, http.StatusUnauthorized, nil)
	})

	t.Run("Delete Device", func(t *testing.T) {
		defer testDB.ClearDB(t)
		ctx := t.Context()

		testDevice := createTestDevice()

		err := device.InsertDevice(ctx, db, testDevice)
		require.NoError(t, err)
		require.NotZero(t, testDevice.ID)

		deleteReq := JSONRequestWithApiKey("DELETE", fmt.Sprintf("/api/v1/devices/%d", testDevice.ID), nil)
		makeRequest(t, app, deleteReq, http.StatusOK, nil)

		err = device.GetDeviceByID(ctx, db, testDevice)
		require.Error(t, err)
	})

	t.Run("Delete Non-existent Device", func(t *testing.T) {
		deleteReq := JSONRequestWithApiKey("DELETE", "/api/v1/devices/9999999", nil)
		makeRequest(t, app, deleteReq, http.StatusOK, nil)
	})

	t.Run("Update Device Employee", func(t *testing.T) {
		defer testDB.ClearDB(t)
		ctx := t.Context()

		testDevice := createTestDevice(withEmployee("jsm"))

		err := device.InsertDevice(ctx, db, testDevice)
		require.NoError(t, err, "Failed to insert device")
		require.NotZero(t, testDevice.ID)

		employeeData := map[string]interface{}{
			"employee": "jdo",
		}
		employeeJsonData, err := json.Marshal(employeeData)
		require.NoError(t, err)

		updateReq := JSONRequestWithApiKey("PUT", fmt.Sprintf("/api/v1/devices/%d/employee", testDevice.ID), employeeJsonData)
		makeRequest(t, app, updateReq, http.StatusOK, nil)

		updatedDevice := &device.Device{ID: testDevice.ID}
		err = device.GetDeviceByID(ctx, db, updatedDevice)
		require.NoError(t, err)
		assert.Equal(t, employeeData["employee"], *updatedDevice.Employee, "Expected device employee be the same but is not")
		assert.Equal(t, testDevice.Name, updatedDevice.Name)
		assert.Equal(t, testDevice.Type, updatedDevice.Type)
		assert.Equal(t, testDevice.IP, updatedDevice.IP)
		assert.Equal(t, testDevice.MAC, updatedDevice.MAC)
	})

	t.Run("Remove Device Employee", func(t *testing.T) {
		defer testDB.ClearDB(t)
		ctx := t.Context()

		testDevice := createTestDevice(withEmployee("jdo"))
		err := device.InsertDevice(ctx, db, testDevice)
		require.NoError(t, err)
		require.NotZero(t, testDevice.ID)

		deleteReq := JSONRequestWithApiKey("DELETE", fmt.Sprintf("/api/v1/devices/%d/employee", testDevice.ID), nil)
		makeRequest(t, app, deleteReq, http.StatusOK, nil)

		updatedDevice := &device.Device{ID: testDevice.ID}
		err = device.GetDeviceByID(context.Background(), db, updatedDevice)
		require.NoError(t, err)
		assert.Nil(t, updatedDevice.Employee, "Expected device to have no employee but employee field is not nil")
		assert.Equal(t, testDevice.Name, updatedDevice.Name, "Expected device name be the same but is not")
		assert.Equal(t, testDevice.Type, updatedDevice.Type, "Expected device type be the same but is not")
		assert.Equal(t, testDevice.IP, updatedDevice.IP, "Expected device IP be the same but is not")
		assert.Equal(t, testDevice.MAC, updatedDevice.MAC, "Expected device MAC be the same but is not")
	})

	t.Run("Get Devices with Filters", func(t *testing.T) {
		defer testDB.ClearDB(t)
		ctx := t.Context()

		testDevices := []*device.Device{
			createTestDevice(
				withName("Device 1"),
				withType("laptop"),
				withIP("192.168.1.100"),
				withMAC("aa:bb:cc:dd:ee:ff"),
				withEmployee("jdo"),
			),
			createTestDevice(
				withName("Device 2"),
				withType("phone"),
				withIP("192.168.2.101"),
				withMAC("bb:cc:dd:ee:ff:aa"),
				withEmployee("jsm"),
			),
			createTestDevice(
				withName("Device 3"),
				withType("laptop"),
				withIP("10.0.1.102"),
				withMAC("cc:dd:ee:ff:aa:bb"),
				withEmployee("jdo"),
			),
		}

		for _, testDevice := range testDevices {
			err := device.InsertDevice(ctx, db, testDevice)
			require.NoError(t, err)
			require.NotZero(t, testDevice.ID)
		}

		// Test 1: Get all devices (no filters)
		getAllReq := httptest.NewRequest("GET", "/api/v1/devices", nil)
		SetAuthHeader(getAllReq)

		var getAllResponse map[string]interface{}
		makeRequest(t, app, getAllReq, http.StatusOK, &getAllResponse)
		assert.Equal(t, float64(3), getAllResponse["count"], "Expected 3 devices but got %0.f", getAllResponse["count"])

		// Test 2: Filter by employee (exact match)
		getByEmployeeReq := httptest.NewRequest("GET", "/api/v1/devices?employee=jdo", nil)
		SetAuthHeader(getByEmployeeReq)

		var getByEmployeeResponse map[string]interface{}
		makeRequest(t, app, getByEmployeeReq, http.StatusOK, &getByEmployeeResponse)
		assert.Equal(t, float64(2), getByEmployeeResponse["count"], "Expected 2 devices with employee jdo but got %0.f", getByEmployeeResponse["count"])

		// Test 3: Filter by type (exact match)
		getByTypeReq := httptest.NewRequest("GET", "/api/v1/devices?type=phone", nil)
		SetAuthHeader(getByTypeReq)

		var getByTypeResponse map[string]interface{}
		makeRequest(t, app, getByTypeReq, http.StatusOK, &getByTypeResponse)
		assert.Equal(t, float64(1), getByTypeResponse["count"], "Expected 1 device with type phone but got %0.f", getByTypeResponse["count"])

		// Test 4: Filter by IP (like search)
		getByIpReq := httptest.NewRequest("GET", "/api/v1/devices?ip=192.168", nil)
		SetAuthHeader(getByIpReq)

		var getByIpResponse map[string]interface{}
		makeRequest(t, app, getByIpReq, http.StatusOK, &getByIpResponse)
		assert.Equal(t, float64(2), getByIpResponse["count"], "Expected 2 devices with IP 192.168 but got %0.f", getByIpResponse["count"])

		// Test 5: Filter by MAC (like search)
		getByMacReq := httptest.NewRequest("GET", "/api/v1/devices?mac=aa:bb", nil)
		SetAuthHeader(getByMacReq)

		var getByMacResponse map[string]interface{}
		makeRequest(t, app, getByMacReq, http.StatusOK, &getByMacResponse)
		assert.Equal(t, float64(2), getByMacResponse["count"], "Expected 1 device with MAC aa:bb but got %0.f", getByMacResponse["count"])

		// Test 6: Combine multiple filters
		getCombinedReq := httptest.NewRequest("GET", "/api/v1/devices?employee=jdo&type=laptop", nil)
		SetAuthHeader(getCombinedReq)

		var getCombinedResponse map[string]interface{}
		makeRequest(t, app, getCombinedReq, http.StatusOK, &getCombinedResponse)
		assert.Equal(t, float64(2), getCombinedResponse["count"], "Expected 2 devices with employee jdo and type laptop but got %0.f", getCombinedResponse["count"])
	})
}

func makeRequest(t *testing.T, app *fiber.App, req *http.Request, expectedStatus int, response interface{}) {
	resp, err := app.Test(req, 5000)
	require.NoError(t, err)
	require.Equal(t, expectedStatus, resp.StatusCode, "Expected status code %d but got %d", expectedStatus, resp.StatusCode)

	if response != nil && expectedStatus < 300 {
		err = json.NewDecoder(resp.Body).Decode(response)
		require.NoError(t, err)
	}
}

func BenchmarkDeviceCreation(b *testing.B) {
	ctx := context.Background()
	testDB, err := NewTestDB(ctx)
	require.NoError(b, err)
	defer testDB.Terminate()

	db, err := testDB.GetConnectionPool()
	require.NoError(b, err)

	app := internal.CreateHttpServer(db, testAPIKey)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		deviceData := createTestDevice()

		jsonData, _ := json.Marshal(deviceData)
		req := httptest.NewRequest("POST", "/api/v1/devices", bytes.NewBuffer(jsonData))
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
