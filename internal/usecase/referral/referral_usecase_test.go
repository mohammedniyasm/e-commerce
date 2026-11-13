package referral

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	domainErrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/domain/models"
	"ecommerce/internal/mocks"
	"ecommerce/internal/usecase/interfaces"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type referralFixture struct {
	uc           interfaces.ReferralUseCase
	userRepo     *mocks.MockUserRepository
	referralRepo *mocks.MockReferralRepository
	settingsRepo *mocks.MockReferralSettingsRepository
}

func newReferralFixture(t *testing.T) *referralFixture {
	t.Helper()
	userRepo := mocks.NewMockUserRepository(t)
	referralRepo := mocks.NewMockReferralRepository(t)
	settingsRepo := mocks.NewMockReferralSettingsRepository(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	return &referralFixture{
		uc:           NewReferralUseCase(userRepo, referralRepo, settingsRepo, log),
		userRepo:     userRepo,
		referralRepo: referralRepo,
		settingsRepo: settingsRepo,
	}
}

func TestReferralUseCase_GetReferralDashboard(t *testing.T) {
	ctx := context.Background()
	userID := uint(12)

	t.Run("user lookup error", func(t *testing.T) {
		f := newReferralFixture(t)
		f.userRepo.EXPECT().FindByID(ctx, userID).Return(nil, domainErrors.ErrUserNotFound).Once()

		code, total, completed, pending, rewards, list, count, err := f.uc.GetReferralDashboard(ctx, userID, nil, nil, 10, 0)
		assert.Empty(t, code)
		assert.Zero(t, total+completed+pending+count)
		assert.Zero(t, rewards)
		assert.Nil(t, list)
		assert.ErrorIs(t, err, domainErrors.ErrUserNotFound)
	})

	t.Run("counts error", func(t *testing.T) {
		f := newReferralFixture(t)
		dbErr := errors.New("counts failed")
		f.userRepo.EXPECT().FindByID(ctx, userID).Return(&models.User{ID: userID, ReferralCode: "REF123"}, nil).Once()
		f.referralRepo.EXPECT().GetReferralCounts(ctx, userID).Return(int64(0), int64(0), int64(0), dbErr).Once()

		_, _, _, _, _, _, _, err := f.uc.GetReferralDashboard(ctx, userID, nil, nil, 10, 0)
		assert.ErrorIs(t, err, dbErr)
	})

	t.Run("total rewards error", func(t *testing.T) {
		f := newReferralFixture(t)
		dbErr := errors.New("rewards failed")
		f.userRepo.EXPECT().FindByID(ctx, userID).Return(&models.User{ID: userID, ReferralCode: "REF123"}, nil).Once()
		f.referralRepo.EXPECT().GetReferralCounts(ctx, userID).Return(int64(3), int64(2), int64(1), nil).Once()
		f.referralRepo.EXPECT().GetTotalRewardsEarned(ctx, userID).Return(float64(0), dbErr).Once()

		_, _, _, _, _, _, _, err := f.uc.GetReferralDashboard(ctx, userID, nil, nil, 10, 0)
		assert.ErrorIs(t, err, dbErr)
	})

	t.Run("list referrals error", func(t *testing.T) {
		f := newReferralFixture(t)
		dbErr := errors.New("list failed")
		f.userRepo.EXPECT().FindByID(ctx, userID).Return(&models.User{ID: userID, ReferralCode: "REF123"}, nil).Once()
		f.referralRepo.EXPECT().GetReferralCounts(ctx, userID).Return(int64(3), int64(2), int64(1), nil).Once()
		f.referralRepo.EXPECT().GetTotalRewardsEarned(ctx, userID).Return(float64(200), nil).Once()
		f.referralRepo.EXPECT().GetByReferrerID(ctx, userID, (*models.ReferralStatus)(nil), (*models.ReferralRewardStatus)(nil), 10, 0).
			Return(nil, int64(0), dbErr).Once()

		_, _, _, _, _, _, _, err := f.uc.GetReferralDashboard(ctx, userID, nil, nil, 10, 0)
		assert.ErrorIs(t, err, dbErr)
	})

	t.Run("success", func(t *testing.T) {
		f := newReferralFixture(t)
		st := models.ReferralStatus("completed")
		rst := models.ReferralRewardStatus("credited")
		referrals := []models.Referral{{ID: 1}}

		f.userRepo.EXPECT().FindByID(ctx, userID).Return(&models.User{ID: userID, ReferralCode: "REF123"}, nil).Once()
		f.referralRepo.EXPECT().GetReferralCounts(ctx, userID).Return(int64(5), int64(3), int64(2), nil).Once()
		f.referralRepo.EXPECT().GetTotalRewardsEarned(ctx, userID).Return(150.5, nil).Once()
		f.referralRepo.EXPECT().GetByReferrerID(ctx, userID, &st, &rst, 20, 10).Return(referrals, int64(5), nil).Once()

		code, total, pending, completed, rewards, gotList, gotCount, err := f.uc.GetReferralDashboard(ctx, userID, &st, &rst, 20, 10)
		require.NoError(t, err)
		assert.Equal(t, "REF123", code)
		assert.Equal(t, int64(5), total)
		assert.Equal(t, int64(3), pending)
		assert.Equal(t, int64(2), completed)
		assert.Equal(t, 150.5, rewards)
		assert.Equal(t, referrals, gotList)
		assert.Equal(t, int64(5), gotCount)
	})
}

func TestReferralUseCase_AdminAndQueries(t *testing.T) {
	ctx := context.Background()

	t.Run("GetAdminReferralDashboard", func(t *testing.T) {
		f := newReferralFixture(t)
		dbErr := errors.New("stats err")
		f.referralRepo.EXPECT().GetReferralDashboardStats(ctx).Return(int64(0), int64(0), int64(0), float64(0), float64(0), dbErr).Once()
		_, _, _, _, _, _, _, err := f.uc.GetAdminReferralDashboard(ctx)
		assert.ErrorIs(t, err, dbErr)

		f.referralRepo.EXPECT().GetReferralDashboardStats(ctx).Return(int64(10), int64(7), int64(3), float64(700), float64(300), nil).Once()

		tot, pend, comp, refRew, refdRew, totalPaid, convRate, err := f.uc.GetAdminReferralDashboard(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(10), tot)
		assert.Equal(t, int64(7), pend)
		assert.Equal(t, int64(3), comp)
		assert.Equal(t, 700.0, refRew)
		assert.Equal(t, 300.0, refdRew)
		assert.Equal(t, 1000.0, totalPaid)
		assert.Equal(t, 30.0, convRate)

		f.referralRepo.EXPECT().GetReferralDashboardStats(ctx).Return(int64(0), int64(0), int64(0), float64(0), float64(0), nil).Once()
		tot, _, _, _, _, _, convRate, err = f.uc.GetAdminReferralDashboard(ctx)
		require.NoError(t, err)
		assert.Zero(t, tot)
		assert.Zero(t, convRate)
	})

	t.Run("GetTopReferrers", func(t *testing.T) {
		f := newReferralFixture(t)
		expected := []models.TopReferrer{{}}
		f.referralRepo.EXPECT().GetTopReferrers(ctx, 5).Return(expected, nil).Once()

		got, err := f.uc.GetTopReferrers(ctx, 5)
		require.NoError(t, err)
		assert.Equal(t, expected, got)
	})

	t.Run("GetReferrals", func(t *testing.T) {
		f := newReferralFixture(t)
		now := time.Now()
		uid := uint(9)
		expected := []models.Referral{{ID: 2}}
		f.referralRepo.EXPECT().GetReferrals(ctx, (*models.ReferralStatus)(nil), (*models.ReferralRewardStatus)(nil), &uid, &now, &now, 15, 5).
			Return(expected, int64(1), nil).Once()

		got, count, err := f.uc.GetReferrals(ctx, nil, nil, &uid, &now, &now, 15, 5)
		require.NoError(t, err)
		assert.Equal(t, expected, got)
		assert.Equal(t, int64(1), count)
	})
}

func TestReferralUseCase_Settings(t *testing.T) {
	ctx := context.Background()

	t.Run("GetReferralSettings", func(t *testing.T) {
		f := newReferralFixture(t)
		expected := &models.ReferralSettings{ReferrerReward: 100}
		f.settingsRepo.EXPECT().Get(ctx).Return(expected, nil).Once()

		got, err := f.uc.GetReferralSettings(ctx)
		require.NoError(t, err)
		assert.Equal(t, expected, got)
	})

	t.Run("UpdateReferralSettings validation errors", func(t *testing.T) {
		tests := []struct {
			name     string
			referrer float64
			referred float64
			minOrder float64
		}{
			{name: "negative referrer reward", referrer: -1, referred: 50, minOrder: 500},
			{name: "negative referred reward", referrer: 100, referred: -1, minOrder: 500},
			{name: "negative minimum order", referrer: 100, referred: 50, minOrder: -10},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				f := newReferralFixture(t)
				got, err := f.uc.UpdateReferralSettings(ctx, tc.referrer, tc.referred, tc.minOrder, true)
				assert.Nil(t, got)
				assert.ErrorIs(t, err, domainErrors.ErrInvalidReferralSettings)
			})
		}
	})

	t.Run("UpdateReferralSettings repo errors and success", func(t *testing.T) {
		f := newReferralFixture(t)
		getErr := errors.New("get failed")
		f.settingsRepo.EXPECT().Get(ctx).Return(nil, getErr).Once()
		got, err := f.uc.UpdateReferralSettings(ctx, 100, 50, 500, true)
		assert.Nil(t, got)
		assert.ErrorIs(t, err, getErr)

		existing := &models.ReferralSettings{ReferrerReward: 50, ReferredReward: 25, MinimumOrderAmount: 200, IsActive: false}
		updateErr := errors.New("update failed")
		f.settingsRepo.EXPECT().Get(ctx).Return(existing, nil).Once()
		f.settingsRepo.EXPECT().Update(ctx, existing).Return(updateErr).Once()
		got, err = f.uc.UpdateReferralSettings(ctx, 100, 50, 500, true)
		assert.Nil(t, got)
		assert.ErrorIs(t, err, updateErr)

		existing2 := &models.ReferralSettings{}
		f.settingsRepo.EXPECT().Get(ctx).Return(existing2, nil).Once()
		f.settingsRepo.EXPECT().Update(ctx, existing2).Return(nil).Once()
		got, err = f.uc.UpdateReferralSettings(ctx, 150, 75, 1000, true)
		require.NoError(t, err)
		assert.Equal(t, 150.0, got.ReferrerReward)
		assert.Equal(t, 75.0, got.ReferredReward)
		assert.Equal(t, 1000.0, got.MinimumOrderAmount)
		assert.True(t, got.IsActive)
	})
}
