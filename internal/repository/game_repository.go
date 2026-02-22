package repository

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

var ErrGameNotFound = errors.New("game not found")

type GameRecord struct {
	ID            int64
	RoomID        int64
	BlackPlayerID int64
	WhitePlayerID int64
	WinnerID      int64
	BoardState    string
	MoveHistory   string
	CreatedAt     time.Time
	EndedAt       time.Time
}

func CreateGameRecord(roomID, blackPlayerID, whitePlayerID int64) (int64, error) {
	query := `INSERT INTO games (room_id, black_player_id, white_player_id, created_at) VALUES (?, ?, ?, ?)`
	result, err := DB.Exec(query, roomID, blackPlayerID, whitePlayerID, time.Now())
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func UpdateGameResult(gameID, winnerID int64, boardState string) error {
	query := `UPDATE games SET winner_id = ?, board_state = ?, ended_at = ? WHERE id = ?`
	_, err := DB.Exec(query, winnerID, boardState, time.Now(), gameID)
	return err
}

func GetGameByID(id int64) (*GameRecord, error) {
	record := &GameRecord{}
	query := `SELECT id, room_id, black_player_id, white_player_id, winner_id, board_state, created_at, ended_at FROM games WHERE id = ?`
	err := DB.QueryRow(query, id).Scan(
		&record.ID,
		&record.RoomID,
		&record.BlackPlayerID,
		&record.WhitePlayerID,
		&record.WinnerID,
		&record.BoardState,
		&record.CreatedAt,
		&record.EndedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrGameNotFound
		}
		return nil, err
	}
	return record, nil
}

func GetUserGames(userID int64, limit, offset int) ([]*GameRecord, error) {
	query := `SELECT id, room_id, black_player_id, white_player_id, winner_id, board_state, created_at, ended_at 
			  FROM games 
			  WHERE black_player_id = ? OR white_player_id = ? 
			  ORDER BY created_at DESC 
			  LIMIT ? OFFSET ?`

	rows, err := DB.Query(query, userID, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := make([]*GameRecord, 0)
	for rows.Next() {
		record := &GameRecord{}
		err := rows.Scan(
			&record.ID,
			&record.RoomID,
			&record.BlackPlayerID,
			&record.WhitePlayerID,
			&record.WinnerID,
			&record.BoardState,
			&record.CreatedAt,
			&record.EndedAt,
		)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	return records, nil
}

func GetUserGameCount(userID int64) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM games WHERE black_player_id = ? OR white_player_id = ?`
	err := DB.QueryRow(query, userID, userID).Scan(&count)
	return count, err
}

type RankEntry struct {
	UserID    int64  `json:"user_id"`
	Username  string `json:"username"`
	Score     int    `json:"score"`
	WinCount  int    `json:"win_count"`
	LoseCount int    `json:"lose_count"`
	WinRate   string `json:"win_rate"`
	Rank      int    `json:"rank"`
}

func GetLeaderboard(limit, offset int) ([]*RankEntry, error) {
	query := `SELECT id, username, score, win_count, lose_count 
			  FROM users 
			  ORDER BY score DESC 
			  LIMIT ? OFFSET ?`

	rows, err := DB.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := make([]*RankEntry, 0)
	rank := offset + 1
	for rows.Next() {
		entry := &RankEntry{Rank: rank}
		err := rows.Scan(
			&entry.UserID,
			&entry.Username,
			&entry.Score,
			&entry.WinCount,
			&entry.LoseCount,
		)
		if err != nil {
			return nil, err
		}

		total := entry.WinCount + entry.LoseCount
		if total > 0 {
			winRate := float64(entry.WinCount) / float64(total) * 100
			entry.WinRate = fmt.Sprintf("%.1f%%", winRate)
		} else {
			entry.WinRate = "0.0%"
		}

		entries = append(entries, entry)
		rank++
	}

	return entries, nil
}

func GetUserRank(userID int64) (int, error) {
	query := `SELECT COUNT(*) + 1 FROM users WHERE score > (SELECT score FROM users WHERE id = ?)`
	var rank int
	err := DB.QueryRow(query, userID).Scan(&rank)
	return rank, err
}

func SaveBoardState(roomID int64, board [][]int) error {
	data, err := json.Marshal(board)
	if err != nil {
		return err
	}

	query := `UPDATE games SET board_state = ? WHERE room_id = ? ORDER BY id DESC LIMIT 1`
	_, err = DB.Exec(query, string(data), roomID)
	return err
}

func GetUserStats(userID int64) (score, winCount, loseCount int, err error) {
	query := `SELECT score, win_count, lose_count FROM users WHERE id = ?`
	err = DB.QueryRow(query, userID).Scan(&score, &winCount, &loseCount)
	return
}
