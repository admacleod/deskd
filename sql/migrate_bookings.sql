CREATE TABLE IF NOT EXISTS bookings (
    user TEXT NOT NULL,
    desk TEXT NOT NULL,
    day TEXT NOT NULL,
    FOREIGN KEY(desk) REFERENCES desks(desk) ON DELETE CASCADE,
    UNIQUE(desk, day),
    UNIQUE(user, day)
) STRICT;