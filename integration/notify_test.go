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

	t.Run("Notification Triggered When Employee Has 3+ Devices", func(t *testing.T) {
		defer testDB.ClearDB(t)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		notificationChan := device.DeviceCountListener(ctx, testDB.ConnString)

		time.Sleep(100 * time.Millisecond)

		_, db := testDB.CreateApp(t)
		defer db.Close(context.Background())

		// Insert 2 devices for employee "jdo" - should not trigger notification
		testDevices := []*device.Device{
			{
				Name:     "Device 1",
				Type:     "laptop",
				IP:       "192.168.1.100",
				MAC:      "aa:bb:cc:dd:ee:01",
				Employee: stringPtr("jdo"),
			},
			{
				Name:     "Device 2",
				Type:     "phone",
				IP:       "192.168.1.101",
				MAC:      "aa:bb:cc:dd:ee:02",
				Employee: stringPtr("jdo"),
			},
		}

		for _, testDevice := range testDevices {
			err := device.InsertDevice(context.Background(), db, testDevice)
			require.NoError(t, err)
		}

		// Wait a bit to ensure no notification is sent
		select {
		case notification := <-notificationChan:
			t.Fatalf("Unexpected notification received: %+v", notification)
		case <-time.After(500 * time.Millisecond):
			// Expected - no notification should be sent for < 4 devices
		}

		// Insert third device - should trigger notification
		thirdDevice := &device.Device{
			Name:     "Device 3",
			Type:     "tablet",
			IP:       "192.168.1.102",
			MAC:      "aa:bb:cc:dd:ee:03",
			Employee: stringPtr("jdo"),
		}

		err := device.InsertDevice(context.Background(), db, thirdDevice)
		require.NoError(t, err)

		select {
		case notification := <-notificationChan:
			assert.Equal(t, "jdo", notification.Employee)
			assert.Equal(t, 3, notification.Count)
		case <-time.After(2 * time.Second):
			t.Fatal("Expected notification not received within timeout")
		}

		// Insert fourth device - should trigger another notification
		fourthDevice := &device.Device{
			Name:     "Device 4",
			Type:     "desktop",
			IP:       "192.168.1.103",
			MAC:      "aa:bb:cc:dd:ee:04",
			Employee: stringPtr("jdo"),
		}

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

		notificationChan := device.DeviceCountListener(ctx, testDB.ConnString)

		time.Sleep(100 * time.Millisecond)

		_, db := testDB.CreateApp(t)
		defer db.Close(context.Background())

		// Create 3 devices for each of two employees
		employees := []string{"ali", "bob"}
		receivedNotifications := make(map[string]int)

		for _, emp := range employees {
			for i := 1; i <= 3; i++ {
				testDevice := &device.Device{
					Name:     emp + " Device " + string(rune('0'+i)),
					Type:     "laptop",
					IP:       "192.168.1." + string(rune('0'+i)),
					MAC:      emp[:1] + emp[:1] + ":bb:cc:dd:ee:0" + string(rune('0'+i)),
					Employee: &emp,
				}

				err := device.InsertDevice(context.Background(), db, testDevice)
				require.NoError(t, err)
			}
		}

		// Collect notifications for both employees
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

		notificationChan := device.DeviceCountListener(ctx, testDB.ConnString)

		time.Sleep(100 * time.Millisecond)

		_, db := testDB.CreateApp(t)
		defer db.Close(context.Background())

		for i := 1; i <= 3; i++ {
			testDevice := &device.Device{
				Name:     "Cha Device " + string(rune('0'+i)),
				Type:     "laptop",
				IP:       "192.168.1.10" + string(rune('0'+i)),
				MAC:      "cc:cc:cc:dd:ee:0" + string(rune('0'+i)),
				Employee: stringPtr("jsm"),
			}

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

		testDevice := &device.Device{
			Name:     "Device",
			Type:     "laptop",
			IP:       "192.168.1.99",
			MAC:      "cc:cc:cc:dd:ee:99",
			Employee: stringPtr("jdo"),
		}

		err := device.InsertDevice(context.Background(), db, testDevice)
		require.NoError(t, err)

		var deviceID int
		err = db.QueryRow(context.Background(),
			"UPDATE device SET employee = 'jsm' WHERE employee = 'jdo' AND id = (SELECT id FROM device WHERE employee = 'jdo' LIMIT 1) RETURNING id").Scan(&deviceID)
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
