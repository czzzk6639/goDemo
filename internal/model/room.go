package model

import "time"

type RoomStatus int

const (
	RoomStatusWaiting RoomStatus = iota
	RoomStatusPlaying
	RoomStatusFinished
)

type Room struct {
	ID        int64      `json:"id"`
	Name      string     `json:"name"`
	CreatorID int64      `json:"creator_id"`
	Players   []int64    `json:"players"`
	Status    RoomStatus `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
}

func NewRoom(id int64, name string, creatorID int64) *Room {
	return &Room{
		ID:        id,
		Name:      name,
		CreatorID: creatorID,
		Players:   []int64{creatorID},
		Status:    RoomStatusWaiting,
		CreatedAt: time.Now(),
	}
}

func (r *Room) IsFull() bool {
	return len(r.Players) >= 2
}

func (r *Room) HasPlayer(userID int64) bool {
	for _, p := range r.Players {
		if p == userID {
			return true
		}
	}
	return false
}

func (r *Room) AddPlayer(userID int64) bool {
	if r.IsFull() || r.HasPlayer(userID) {
		return false
	}
	r.Players = append(r.Players, userID)
	return true
}

func (r *Room) RemovePlayer(userID int64) bool {
	for i, p := range r.Players {
		if p == userID {
			r.Players = append(r.Players[:i], r.Players[i+1:]...)
			return true
		}
	}
	return false
}

func (r *Room) IsEmpty() bool {
	return len(r.Players) == 0
}

func (r *Room) CanStart() bool {
	return len(r.Players) == 2
}

func (r *Room) OtherPlayer(userID int64) int64 {
	for _, p := range r.Players {
		if p != userID {
			return p
		}
	}
	return 0
}
