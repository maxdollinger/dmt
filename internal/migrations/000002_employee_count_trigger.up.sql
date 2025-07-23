CREATE OR REPLACE FUNCTION notify_device_count()
RETURNS TRIGGER AS $$
DECLARE
    device_count INTEGER;
    employee TEXT;
BEGIN
    employee := NEW.employee;

    SELECT COUNT(*)
    INTO device_count
    FROM device
    WHERE device.employee = NEW.employee;

    PERFORM pg_notify('device_count', 
        json_build_object(
            'employee', employee,
            'count', device_count
        )::text
    );
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER device_count_notification_trigger
    AFTER INSERT OR UPDATE OF employee ON device
    FOR EACH ROW
    EXECUTE FUNCTION notify_device_count();
