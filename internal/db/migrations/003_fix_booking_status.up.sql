ALTER TABLE bookings DROP CONSTRAINT bookings_status_check;
UPDATE bookings SET status = 'canceled' WHERE status = 'cancelled';
ALTER TABLE bookings ADD CONSTRAINT bookings_status_check CHECK (status IN ('active', 'canceled'));
