package service

import (
	"game-server/internal/repository"
)

type RankService struct{}

func NewRankService() *RankService {
	return &RankService{}
}

func (s *RankService) GetLeaderboard(limit, offset int) ([]*repository.RankEntry, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	return repository.GetLeaderboard(limit, offset)
}

func (s *RankService) GetUserRank(userID int64) (int, error) {
	return repository.GetUserRank(userID)
}

func (s *RankService) GetUserStats(userID int64) (score, winCount, loseCount int, err error) {
	return repository.GetUserStats(userID)
}
