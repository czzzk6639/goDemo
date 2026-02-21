CREATE DATABASE IF NOT EXISTS game_server DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE game_server;

CREATE TABLE IF NOT EXISTS users (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    username VARCHAR(50) UNIQUE NOT NULL,
    password VARCHAR(100) NOT NULL,
    score INT DEFAULT 1000,
    win_count INT DEFAULT 0,
    lose_count INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_username (username),
    INDEX idx_score (score)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS games (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    room_id VARCHAR(50) NOT NULL,
    black_player_id BIGINT NOT NULL,
    white_player_id BIGINT NOT NULL,
    winner_id BIGINT,
    board_state TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    ended_at TIMESTAMP NULL,
    INDEX idx_room_id (room_id),
    INDEX idx_players (black_player_id, white_player_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
