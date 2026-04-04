ALTER TABLE bookings DROP CONSTRAINT bookings_status_check;
UPDATE bookings SET status = 'cancelled' WHERE status = 'canceled';
ALTER TABLE bookings ADD CONSTRAINT bookings_status_check CHECK (status IN ('active', 'cancelled'));
