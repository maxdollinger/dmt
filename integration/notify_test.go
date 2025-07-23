package integration

import (
	"context"
	"dmt/pkg/device"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNotifyDeviceCount(t *testing.T) {
	testDB := SetupTestDB(t)
	defer testDB.Cleanup(t)
	_, db := testDB.CreateApp(t)

	t.Run("Notification Triggered When Employee Has 3+ Devices", func(t *testing.T) {
		defer testDB.ClearDB(t)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		conn, err := db.Acquire(ctx)
		assert.NoError(t, err)

		notificationChan, err := device.DeviceCountListener(ctx, conn.Conn())
		assert.NoError(t, err)

		time.Sleep(100 * time.Millisecond)

		_, db := testDB.CreateApp(t)
		defer db.Close()

		testDevices := createTestDevicesForEmployee(2, "jdo")

		for _, testDevice := range testDevices {
			err := device.InsertDevice(context.Background(), db, testDevice)
			require.NoError(t, err)
		}

		select {
		case notification := <-notificationChan:
			t.Fatalf("Unexpected notification received: %+v", notification)
		case <-time.After(500 * time.Millisecond):
			// nothing should happen
		}

		thirdDevice := createTestDevice(withEmployee("jdo"))

		err = device.InsertDevice(context.Background(), db, thirdDevice)
		require.NoError(t, err)

		select {
		case notification := <-notificationChan:
			assert.Equal(t, "jdo", notification.Employee)
			assert.Equal(t, 3, notification.Count)
		case <-time.After(2 * time.Second):
			t.Fatal("Expected notification not received within timeout")
		}

		fourthDevice := createTestDevice(withEmployee("jdo"))

		err = device.InsertDevice(context.Background(), db, fourthDevice)
		require.NoError(t, err)

		select {
		case notification := <-notificationChan:
			assert.Equal(t, "jdo", notification.Employee)
			assert.Equal(t, 4, notification.Count)
		case <-time.After(2 * time.Second):
			t.Fatal("Expected second notification not received within timeout")
		}

		cancel()
	})

	t.Run("Multiple Employees Trigger Separate Notifications", func(t *testing.T) {
		defer testDB.ClearDB(t)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		conn, err := db.Acquire(ctx)
		assert.NoError(t, err)

		notificationChan, err := device.DeviceCountListener(ctx, conn.Conn())
		assert.NoError(t, err)

		time.Sleep(100 * time.Millisecond)

		_, db := testDB.CreateApp(t)
		defer db.Close()

		employees := []string{"ali", "bob"}
		receivedNotifications := make(map[string]int)

		for _, emp := range employees {
			devices := createTestDevicesForEmployee(3, emp)
			for _, testDevice := range devices {
				err := device.InsertDevice(context.Background(), db, testDevice)
				require.NoError(t, err)
			}
		}

		timeout := time.After(5 * time.Second)
		for len(receivedNotifications) < 2 {
			select {
			case notification := <-notificationChan:
				receivedNotifications[notification.Employee] = notification.Count
			case <-timeout:
				t.Fatal("Did not receive notifications for both employees within timeout")
			}
		}

		assert.Equal(t, 3, receivedNotifications["ali"])
		assert.Equal(t, 3, receivedNotifications["bob"])

		cancel()
	})

	t.Run("Update Device Also Triggers Notification", func(t *testing.T) {
		defer testDB.ClearDB(t)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		conn, err := db.Acquire(ctx)
		assert.NoError(t, err)

		notificationChan, err := device.DeviceCountListener(ctx, conn.Conn())
		assert.NoError(t, err)

		time.Sleep(100 * time.Millisecond)

		_, db := testDB.CreateApp(t)
		defer db.Close()

		devices := createTestDevicesForEmployee(3, "jsm")

		for _, testDevice := range devices {
			err := device.InsertDevice(context.Background(), db, testDevice)
			require.NoError(t, err)
		}

		select {
		case notification := <-notificationChan:
			assert.Equal(t, "jsm", notification.Employee)
			assert.Equal(t, 3, notification.Count)
		case <-time.After(2 * time.Second):
			t.Fatal("Expected notification not received within timeout")
		}

		testDevice := createTestDevice(withEmployee("jdo"))

		err = device.InsertDevice(context.Background(), db, testDevice)
		require.NoError(t, err)

		testDevice.Employee = stringPtr("jsm")
		err = device.UpdateDevice(context.Background(), db, testDevice)
		require.NoError(t, err)

		select {
		case notification := <-notificationChan:
			assert.Equal(t, "jsm", notification.Employee)
			assert.Equal(t, 4, notification.Count)
		case <-time.After(2 * time.Second):
			t.Fatal("Expected update notification not received within timeout")
		}

		cancel()
	})
}
