package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"

	"game-server/internal/model"
	"game-server/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidPassword = errors.New("invalid password")

type UserService struct{}

func NewUserService() *UserService {
	return &UserService{}
}

func (s *UserService) Register(req *model.UserRegisterRequest) (*model.User, error) {
	_, err := repository.GetUserByUsername(req.Username)
	if err == nil {
		return nil, repository.ErrUserAlreadyExists
	}
	if !errors.Is(err, repository.ErrUserNotFound) {
		return nil, err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Username: req.Username,
		Password: string(hashedPassword),
		Score:    1000,
	}

	if err := repository.CreateUser(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) Login(req *model.UserLoginRequest) (*model.User, string, error) {
	user, err := repository.GetUserByUsername(req.Username)
	if err != nil {
		return nil, "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, "", ErrInvalidPassword
	}

	token, err := generateToken()
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

func (s *UserService) GetUserByID(id int64) (*model.User, error) {
	return repository.GetUserByID(id)
}

func (s *UserService) UpdateScore(userID int64, scoreDelta int, isWin bool) error {
	return repository.UpdateUserScore(userID, scoreDelta, isWin)
}

func (s *UserService) GetUserStats(userID int64) (score, winCount, loseCount int, err error) {
	return repository.GetUserStats(userID)
}

func (s *UserService) GetUserRank(userID int64) (int, error) {
	return repository.GetUserRank(userID)
}

func generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
