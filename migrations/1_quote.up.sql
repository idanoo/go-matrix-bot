CREATE TABLE `quotes` (
    `room_id` VARCHAR(255) NOT NULL,
    `user_id` VARCHAR(255) NOT NULL,
    `timestamp` INT(11) NOT NULL,
    `quote` TEXT,
    PRIMARY KEY(`user_id`, `room_id`, `quote`)
);
