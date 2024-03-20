-- SPDX-FileCopyrightText: 2024 Alisdair MacLeod <copying@alisdairmacleod.co.uk>
-- SPDX-License-Identifier: AGPL-3.0-only
ALTER TABLE bookings ADD COLUMN user TEXT;
UPDATE bookings
    SET user = users.email
    FROM users
    WHERE bookings.user_id = users.id;
DROP TABLE users;
ALTER TABLE bookings RENAME TO bookings_old;
CREATE TABLE IF NOT EXISTS bookings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user TEXT,
    desk_id INTEGER,
    start DATE,
    end DATE,
    FOREIGN KEY (desk_id)
        REFERENCES desks (id)
        ON DELETE CASCADE
        ON UPDATE NO ACTION
);
INSERT INTO bookings (id, user, desk_id, start, end)
    SELECT id, user, desk_id, start, end FROM bookings_old;
DROP TABLE bookings_old;