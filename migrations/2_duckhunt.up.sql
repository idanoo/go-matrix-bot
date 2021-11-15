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
