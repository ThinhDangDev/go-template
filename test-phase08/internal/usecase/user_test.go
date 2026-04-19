package usecase_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/user/test-phase08/internal/domain/entity"
	"github.com/user/test-phase08/internal/usecase"
)

// MockUserRepository is a mock implementation of repository.UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *entity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *entity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) List(ctx context.Context, offset, limit int) ([]*entity.User, int64, error) {
	args := m.Called(ctx, offset, limit)
	return args.Get(0).([]*entity.User), args.Get(1).(int64), args.Error(2)
}

func TestUserUsecase_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		uc := usecase.NewUserUsecase(mockRepo)

		input := usecase.CreateUserInput{
			Email:    "test@example.com",
			Password: "password123",
			Name:     "Test User",
		}

		// Email doesn't exist
		mockRepo.On("GetByEmail", mock.Anything, input.Email).Return(nil, nil)
		// Create succeeds
		mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.User")).Return(nil)

		user, err := uc.Create(context.Background(), input)

		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, input.Email, user.Email)
		assert.Equal(t, input.Name, user.Name)
		mockRepo.AssertExpectations(t)
	})

	t.Run("email already exists", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		uc := usecase.NewUserUsecase(mockRepo)

		existingUser := &entity.User{
			ID:    uuid.New(),
			Email: "test@example.com",
		}

		input := usecase.CreateUserInput{
			Email:    "test@example.com",
			Password: "password123",
			Name:     "Test User",
		}

		mockRepo.On("GetByEmail", mock.Anything, input.Email).Return(existingUser, nil)

		user, err := uc.Create(context.Background(), input)

		assert.Nil(t, user)
		assert.ErrorIs(t, err, usecase.ErrEmailExists)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserUsecase_GetByID(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		uc := usecase.NewUserUsecase(mockRepo)

		userID := uuid.New()
		expectedUser := &entity.User{
			ID:    userID,
			Email: "test@example.com",
			Name:  "Test User",
		}

		mockRepo.On("GetByID", mock.Anything, userID).Return(expectedUser, nil)

		user, err := uc.GetByID(context.Background(), userID)

		require.NoError(t, err)
		assert.Equal(t, expectedUser, user)
		mockRepo.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		uc := usecase.NewUserUsecase(mockRepo)

		userID := uuid.New()
		mockRepo.On("GetByID", mock.Anything, userID).Return(nil, nil)

		user, err := uc.GetByID(context.Background(), userID)

		assert.Nil(t, user)
		assert.ErrorIs(t, err, usecase.ErrUserNotFound)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserUsecase_List(t *testing.T) {
	mockRepo := new(MockUserRepository)
	uc := usecase.NewUserUsecase(mockRepo)

	users := []*entity.User{
		{ID: uuid.New(), Email: "user1@example.com"},
		{ID: uuid.New(), Email: "user2@example.com"},
	}
	var total int64 = 2

	mockRepo.On("List", mock.Anything, 0, 20).Return(users, total, nil)

	input := usecase.ListInput{Page: 1, PageSize: 20}
	result, err := uc.List(context.Background(), input)

	require.NoError(t, err)
	assert.Len(t, result.Users, 2)
	assert.Equal(t, total, result.Total)
	assert.Equal(t, 1, result.TotalPages)
	mockRepo.AssertExpectations(t)
}
