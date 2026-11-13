package address

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
)

type fixture struct {
	uc   *AddressUseCase
	repo *mocks.MockAddressRepository
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	repo := mocks.NewMockAddressRepository(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	return &fixture{
		uc:   NewAddressUseCase(repo, log),
		repo: repo,
	}
}

func validAddressInput() *models.Address {
	return &models.Address{
		Name:         "  John Doe  ",
		Phone:        " 9876543210 ",
		AddressLine1: " 123 Main St ",
		AddressLine2: " Apt 4B ",
		City:         " Kochi ",
		PostalCode:   " 682001 ",
		State:        " Kerala ",
		Country:      "",
	}
}

func TestAddressUseCase_AddAddress(t *testing.T) {
	t.Run("validation errors", func(t *testing.T) {
		tests := []struct {
			name    string
			userID  uint
			mutate  func(*models.Address)
			wantErr error
		}{
			{
				name:    "invalid user id",
				userID:  0,
				mutate:  func(_ *models.Address) {},
				wantErr: domainerrors.ErrInvalidUserID,
			},
			{
				name:    "invalid name",
				userID:  1,
				mutate:  func(a *models.Address) { a.Name = "123!" },
				wantErr: domainerrors.ErrInvalidName,
			},
			{
				name:    "invalid phone",
				userID:  1,
				mutate:  func(a *models.Address) { a.Phone = "12345" },
				wantErr: domainerrors.ErrInvalidPhone,
			},
			{
				name:    "missing address line 1",
				userID:  1,
				mutate:  func(a *models.Address) { a.AddressLine1 = "   " },
				wantErr: domainerrors.ErrInvalidAddress,
			},
			{
				name:    "missing city",
				userID:  1,
				mutate:  func(a *models.Address) { a.City = "" },
				wantErr: domainerrors.ErrInvalidAddress,
			},
			{
				name:    "missing postal code",
				userID:  1,
				mutate:  func(a *models.Address) { a.PostalCode = "" },
				wantErr: domainerrors.ErrInvalidAddress,
			},
			{
				name:    "missing state",
				userID:  1,
				mutate:  func(a *models.Address) { a.State = "" },
				wantErr: domainerrors.ErrInvalidAddress,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				f := newFixture(t)
				addr := validAddressInput()
				tt.mutate(addr)

				got, err := f.uc.AddAddress(context.Background(), tt.userID, addr)
				require.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, got)
			})
		}
	})

	t.Run("first address sets default and trims fields", func(t *testing.T) {
		f := newFixture(t)
		addr := validAddressInput()

		f.repo.EXPECT().CountByUserID(mock.Anything, uint(10)).Return(int64(0), nil).Once()
		f.repo.EXPECT().Create(mock.Anything, mock.MatchedBy(func(a *models.Address) bool {
			return a.UserID == 10 &&
				a.Name == "John Doe" &&
				a.Phone == "9876543210" &&
				a.AddressLine1 == "123 Main St" &&
				a.AddressLine2 == "Apt 4B" &&
				a.City == "Kochi" &&
				a.PostalCode == "682001" &&
				a.State == "Kerala" &&
				a.Country == "India" &&
				a.IsDefault
		})).Run(func(_ context.Context, a *models.Address) {
			a.ID = 7
		}).Return(nil).Once()

		got, err := f.uc.AddAddress(context.Background(), 10, addr)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, uint(7), got.ID)
		assert.True(t, got.IsDefault)
		assert.Equal(t, "India", got.Country)
	})

	t.Run("subsequent default address clears existing default", func(t *testing.T) {
		f := newFixture(t)
		addr := validAddressInput()
		addr.Country = "USA"
		addr.IsDefault = true

		f.repo.EXPECT().CountByUserID(mock.Anything, uint(10)).Return(int64(2), nil).Once()
		f.repo.EXPECT().ClearDefaultByUserID(mock.Anything, uint(10)).Return(nil).Once()
		f.repo.EXPECT().Create(mock.Anything, addr).Return(nil).Once()

		got, err := f.uc.AddAddress(context.Background(), 10, addr)
		require.NoError(t, err)
		assert.Equal(t, "USA", got.Country)
		assert.True(t, got.IsDefault)
	})

	t.Run("subsequent non-default address does not clear default", func(t *testing.T) {
		f := newFixture(t)
		addr := validAddressInput()
		addr.IsDefault = false

		f.repo.EXPECT().CountByUserID(mock.Anything, uint(10)).Return(int64(1), nil).Once()
		f.repo.EXPECT().Create(mock.Anything, addr).Return(nil).Once()

		got, err := f.uc.AddAddress(context.Background(), 10, addr)
		require.NoError(t, err)
		assert.False(t, got.IsDefault)
	})

	t.Run("repository errors", func(t *testing.T) {
		dbErr := errors.New("db error")

		t.Run("count fails", func(t *testing.T) {
			f := newFixture(t)
			f.repo.EXPECT().CountByUserID(mock.Anything, uint(10)).Return(int64(0), dbErr).Once()

			got, err := f.uc.AddAddress(context.Background(), 10, validAddressInput())
			require.ErrorIs(t, err, dbErr)
			assert.Nil(t, got)
		})

		t.Run("clear default fails", func(t *testing.T) {
			f := newFixture(t)
			addr := validAddressInput()
			addr.IsDefault = true
			f.repo.EXPECT().CountByUserID(mock.Anything, uint(10)).Return(int64(1), nil).Once()
			f.repo.EXPECT().ClearDefaultByUserID(mock.Anything, uint(10)).Return(dbErr).Once()

			got, err := f.uc.AddAddress(context.Background(), 10, addr)
			require.ErrorIs(t, err, dbErr)
			assert.Nil(t, got)
		})

		t.Run("create fails", func(t *testing.T) {
			f := newFixture(t)
			f.repo.EXPECT().CountByUserID(mock.Anything, uint(10)).Return(int64(0), nil).Once()
			f.repo.EXPECT().Create(mock.Anything, mock.Anything).Return(dbErr).Once()

			got, err := f.uc.AddAddress(context.Background(), 10, validAddressInput())
			require.ErrorIs(t, err, dbErr)
			assert.Nil(t, got)
		})
	})
}

func TestAddressUseCase_GetAddresses(t *testing.T) {
	t.Run("invalid user id", func(t *testing.T) {
		f := newFixture(t)
		got, err := f.uc.GetAddresses(context.Background(), 0)
		require.ErrorIs(t, err, domainerrors.ErrInvalidUserID)
		assert.Nil(t, got)
	})

	t.Run("repo error", func(t *testing.T) {
		f := newFixture(t)
		dbErr := errors.New("query failed")
		f.repo.EXPECT().GetByUserID(mock.Anything, uint(5)).Return(nil, dbErr).Once()

		got, err := f.uc.GetAddresses(context.Background(), 5)
		require.ErrorIs(t, err, dbErr)
		assert.Nil(t, got)
	})

	t.Run("success", func(t *testing.T) {
		f := newFixture(t)
		want := []models.Address{{ID: 1, UserID: 5, City: "Kochi"}}
		f.repo.EXPECT().GetByUserID(mock.Anything, uint(5)).Return(want, nil).Once()

		got, err := f.uc.GetAddresses(context.Background(), 5)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})
}

func TestAddressUseCase_UpdateAddress(t *testing.T) {
	t.Run("validation errors", func(t *testing.T) {
		tests := []struct {
			name      string
			userID    uint
			addressID uint
			mutate    func(*models.Address)
			wantErr   error
		}{
			{"zero user id", 0, 1, func(_ *models.Address) {}, domainerrors.ErrInvalidUserID},
			{"zero address id", 1, 0, func(_ *models.Address) {}, domainerrors.ErrInvalidUserID},
			{"invalid name", 1, 2, func(a *models.Address) { a.Name = "A" }, domainerrors.ErrInvalidName},
			{"invalid phone", 1, 2, func(a *models.Address) { a.Phone = "bad" }, domainerrors.ErrInvalidPhone},
			{"missing city", 1, 2, func(a *models.Address) { a.City = "" }, domainerrors.ErrInvalidAddress},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				f := newFixture(t)
				addr := validAddressInput()
				tt.mutate(addr)

				got, err := f.uc.UpdateAddress(context.Background(), tt.userID, tt.addressID, addr)
				require.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, got)
			})
		}
	})

	t.Run("find address error", func(t *testing.T) {
		f := newFixture(t)
		f.repo.EXPECT().FindByID(mock.Anything, uint(1), uint(2)).Return(nil, domainerrors.ErrRecordNotFound).Once()

		got, err := f.uc.UpdateAddress(context.Background(), 1, 2, validAddressInput())
		require.ErrorIs(t, err, domainerrors.ErrRecordNotFound)
		assert.Nil(t, got)
	})

	t.Run("promoting to default clears existing default", func(t *testing.T) {
		f := newFixture(t)
		addr := validAddressInput()
		addr.IsDefault = true

		f.repo.EXPECT().FindByID(mock.Anything, uint(1), uint(2)).Return(&models.Address{ID: 2, IsDefault: false}, nil).Once()
		f.repo.EXPECT().ClearDefaultByUserID(mock.Anything, uint(1)).Return(nil).Once()
		f.repo.EXPECT().Update(mock.Anything, uint(1), uint(2), mock.MatchedBy(func(a *models.Address) bool {
			return a.ID == 2 && a.UserID == 1 && a.Country == "India"
		})).Return(nil).Once()

		got, err := f.uc.UpdateAddress(context.Background(), 1, 2, addr)
		require.NoError(t, err)
		assert.Equal(t, uint(2), got.ID)
		assert.Equal(t, uint(1), got.UserID)
	})

	t.Run("clear default failure aborts update", func(t *testing.T) {
		f := newFixture(t)
		addr := validAddressInput()
		addr.IsDefault = true
		dbErr := errors.New("clear failed")

		f.repo.EXPECT().FindByID(mock.Anything, uint(1), uint(2)).Return(&models.Address{ID: 2, IsDefault: false}, nil).Once()
		f.repo.EXPECT().ClearDefaultByUserID(mock.Anything, uint(1)).Return(dbErr).Once()

		got, err := f.uc.UpdateAddress(context.Background(), 1, 2, addr)
		require.ErrorIs(t, err, dbErr)
		assert.Nil(t, got)
	})

	t.Run("update failure returns error", func(t *testing.T) {
		f := newFixture(t)
		addr := validAddressInput()
		addr.IsDefault = true
		dbErr := errors.New("update failed")

		f.repo.EXPECT().FindByID(mock.Anything, uint(1), uint(2)).Return(&models.Address{ID: 2, IsDefault: true}, nil).Once()
		f.repo.EXPECT().Update(mock.Anything, uint(1), uint(2), addr).Return(dbErr).Once()

		got, err := f.uc.UpdateAddress(context.Background(), 1, 2, addr)
		require.ErrorIs(t, err, dbErr)
		assert.Nil(t, got)
	})
}

func TestAddressUseCase_DeleteAddress(t *testing.T) {
	t.Run("invalid ids", func(t *testing.T) {
		f := newFixture(t)
		require.ErrorIs(t, f.uc.DeleteAddress(context.Background(), 0, 1), domainerrors.ErrInvalidUserID)
		require.ErrorIs(t, f.uc.DeleteAddress(context.Background(), 1, 0), domainerrors.ErrInvalidUserID)
	})

	t.Run("find error", func(t *testing.T) {
		f := newFixture(t)
		f.repo.EXPECT().FindByID(mock.Anything, uint(1), uint(2)).Return(nil, domainerrors.ErrRecordNotFound).Once()

		err := f.uc.DeleteAddress(context.Background(), 1, 2)
		require.ErrorIs(t, err, domainerrors.ErrRecordNotFound)
	})

	t.Run("delete error", func(t *testing.T) {
		f := newFixture(t)
		dbErr := errors.New("delete failed")
		f.repo.EXPECT().FindByID(mock.Anything, uint(1), uint(2)).Return(&models.Address{ID: 2, IsDefault: false}, nil).Once()
		f.repo.EXPECT().Delete(mock.Anything, uint(1), uint(2)).Return(dbErr).Once()

		err := f.uc.DeleteAddress(context.Background(), 1, 2)
		require.ErrorIs(t, err, dbErr)
	})

	t.Run("delete non-default success", func(t *testing.T) {
		f := newFixture(t)
		f.repo.EXPECT().FindByID(mock.Anything, uint(1), uint(2)).Return(&models.Address{ID: 2, IsDefault: false}, nil).Once()
		f.repo.EXPECT().Delete(mock.Anything, uint(1), uint(2)).Return(nil).Once()

		err := f.uc.DeleteAddress(context.Background(), 1, 2)
		require.NoError(t, err)
	})

	t.Run("delete default promotes another address", func(t *testing.T) {
		f := newFixture(t)
		other := &models.Address{ID: 3, UserID: 1, IsDefault: false}
		f.repo.EXPECT().FindByID(mock.Anything, uint(1), uint(2)).Return(&models.Address{ID: 2, IsDefault: true}, nil).Once()
		f.repo.EXPECT().Delete(mock.Anything, uint(1), uint(2)).Return(nil).Once()
		f.repo.EXPECT().FindAnotherAddress(mock.Anything, uint(1), uint(2)).Return(other, nil).Once()
		f.repo.EXPECT().ClearDefaultByUserID(mock.Anything, uint(1)).Return(nil).Once()
		f.repo.EXPECT().Update(mock.Anything, uint(1), uint(3), mock.MatchedBy(func(a *models.Address) bool {
			return a.ID == 3 && a.IsDefault
		})).Return(nil).Once()

		err := f.uc.DeleteAddress(context.Background(), 1, 2)
		require.NoError(t, err)
	})

	t.Run("delete default when no other address exists", func(t *testing.T) {
		f := newFixture(t)
		f.repo.EXPECT().FindByID(mock.Anything, uint(1), uint(2)).Return(&models.Address{ID: 2, IsDefault: true}, nil).Once()
		f.repo.EXPECT().Delete(mock.Anything, uint(1), uint(2)).Return(nil).Once()
		f.repo.EXPECT().FindAnotherAddress(mock.Anything, uint(1), uint(2)).Return(nil, domainerrors.ErrRecordNotFound).Once()

		err := f.uc.DeleteAddress(context.Background(), 1, 2)
		require.NoError(t, err)
	})

	t.Run("delete default fails when clearing or updating fallback", func(t *testing.T) {
		dbErr := errors.New("tx failure")

		t.Run("clear default fails", func(t *testing.T) {
			f := newFixture(t)
			other := &models.Address{ID: 3, UserID: 1}
			f.repo.EXPECT().FindByID(mock.Anything, uint(1), uint(2)).Return(&models.Address{ID: 2, IsDefault: true}, nil).Once()
			f.repo.EXPECT().Delete(mock.Anything, uint(1), uint(2)).Return(nil).Once()
			f.repo.EXPECT().FindAnotherAddress(mock.Anything, uint(1), uint(2)).Return(other, nil).Once()
			f.repo.EXPECT().ClearDefaultByUserID(mock.Anything, uint(1)).Return(dbErr).Once()

			err := f.uc.DeleteAddress(context.Background(), 1, 2)
			require.ErrorIs(t, err, dbErr)
		})

		t.Run("update fallback fails", func(t *testing.T) {
			f := newFixture(t)
			other := &models.Address{ID: 3, UserID: 1}
			f.repo.EXPECT().FindByID(mock.Anything, uint(1), uint(2)).Return(&models.Address{ID: 2, IsDefault: true}, nil).Once()
			f.repo.EXPECT().Delete(mock.Anything, uint(1), uint(2)).Return(nil).Once()
			f.repo.EXPECT().FindAnotherAddress(mock.Anything, uint(1), uint(2)).Return(other, nil).Once()
			f.repo.EXPECT().ClearDefaultByUserID(mock.Anything, uint(1)).Return(nil).Once()
			f.repo.EXPECT().Update(mock.Anything, uint(1), uint(3), other).Return(dbErr).Once()

			err := f.uc.DeleteAddress(context.Background(), 1, 2)
			require.ErrorIs(t, err, dbErr)
		})
	})
}

func TestAddressUseCase_SetDefaultAddress(t *testing.T) {
	t.Run("invalid ids", func(t *testing.T) {
		f := newFixture(t)
		require.ErrorIs(t, f.uc.SetDefaultAddress(context.Background(), 0, 1), domainerrors.ErrInvalidUserID)
		require.ErrorIs(t, f.uc.SetDefaultAddress(context.Background(), 1, 0), domainerrors.ErrInvalidUserID)
	})

	t.Run("repo error", func(t *testing.T) {
		f := newFixture(t)
		dbErr := errors.New("set default failed")
		f.repo.EXPECT().SetDefault(mock.Anything, uint(1), uint(2)).Return(dbErr).Once()

		err := f.uc.SetDefaultAddress(context.Background(), 1, 2)
		require.ErrorIs(t, err, dbErr)
	})

	t.Run("success", func(t *testing.T) {
		f := newFixture(t)
		f.repo.EXPECT().SetDefault(mock.Anything, uint(1), uint(2)).Return(nil).Once()

		err := f.uc.SetDefaultAddress(context.Background(), 1, 2)
		require.NoError(t, err)
	})
}
