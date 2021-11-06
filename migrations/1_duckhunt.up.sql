CREATE TABLE `duck_scores` (
    `user_id` VARCHAR(255) NOT NULL,
    `room_id` VARCHAR(255) NOT NULL,
    `score` INT(11) NOT NULL,
    PRIMARY KEY(`user_id`)
);

CREATE TABLE `duck_hunt` (
    `room_id` VARCHAR(255) NOT NULL,
    `enabled` INT(1) NOT NULL,
    PRIMARY KEY(`room_id`)
);

CREATE TABLE `quotes` (
    `room_id` VARCHAR(255) NOT NULL,
    `user_id` VARCHAR(255) NOT NULL,
    `timestamp` INT(11) NOT NULL,
    `quote` TEXT,
    PRIMARY KEY(`user_id`, `room_id`, `quote`)
);
