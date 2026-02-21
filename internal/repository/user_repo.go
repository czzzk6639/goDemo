package repository

import (
	"database/sql"
	"errors"

	"game-server/internal/model"
)

var ErrUserNotFound = errors.New("user not found")
var ErrUserAlreadyExists = errors.New("user already exists")

func CreateUser(user *model.User) error {
	query := `INSERT INTO users (username, password, score, win_count, lose_count) VALUES (?, ?, ?, ?, ?)`
	result, err := DB.Exec(query, user.Username, user.Password, user.Score, user.WinCount, user.LoseCount)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	user.ID = id
	return nil
}

func GetUserByUsername(username string) (*model.User, error) {
	user := &model.User{}
	query := `SELECT id, username, password, score, win_count, lose_count, created_at FROM users WHERE username = ?`
	err := DB.QueryRow(query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Password,
		&user.Score,
		&user.WinCount,
		&user.LoseCount,
		&user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func GetUserByID(id int64) (*model.User, error) {
	user := &model.User{}
	query := `SELECT id, username, password, score, win_count, lose_count, created_at FROM users WHERE id = ?`
	err := DB.QueryRow(query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Password,
		&user.Score,
		&user.WinCount,
		&user.LoseCount,
		&user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func UpdateUserScore(userID int64, scoreDelta int, isWin bool) error {
	var query string
	if isWin {
		query = `UPDATE users SET score = score + ?, win_count = win_count + 1 WHERE id = ?`
	} else {
		query = `UPDATE users SET score = score + ?, lose_count = lose_count + 1 WHERE id = ?`
	}
	_, err := DB.Exec(query, scoreDelta, userID)
	return err
}
