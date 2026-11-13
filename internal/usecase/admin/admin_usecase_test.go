package admin

import (
	"context"
	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/domain/models"
	"ecommerce/internal/mocks"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

type fixture struct {
	uc       *AdminUseCase
	userRepo *mocks.MockUserRepository
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	repo := mocks.NewMockUserRepository(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	return &fixture{
		uc:       NewAdminUseCase(repo, log),
		userRepo: repo,
	}
}

func TestAdminUseCase_ListUsers(t *testing.T) {
	t.Run("normalizes pagination bounds", func(t *testing.T) {
		tests := []struct {
			name       string
			page       int
			limit      int
			wantOffset int
			wantLimit  int
		}{
			{"zero page and limit", 0, 0, 0, 10},
			{"negative page and huge limit", -2, 500, 0, 100},
			{"custom page and limit", 3, 20, 40, 20},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				f := newFixture(t)
				expected := []models.User{{ID: 1, Name: "Alice"}}
				f.userRepo.EXPECT().ListUsers(mock.Anything, "alice", tt.wantOffset, tt.wantLimit).
					Return(expected, int64(1), nil).Once()

				users, total, err := f.uc.ListUsers(context.Background(), "alice", tt.page, tt.limit)
				require.NoError(t, err)
				assert.Equal(t, int64(1), total)
				assert.Equal(t, expected, users)
			})
		}
	})

	t.Run("repo error", func(t *testing.T) {
		f := newFixture(t)
		dbErr := errors.New("list failed")
		f.userRepo.EXPECT().ListUsers(mock.Anything, "", 0, 10).Return(nil, int64(0), dbErr).Once()

		users, total, err := f.uc.ListUsers(context.Background(), "", 1, 10)
		require.ErrorIs(t, err, dbErr)
		assert.Nil(t, users)
		assert.Zero(t, total)
	})
}

func TestAdminUseCase_BlockAndUnblockUser(t *testing.T) {
	t.Run("block user", func(t *testing.T) {
		f := newFixture(t)
		require.ErrorIs(t, f.uc.BlockUser(context.Background(), 0), domainerrors.ErrUserNotFound)

		dbErr := errors.New("update failed")
		f.userRepo.EXPECT().SetBlocked(mock.Anything, uint(5), true).Return(dbErr).Once()
		require.ErrorIs(t, f.uc.BlockUser(context.Background(), 5), dbErr)

		f.userRepo.EXPECT().SetBlocked(mock.Anything, uint(5), true).Return(nil).Once()
		require.NoError(t, f.uc.BlockUser(context.Background(), 5))
	})

	t.Run("unblock user", func(t *testing.T) {
		f := newFixture(t)
		require.ErrorIs(t, f.uc.UnblockUser(context.Background(), 0), domainerrors.ErrUserNotFound)

		dbErr := errors.New("update failed")
		f.userRepo.EXPECT().SetBlocked(mock.Anything, uint(5), false).Return(dbErr).Once()
		require.ErrorIs(t, f.uc.UnblockUser(context.Background(), 5), dbErr)

		f.userRepo.EXPECT().SetBlocked(mock.Anything, uint(5), false).Return(nil).Once()
		require.NoError(t, f.uc.UnblockUser(context.Background(), 5))
	})
}

func TestAdminUseCase_AddCustomer(t *testing.T) {
	t.Run("validation errors", func(t *testing.T) {
		tests := []struct {
			name     string
			inName   string
			email    string
			phone    string
			password string
			wantErr  error
		}{
			{"empty field", "", "a@b.com", "9876543210", "Str0ng@123", domainerrors.ErrInvalidCredentials},
			{"invalid name", "123", "a@b.com", "9876543210", "Str0ng@123", domainerrors.ErrInvalidName},
			{"invalid phone", "John Doe", "a@b.com", "123", "Str0ng@123", domainerrors.ErrInvalidPhone},
			{"weak password", "John Doe", "a@b.com", "9876543210", "weak", domainerrors.ErrWeakPassword},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				f := newFixture(t)
				got, err := f.uc.AddCustomer(context.Background(), tt.inName, tt.email, tt.phone, tt.password)
				require.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, got)
			})
		}
	})

	t.Run("email already exists", func(t *testing.T) {
		f := newFixture(t)
		f.userRepo.EXPECT().FindByEmail(mock.Anything, "john@example.com").
			Return(&models.User{ID: 9}, nil).Once()

		got, err := f.uc.AddCustomer(context.Background(), "John Doe", "  JOHN@example.com ", "9876543210", "Str0ng@123")
		require.ErrorIs(t, err, domainerrors.ErrEmailAlreadyExists)
		assert.Nil(t, got)
	})

	t.Run("find email unexpected error", func(t *testing.T) {
		f := newFixture(t)
		dbErr := errors.New("db error")
		f.userRepo.EXPECT().FindByEmail(mock.Anything, "john@example.com").Return(nil, dbErr).Once()

		got, err := f.uc.AddCustomer(context.Background(), "John Doe", "john@example.com", "9876543210", "Str0ng@123")
		require.ErrorIs(t, err, dbErr)
		assert.Nil(t, got)
	})

	t.Run("create error", func(t *testing.T) {
		f := newFixture(t)
		dbErr := errors.New("insert failed")
		f.userRepo.EXPECT().FindByEmail(mock.Anything, "john@example.com").Return(nil, domainerrors.ErrUserNotFound).Once()
		f.userRepo.EXPECT().Create(mock.Anything, mock.AnythingOfType("*models.User")).Return(dbErr).Once()

		got, err := f.uc.AddCustomer(context.Background(), "John Doe", "john@example.com", "9876543210", "Str0ng@123")
		require.ErrorIs(t, err, dbErr)
		assert.Nil(t, got)
	})

	t.Run("success hashes password and normalizes fields", func(t *testing.T) {
		f := newFixture(t)
		f.userRepo.EXPECT().FindByEmail(mock.Anything, "john@example.com").Return(nil, domainerrors.ErrUserNotFound).Once()
		f.userRepo.EXPECT().Create(mock.Anything, mock.MatchedBy(func(u *models.User) bool {
			return u.Name == "John Doe" &&
				u.Email == "john@example.com" &&
				u.Phone == "9876543210" &&
				u.Role == models.RoleUser &&
				!u.IsBlocked &&
				u.Password != nil &&
				bcrypt.CompareHashAndPassword([]byte(*u.Password), []byte("Str0ng@123")) == nil
		})).Run(func(_ context.Context, u *models.User) {
			u.ID = 42
		}).Return(nil).Once()

		got, err := f.uc.AddCustomer(context.Background(), "  John Doe ", " JOHN@EXAMPLE.COM ", " 9876543210 ", "Str0ng@123")
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, uint(42), got.ID)
	})
}

func TestAdminUseCase_GetCustomer(t *testing.T) {
	t.Run("invalid id", func(t *testing.T) {
		f := newFixture(t)
		got, err := f.uc.GetCustomer(context.Background(), 0)
		require.ErrorIs(t, err, domainerrors.ErrUserNotFound)
		assert.Nil(t, got)
	})

	t.Run("repo error", func(t *testing.T) {
		f := newFixture(t)
		f.userRepo.EXPECT().FindByID(mock.Anything, uint(7)).Return(nil, domainerrors.ErrUserNotFound).Once()

		got, err := f.uc.GetCustomer(context.Background(), 7)
		require.ErrorIs(t, err, domainerrors.ErrUserNotFound)
		assert.Nil(t, got)
	})

	t.Run("non-user role rejected", func(t *testing.T) {
		f := newFixture(t)
		f.userRepo.EXPECT().FindByID(mock.Anything, uint(7)).Return(&models.User{ID: 7, Role: models.RoleAdmin}, nil).Once()

		got, err := f.uc.GetCustomer(context.Background(), 7)
		require.ErrorIs(t, err, domainerrors.ErrUserNotFound)
		assert.Nil(t, got)
	})

	t.Run("success", func(t *testing.T) {
		f := newFixture(t)
		want := &models.User{ID: 7, Role: models.RoleUser, Name: "Alice"}
		f.userRepo.EXPECT().FindByID(mock.Anything, uint(7)).Return(want, nil).Once()

		got, err := f.uc.GetCustomer(context.Background(), 7)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})
}

func TestAdminUseCase_UpdateCustomer(t *testing.T) {
	t.Run("validation errors", func(t *testing.T) {
		tests := []struct {
			name    string
			userID  uint
			inName  string
			email   string
			phone   string
			wantErr error
		}{
			{"zero id", 0, "John Doe", "a@b.com", "9876543210", domainerrors.ErrUserNotFound},
			{"empty field", 1, "John Doe", "", "9876543210", domainerrors.ErrInvalidCredentials},
			{"invalid name", 1, "1", "a@b.com", "9876543210", domainerrors.ErrInvalidName},
			{"invalid phone", 1, "John Doe", "a@b.com", "123", domainerrors.ErrInvalidPhone},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				f := newFixture(t)
				got, err := f.uc.UpdateCustomer(context.Background(), tt.userID, tt.inName, tt.email, tt.phone)
				require.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, got)
			})
		}
	})

	t.Run("find user failure or non-customer role", func(t *testing.T) {
		f := newFixture(t)
		f.userRepo.EXPECT().FindByID(mock.Anything, uint(3)).Return(nil, domainerrors.ErrUserNotFound).Once()
		_, err := f.uc.UpdateCustomer(context.Background(), 3, "John Doe", "a@b.com", "9876543210")
		require.ErrorIs(t, err, domainerrors.ErrUserNotFound)

		f.userRepo.EXPECT().FindByID(mock.Anything, uint(3)).Return(&models.User{ID: 3, Role: models.RoleAdmin}, nil).Once()
		_, err = f.uc.UpdateCustomer(context.Background(), 3, "John Doe", "a@b.com", "9876543210")
		require.ErrorIs(t, err, domainerrors.ErrUserNotFound)
	})

	t.Run("email conflict and lookup error", func(t *testing.T) {
		f := newFixture(t)
		existing := &models.User{ID: 3, Role: models.RoleUser, Email: "old@example.com"}

		f.userRepo.EXPECT().FindByID(mock.Anything, uint(3)).Return(existing, nil).Once()
		f.userRepo.EXPECT().FindByEmail(mock.Anything, "taken@example.com").Return(&models.User{ID: 99}, nil).Once()
		_, err := f.uc.UpdateCustomer(context.Background(), 3, "John Doe", "taken@example.com", "9876543210")
		require.ErrorIs(t, err, domainerrors.ErrEmailAlreadyExists)

		dbErr := errors.New("lookup failed")
		f.userRepo.EXPECT().FindByID(mock.Anything, uint(3)).Return(existing, nil).Once()
		f.userRepo.EXPECT().FindByEmail(mock.Anything, "new@example.com").Return(nil, dbErr).Once()
		_, err = f.uc.UpdateCustomer(context.Background(), 3, "John Doe", "new@example.com", "9876543210")
		require.ErrorIs(t, err, dbErr)
	})

	t.Run("update profile and email errors", func(t *testing.T) {
		dbErr := errors.New("write failed")
		existing := &models.User{ID: 3, Role: models.RoleUser, Email: "old@example.com"}

		t.Run("update profile fails", func(t *testing.T) {
			f := newFixture(t)
			f.userRepo.EXPECT().FindByID(mock.Anything, uint(3)).Return(existing, nil).Once()
			f.userRepo.EXPECT().UpdateProfile(mock.Anything, uint(3), "John Doe", "9876543210").Return(dbErr).Once()

			_, err := f.uc.UpdateCustomer(context.Background(), 3, "John Doe", "old@example.com", "9876543210")
			require.ErrorIs(t, err, dbErr)
		})

		t.Run("update email fails", func(t *testing.T) {
			f := newFixture(t)
			f.userRepo.EXPECT().FindByID(mock.Anything, uint(3)).Return(existing, nil).Once()
			f.userRepo.EXPECT().FindByEmail(mock.Anything, "new@example.com").Return(nil, domainerrors.ErrUserNotFound).Once()
			f.userRepo.EXPECT().UpdateProfile(mock.Anything, uint(3), "John Doe", "9876543210").Return(nil).Once()
			f.userRepo.EXPECT().UpdateEmail(mock.Anything, uint(3), "new@example.com").Return(dbErr).Once()

			_, err := f.uc.UpdateCustomer(context.Background(), 3, "John Doe", "new@example.com", "9876543210")
			require.ErrorIs(t, err, dbErr)
		})
	})

	t.Run("success with email change", func(t *testing.T) {
		f := newFixture(t)
		existing := &models.User{ID: 3, Role: models.RoleUser, Email: "old@example.com"}
		updated := &models.User{ID: 3, Role: models.RoleUser, Name: "John Doe", Email: "new@example.com", Phone: "9876543210"}

		f.userRepo.EXPECT().FindByID(mock.Anything, uint(3)).Return(existing, nil).Once()
		f.userRepo.EXPECT().FindByEmail(mock.Anything, "new@example.com").Return(nil, domainerrors.ErrUserNotFound).Once()
		f.userRepo.EXPECT().UpdateProfile(mock.Anything, uint(3), "John Doe", "9876543210").Return(nil).Once()
		f.userRepo.EXPECT().UpdateEmail(mock.Anything, uint(3), "new@example.com").Return(nil).Once()
		f.userRepo.EXPECT().FindByID(mock.Anything, uint(3)).Return(updated, nil).Once()

		got, err := f.uc.UpdateCustomer(context.Background(), 3, " John Doe ", " NEW@example.com ", " 9876543210 ")
		require.NoError(t, err)
		assert.Equal(t, updated, got)
	})
}

func TestAdminUseCase_DeleteCustomer(t *testing.T) {
	t.Run("invalid id", func(t *testing.T) {
		f := newFixture(t)
		require.ErrorIs(t, f.uc.DeleteCustomer(context.Background(), 0), domainerrors.ErrUserNotFound)
	})

	t.Run("find error", func(t *testing.T) {
		f := newFixture(t)
		f.userRepo.EXPECT().FindByID(mock.Anything, uint(4)).Return(nil, domainerrors.ErrUserNotFound).Once()
		require.ErrorIs(t, f.uc.DeleteCustomer(context.Background(), 4), domainerrors.ErrUserNotFound)
	})

	t.Run("non-user role rejected", func(t *testing.T) {
		f := newFixture(t)
		f.userRepo.EXPECT().FindByID(mock.Anything, uint(4)).Return(&models.User{ID: 4, Role: models.RoleAdmin}, nil).Once()
		require.ErrorIs(t, f.uc.DeleteCustomer(context.Background(), 4), domainerrors.ErrUserNotFound)
	})

	t.Run("delete repo error", func(t *testing.T) {
		f := newFixture(t)
		dbErr := errors.New("delete failed")
		f.userRepo.EXPECT().FindByID(mock.Anything, uint(4)).Return(&models.User{ID: 4, Role: models.RoleUser}, nil).Once()
		f.userRepo.EXPECT().Delete(mock.Anything, uint(4)).Return(dbErr).Once()
		require.ErrorIs(t, f.uc.DeleteCustomer(context.Background(), 4), dbErr)
	})

	t.Run("success", func(t *testing.T) {
		f := newFixture(t)
		f.userRepo.EXPECT().FindByID(mock.Anything, uint(4)).Return(&models.User{ID: 4, Role: models.RoleUser}, nil).Once()
		f.userRepo.EXPECT().Delete(mock.Anything, uint(4)).Return(nil).Once()
		require.NoError(t, f.uc.DeleteCustomer(context.Background(), 4))
	})
}
