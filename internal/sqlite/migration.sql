-- SPDX-FileCopyrightText: 2024 Alisdair MacLeod <copying@alisdairmacleod.co.uk>
-- SPDX-License-Identifier: AGPL-3.0-only
ALTER TABLE bookings ADD COLUMN desk TEXT;
UPDATE bookings
SET desk = desks.name
FROM desks
WHERE bookings.desk_id = desks.id;
DROP TABLE desks;
ALTER TABLE bookings RENAME TO bookings_old;
CREATE TABLE IF NOT EXISTS bookings (
                                        id INTEGER PRIMARY KEY AUTOINCREMENT,
                                        user TEXT,
                                        desk TEXT,
                                        start DATE,
                                        end DATE
);
INSERT INTO bookings (id, user, desk_id, start, end)
SELECT id, user, desk, start, end FROM bookings_old;
DROP TABLE bookings_old;