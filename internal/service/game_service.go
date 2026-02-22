package service

import (
	"errors"
	"sync"

	"game-server/internal/model"
)

var (
	ErrGameNotFound      = errors.New("game not found")
	ErrRoomAlreadyInGame = errors.New("room already has a game")
)

type GameService struct {
	games map[int64]*model.Game
	mu    sync.RWMutex
}

func NewGameService() *GameService {
	return &GameService{
		games: make(map[int64]*model.Game),
	}
}

func (s *GameService) StartGame(roomID int64, players []int64) (*model.Game, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.games[roomID]; exists {
		return nil, ErrRoomAlreadyInGame
	}

	game := model.NewGame(roomID, players)
	s.games[roomID] = game

	return game, nil
}

func (s *GameService) GetGame(roomID int64) (*model.Game, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	game, ok := s.games[roomID]
	if !ok {
		return nil, ErrGameNotFound
	}

	return game, nil
}

func (s *GameService) MakeMove(roomID, playerID int64, x, y int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	game, ok := s.games[roomID]
	if !ok {
		return ErrGameNotFound
	}

	return game.MakeMove(playerID, x, y)
}

func (s *GameService) EndGame(roomID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.games, roomID)
}

func (s *GameService) GetCurrentPlayer(roomID int64) (int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	game, ok := s.games[roomID]
	if !ok {
		return 0, ErrGameNotFound
	}

	return game.CurrentPlayer(), nil
}

func (s *GameService) GetGameStatus(roomID int64) (model.GameState, int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	game, ok := s.games[roomID]
	if !ok {
		return 0, 0, ErrGameNotFound
	}

	return game.State, game.Winner, nil
}

func (s *GameService) Forfeit(roomID, playerID int64) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	game, ok := s.games[roomID]
	if !ok {
		return 0, ErrGameNotFound
	}

	winner := game.Forfeit(playerID)
	return winner, nil
}

func (s *GameService) GetBoard(roomID int64) ([][]int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	game, ok := s.games[roomID]
	if !ok {
		return nil, ErrGameNotFound
	}

	return game.GetBoardCopy(), nil
}

func (s *GameService) IsPlayerTurn(roomID, playerID int64) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	game, ok := s.games[roomID]
	if !ok {
		return false, ErrGameNotFound
	}

	return game.CurrentPlayer() == playerID, nil
}
