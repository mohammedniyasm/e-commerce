package notification

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	domainErrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/domain/models"
	"ecommerce/internal/mocks"
	"ecommerce/internal/usecase/interfaces"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type notificationFixture struct {
	uc        *NotificationUseCase
	notifRepo *mocks.MockNotificationRepository
	uow       *mocks.MockUnitOfWork
	txUser    *mocks.MockUserRepository
	txNotif   *mocks.MockNotificationRepository
}

func newNotificationFixture(t *testing.T) *notificationFixture {
	t.Helper()
	notifRepo := mocks.NewMockNotificationRepository(t)
	uow := mocks.NewMockUnitOfWork(t)
	txUser := mocks.NewMockUserRepository(t)
	txNotif := mocks.NewMockNotificationRepository(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	return &notificationFixture{
		uc:        NewNotificationUseCase(notifRepo, uow, log),
		notifRepo: notifRepo,
		uow:       uow,
		txUser:    txUser,
		txNotif:   txNotif,
	}
}

func (f *notificationFixture) expectTX(ctx context.Context) {
	f.uow.EXPECT().WithinTransaction(ctx, mock.AnythingOfType("func(interfaces.TransactionRepositories) error")).
		RunAndReturn(func(_ context.Context, fn func(interfaces.TransactionRepositories) error) error {
			return fn(interfaces.TransactionRepositories{
				User:         f.txUser,
				Notification: f.txNotif,
			})
		}).Once()
}

func TestNotificationUseCase_CreateNotification(t *testing.T) {
	ctx := context.Background()

	t.Run("nil notification returns invalid input", func(t *testing.T) {
		f := newNotificationFixture(t)
		err := f.uc.CreateNotification(ctx, nil)
		assert.ErrorIs(t, err, domainErrors.ErrInvalidInput)
	})

	t.Run("ListAdminIDs failure aborts tx", func(t *testing.T) {
		f := newNotificationFixture(t)
		dbErr := errors.New("admin query error")
		f.expectTX(ctx)
		f.txUser.EXPECT().ListAdminIDs(ctx).Return(nil, dbErr).Once()

		err := f.uc.CreateNotification(ctx, &models.Notification{})
		assert.ErrorIs(t, err, dbErr)
	})

	t.Run("Create failure aborts tx", func(t *testing.T) {
		f := newNotificationFixture(t)
		dbErr := errors.New("insert notification failed")
		f.expectTX(ctx)
		f.txUser.EXPECT().ListAdminIDs(ctx).Return([]uint{1, 2}, nil).Once()
		f.txNotif.EXPECT().Create(ctx, mock.AnythingOfType("*models.Notification")).Return(dbErr).Once()

		n := &models.Notification{}
		err := f.uc.CreateNotification(ctx, n)
		assert.ErrorIs(t, err, dbErr)
		assert.NotEqual(t, uuid.Nil, n.ID)
	})

	t.Run("CreateRecipients failure returns error", func(t *testing.T) {
		f := newNotificationFixture(t)
		dbErr := errors.New("insert recipients failed")
		fixedID := uuid.New()
		n := &models.Notification{ID: fixedID}

		f.expectTX(ctx)
		f.txUser.EXPECT().ListAdminIDs(ctx).Return([]uint{10}, nil).Once()
		f.txNotif.EXPECT().Create(ctx, n).Return(nil).Once()
		f.txNotif.EXPECT().CreateRecipients(ctx, mock.MatchedBy(func(r []models.NotificationRecipient) bool {
			return len(r) == 1 && r[0].AdminID == 10 && r[0].NotificationID == fixedID
		})).Return(dbErr).Once()

		err := f.uc.CreateNotification(ctx, n)
		assert.ErrorIs(t, err, dbErr)
	})

	t.Run("no admins still creates notification and empty recipients slice", func(t *testing.T) {
		f := newNotificationFixture(t)
		n := &models.Notification{}

		f.expectTX(ctx)
		f.txUser.EXPECT().ListAdminIDs(ctx).Return([]uint{}, nil).Once()
		f.txNotif.EXPECT().Create(ctx, n).Return(nil).Once()
		f.txNotif.EXPECT().CreateRecipients(ctx, mock.MatchedBy(func(r []models.NotificationRecipient) bool {
			return len(r) == 0
		})).Return(nil).Once()

		err := f.uc.CreateNotification(ctx, n)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, n.ID)
	})

	t.Run("success assigns uuid when nil and creates recipients for all admins", func(t *testing.T) {
		f := newNotificationFixture(t)
		n := &models.Notification{}

		f.expectTX(ctx)
		f.txUser.EXPECT().ListAdminIDs(ctx).Return([]uint{3, 7}, nil).Once()
		f.txNotif.EXPECT().Create(ctx, n).Return(nil).Once()
		f.txNotif.EXPECT().CreateRecipients(ctx, mock.MatchedBy(func(r []models.NotificationRecipient) bool {
			return len(r) == 2 &&
				r[0].AdminID == 3 && r[0].NotificationID == n.ID &&
				r[1].AdminID == 7 && r[1].NotificationID == n.ID
		})).Return(nil).Once()

		err := f.uc.CreateNotification(ctx, n)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, n.ID)
	})
}

func TestNotificationUseCase_GetByID(t *testing.T) {
	ctx := context.Background()

	t.Run("nil uuid returns invalid input", func(t *testing.T) {
		f := newNotificationFixture(t)
		got, err := f.uc.GetByID(ctx, uuid.Nil)
		assert.Nil(t, got)
		assert.ErrorIs(t, err, domainErrors.ErrInvalidInput)
	})

	t.Run("success", func(t *testing.T) {
		f := newNotificationFixture(t)
		id := uuid.New()
		expected := &models.Notification{ID: id}
		f.notifRepo.EXPECT().GetByID(ctx, id).Return(expected, nil).Once()

		got, err := f.uc.GetByID(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, expected, got)
	})

	t.Run("repo error", func(t *testing.T) {
		f := newNotificationFixture(t)
		id := uuid.New()
		f.notifRepo.EXPECT().GetByID(ctx, id).Return(nil, domainErrors.ErrRecordNotFound).Once()

		got, err := f.uc.GetByID(ctx, id)
		assert.Nil(t, got)
		assert.ErrorIs(t, err, domainErrors.ErrRecordNotFound)
	})
}

func TestNotificationUseCase_ListByAdminID(t *testing.T) {
	ctx := context.Background()

	t.Run("zero adminID returns invalid input", func(t *testing.T) {
		f := newNotificationFixture(t)
		got, total, err := f.uc.ListByAdminID(ctx, 0, 10, 1, false)
		assert.Nil(t, got)
		assert.Zero(t, total)
		assert.ErrorIs(t, err, domainErrors.ErrInvalidInput)
	})

	t.Run("defaults limit and page when <= 0 and preserves large limit", func(t *testing.T) {
		tests := []struct {
			name        string
			limit       int
			page        int
			expectLimit int
			expectPage  int
			unreadOnly  bool
		}{
			{name: "zero values default", limit: 0, page: 0, expectLimit: 20, expectPage: 1, unreadOnly: false},
			{name: "negative values default", limit: -5, page: -2, expectLimit: 20, expectPage: 1, unreadOnly: true},
			{name: "positive values unchanged without max cap", limit: 150, page: 3, expectLimit: 150, expectPage: 3, unreadOnly: true},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				f := newNotificationFixture(t)
				expected := []models.NotificationRecipient{{AdminID: 5}}
				f.notifRepo.EXPECT().ListByAdminID(ctx, uint(5), tc.expectLimit, tc.expectPage, tc.unreadOnly).
					Return(expected, int64(1), nil).Once()

				got, total, err := f.uc.ListByAdminID(ctx, 5, tc.limit, tc.page, tc.unreadOnly)
				require.NoError(t, err)
				assert.Equal(t, expected, got)
				assert.Equal(t, int64(1), total)
			})
		}
	})

	t.Run("repo error", func(t *testing.T) {
		f := newNotificationFixture(t)
		dbErr := errors.New("list failed")
		f.notifRepo.EXPECT().ListByAdminID(ctx, uint(5), 20, 1, false).Return(nil, int64(0), dbErr).Once()

		got, total, err := f.uc.ListByAdminID(ctx, 5, 0, 0, false)
		assert.Nil(t, got)
		assert.Zero(t, total)
		assert.ErrorIs(t, err, dbErr)
	})
}

func TestNotificationUseCase_CountUnreadAndMarkRead(t *testing.T) {
	ctx := context.Background()

	t.Run("CountUnreadByAdminID", func(t *testing.T) {
		f := newNotificationFixture(t)
		count, err := f.uc.CountUnreadByAdminID(ctx, 0)
		assert.Zero(t, count)
		assert.ErrorIs(t, err, domainErrors.ErrInvalidInput)

		f.notifRepo.EXPECT().CountUnreadByAdminID(ctx, uint(2)).Return(int64(4), nil).Once()
		count, err = f.uc.CountUnreadByAdminID(ctx, 2)
		require.NoError(t, err)
		assert.Equal(t, int64(4), count)
	})

	t.Run("MarkAsRead", func(t *testing.T) {
		id := uuid.New()
		f := newNotificationFixture(t)

		assert.ErrorIs(t, f.uc.MarkAsRead(ctx, 0, id), domainErrors.ErrInvalidInput)
		assert.ErrorIs(t, f.uc.MarkAsRead(ctx, 2, uuid.Nil), domainErrors.ErrInvalidInput)

		f.notifRepo.EXPECT().MarkAsRead(ctx, uint(2), id).Return(nil).Once()
		require.NoError(t, f.uc.MarkAsRead(ctx, 2, id))
	})

	t.Run("MarkAllAsRead", func(t *testing.T) {
		f := newNotificationFixture(t)
		assert.ErrorIs(t, f.uc.MarkAllAsRead(ctx, 0), domainErrors.ErrInvalidInput)

		f.notifRepo.EXPECT().MarkAllAsRead(ctx, uint(2)).Return(nil).Once()
		require.NoError(t, f.uc.MarkAllAsRead(ctx, 2))
	})
}
