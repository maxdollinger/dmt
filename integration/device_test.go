package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeviceAPI(t *testing.T) {
	// Setup test database container
	testDB := SetupTestDB(t)
	defer testDB.Cleanup(t)

	t.Run("Create and Get Device", func(t *testing.T) {
		app := testDB.CreateApp(t)

		deviceData := map[string]interface{}{
			"name": "Get Test Device",
			"type": "phone",
			"ip":   "192.168.1.101",
			"mac":  "bb:cc:dd:ee:ff:aa",
		}

		jsonData, err := json.Marshal(deviceData)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/devices", bytes.NewBuffer(jsonData))
		req.Header.Set(GetAuthHeader())
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, 5000)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var createResponse map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&createResponse)
		require.NoError(t, err)

		deviceMap := createResponse["device"].(map[string]interface{})
		deviceID := int(deviceMap["id"].(float64))

		getReq := httptest.NewRequest("GET", fmt.Sprintf("/devices/%d", deviceID), nil)
		getReq.Header.Set(GetAuthHeader())

		getResp, err := app.Test(getReq, 5000)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, getResp.StatusCode)

		var getResponse map[string]interface{}
		err = json.NewDecoder(getResp.Body).Decode(&getResponse)
		require.NoError(t, err)

		assert.Equal(t, deviceData["name"], getResponse["name"])
		assert.Equal(t, deviceData["type"], getResponse["type"])
		assert.Equal(t, deviceData["ip"], getResponse["ip"])
		assert.Equal(t, deviceData["mac"], getResponse["mac"])
	})

	t.Run("Get Non-existent Device", func(t *testing.T) {
		app := testDB.CreateApp(t)

		req := httptest.NewRequest("GET", "/devices/99999", nil)
		req.Header.Set(GetAuthHeader())

		resp, err := app.Test(req, 5000)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("Unauthorized Request", func(t *testing.T) {
		app := testDB.CreateApp(t)

		req := httptest.NewRequest("GET", "/devices/1", nil)

		resp, err := app.Test(req, 5000)
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

// BenchmarkDeviceCreation benchmarks device creation performance
func BenchmarkDeviceCreation(b *testing.B) {
	t := &testing.T{}
	testDB := SetupTestDB(t)
	defer testDB.Cleanup(t)

	app := testDB.CreateApp(&testing.T{})

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
		req.Header.Set(GetAuthHeader())
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
