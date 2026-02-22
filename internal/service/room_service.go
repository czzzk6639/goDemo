package service

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"game-server/internal/model"
)

var (
	ErrRoomNotFound      = errors.New("room not found")
	ErrRoomFull          = errors.New("room is full")
	ErrAlreadyInRoom     = errors.New("already in room")
	ErrNotInRoom         = errors.New("not in room")
	ErrNotRoomCreator    = errors.New("not room creator")
	ErrRoomAlreadyExists = errors.New("room already exists")
)

type RoomService struct {
	rooms     map[int64]*model.Room
	mu        sync.RWMutex
	idCounter int64
}

func NewRoomService() *RoomService {
	return &RoomService{
		rooms: make(map[int64]*model.Room),
	}
}

func (s *RoomService) CreateRoom(name string, creatorID int64) (*model.Room, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := atomic.AddInt64(&s.idCounter, 1)
	room := model.NewRoom(id, name, creatorID)
	s.rooms[id] = room

	return room, nil
}

func (s *RoomService) GetRoom(roomID int64) (*model.Room, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	room, ok := s.rooms[roomID]
	if !ok {
		return nil, ErrRoomNotFound
	}
	return room, nil
}

func (s *RoomService) JoinRoom(roomID, userID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	room, ok := s.rooms[roomID]
	if !ok {
		return ErrRoomNotFound
	}

	if room.HasPlayer(userID) {
		return ErrAlreadyInRoom
	}

	if room.IsFull() {
		return ErrRoomFull
	}

	room.AddPlayer(userID)
	return nil
}

func (s *RoomService) LeaveRoom(roomID, userID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	room, ok := s.rooms[roomID]
	if !ok {
		return ErrRoomNotFound
	}

	if !room.HasPlayer(userID) {
		return ErrNotInRoom
	}

	room.RemovePlayer(userID)

	if room.IsEmpty() {
		delete(s.rooms, roomID)
	}

	return nil
}

func (s *RoomService) ListRooms() []*model.Room {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rooms := make([]*model.Room, 0, len(s.rooms))
	for _, room := range s.rooms {
		rooms = append(rooms, room)
	}
	return rooms
}

func (s *RoomService) ListWaitingRooms() []*model.Room {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rooms := make([]*model.Room, 0)
	for _, room := range s.rooms {
		if room.Status == model.RoomStatusWaiting {
			rooms = append(rooms, room)
		}
	}
	return rooms
}

func (s *RoomService) StartGame(roomID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	room, ok := s.rooms[roomID]
	if !ok {
		return ErrRoomNotFound
	}

	if !room.CanStart() {
		return errors.New("not enough players")
	}

	room.Status = model.RoomStatusPlaying
	return nil
}

func (s *RoomService) EndGame(roomID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	room, ok := s.rooms[roomID]
	if !ok {
		return ErrRoomNotFound
	}

	room.Status = model.RoomStatusFinished
	return nil
}

func (s *RoomService) DeleteRoom(roomID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.rooms, roomID)
}

func (s *RoomService) GetPlayerRoom(userID int64) *model.Room {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, room := range s.rooms {
		if room.HasPlayer(userID) {
			return room
		}
	}
	return nil
}

func (s *RoomService) GetRoomPlayers(roomID int64) ([]int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	room, ok := s.rooms[roomID]
	if !ok {
		return nil, ErrRoomNotFound
	}

	players := make([]int64, len(room.Players))
	copy(players, room.Players)
	return players, nil
}

func (s *RoomService) SetRoomStatus(roomID int64, status model.RoomStatus) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	room, ok := s.rooms[roomID]
	if !ok {
		return ErrRoomNotFound
	}

	room.Status = status
	return nil
}

func (s *RoomService) CleanEmptyRooms() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	count := 0
	for id, room := range s.rooms {
		if room.IsEmpty() {
			delete(s.rooms, id)
			count++
		}
	}
	return count
}

func (s *RoomService) CleanInactiveRooms(timeout time.Duration) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	count := 0
	now := time.Now()
	for id, room := range s.rooms {
		if now.Sub(room.CreatedAt) > timeout && room.Status == model.RoomStatusWaiting && len(room.Players) == 1 {
			delete(s.rooms, id)
			count++
		}
	}
	return count
}
