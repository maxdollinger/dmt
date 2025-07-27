package integration

import (
	"bytes"
	"context"
	"dmt/pkg/device"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNotifyDeviceCount(t *testing.T) {
	ctx := t.Context()

	testDB, err := NewTestDB(ctx)
	require.NoError(t, err)
	defer testDB.Terminate()

	notificationContainer, err := NewNotificationContainer(ctx)
	require.NoError(t, err)
	defer notificationContainer.Terminate()

	db, err := testDB.GetConnectionPool()
	require.NoError(t, err)

	t.Run("Send http Notification", func(t *testing.T) {
		client := &http.Client{
			Timeout: 10 * time.Second,
		}

		res, err := client.Post(notificationContainer.GetNotificationURL(), "application/json", bytes.NewBuffer([]byte(`{"level":"warning","employeeAbbreviation":"jdo","message":"Device count warning: Employee jdo has 3 devices"}`)))
		assert.NoError(t, err)

		body, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Contains(t, string(body), "jdo")

		err = notificationContainer.WaitForLog("Notification successfully received", 10*time.Second)
		assert.NoError(t, err)

		err = notificationContainer.WaitForLog("jdo has 3 devices", 10*time.Second)
		assert.NoError(t, err)
	})

	t.Run("Notification Triggered When Employee Has 3+ Devices", func(t *testing.T) {
		defer testDB.ClearDB(t)

		ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
		defer cancel()

		err := device.HandleDeviceCountNotifications(ctx, db, notificationContainer.GetNotificationURL())
		assert.NoError(t, err)

		testDevices := createTestDevicesForEmployee(3, "jdo")
		for _, testDevice := range testDevices {
			err := device.InsertDevice(ctx, db, testDevice)
			require.NoError(t, err)
		}

		err = notificationContainer.WaitForLog("Notification successfully received", 10*time.Second)
		assert.NoError(t, err)

		err = notificationContainer.WaitForLog("jdo has 3 devices", 10*time.Second)
		assert.NoError(t, err)

		fourthDevice := createTestDevice(withEmployee("jsm"))
		err = device.InsertDevice(context.Background(), db, fourthDevice)
		require.NoError(t, err)

		fourthDevice.Employee = stringPtr("jdo")
		device.UpdateDevice(context.Background(), db, fourthDevice)

		err = notificationContainer.WaitForLog("jdo has 4 devices", 10*time.Second)
		assert.NoError(t, err)

		cancel()
	})
}
