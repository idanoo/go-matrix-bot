CREATE TABLE `rss_feeds` (
    `room_id` VARCHAR(255) NOT NULL,
    `url` TEXT NOT NULL,
    `last_updated` INT(11) DEFAULT 0,
    PRIMARY KEY(`room_id`, `url`)
);
