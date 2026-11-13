package profile

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	domainErrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/domain/models"
	"ecommerce/internal/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

type profileFixture struct {
	uc          *ProfileUseCase
	userRepo    *mocks.MockUserRepository
	otpStore    *mocks.MockOTPstore
	emailSender *mocks.MockEmailSender
	storage     *mocks.MockObjectStorage
}

func newProfileFixture(t *testing.T) *profileFixture {
	t.Helper()
	userRepo := mocks.NewMockUserRepository(t)
	otpStore := mocks.NewMockOTPstore(t)
	emailSender := mocks.NewMockEmailSender(t)
	storage := mocks.NewMockObjectStorage(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	return &profileFixture{
		uc:          NewProfileUseCase(userRepo, otpStore, emailSender, storage, log),
		userRepo:    userRepo,
		otpStore:    otpStore,
		emailSender: emailSender,
		storage:     storage,
	}
}

func TestProfileUseCase_GetAndUpdateProfile(t *testing.T) {
	ctx := context.Background()

	t.Run("GetProfile", func(t *testing.T) {
		f := newProfileFixture(t)
		f.userRepo.EXPECT().FindByID(ctx, uint(0)).Return(nil, domainErrors.ErrUserNotFound).Once()
		_, err := f.uc.GetProfile(ctx, 0)
		assert.ErrorIs(t, err, domainErrors.ErrUserNotFound)

		expected := &models.User{ID: 7, Name: "Niyas"}
		f.userRepo.EXPECT().FindByID(ctx, uint(7)).Return(expected, nil).Once()
		got, err := f.uc.GetProfile(ctx, 7)
		require.NoError(t, err)
		assert.Equal(t, expected, got)
	})

	t.Run("UpdateProfile validation and success", func(t *testing.T) {
		f := newProfileFixture(t)
		f.userRepo.EXPECT().FindByID(ctx, uint(0)).Return(nil, domainErrors.ErrUserNotFound).Once()
		_, err := f.uc.UpdateProfile(ctx, 0, "Niyas", "9876543210")
		assert.ErrorIs(t, err, domainErrors.ErrUserNotFound)

		existing := &models.User{ID: 7, Name: "Old", Phone: "9999999999"}
		f.userRepo.EXPECT().FindByID(ctx, uint(7)).Return(existing, nil).Maybe()

		_, err = f.uc.UpdateProfile(ctx, 7, "N", "9876543210")
		assert.ErrorIs(t, err, domainErrors.ErrInvalidName)

		_, err = f.uc.UpdateProfile(ctx, 7, "Niyas", "12345")
		assert.ErrorIs(t, err, domainErrors.ErrInvalidPhone)

		updated := &models.User{ID: 7, Name: "Niyas M", Phone: "9876543210"}
		f.userRepo.EXPECT().UpdateProfile(ctx, uint(7), "Niyas M", "9876543210").Return(nil).Once()
		f.userRepo.EXPECT().FindByID(ctx, uint(7)).Return(updated, nil).Maybe()

		got, err := f.uc.UpdateProfile(ctx, 7, "Niyas M", "9876543210")
		require.NoError(t, err)
		assert.Equal(t, "Niyas M", got.Name)
		assert.Equal(t, "9876543210", got.Phone)
	})
}

func TestProfileUseCase_ChangePassword(t *testing.T) {
	ctx := context.Background()
	hash, err := bcrypt.GenerateFromPassword([]byte("OldPass@1"), bcrypt.MinCost)
	require.NoError(t, err)
	hashStr := string(hash)

	t.Run("validation and wrong current password", func(t *testing.T) {
		f := newProfileFixture(t)
		assert.ErrorIs(t, f.uc.ChangePassword(ctx, 0, "OldPass@1", "NewPass@1"), domainErrors.ErrInvalidUserID)
		assert.ErrorIs(t, f.uc.ChangePassword(ctx, 7, "OldPass@1", "weak"), domainErrors.ErrWeakPassword)

		f.userRepo.EXPECT().FindByID(ctx, uint(7)).Return(&models.User{ID: 7, Password: &hashStr}, nil).Once()
		assert.ErrorIs(t, f.uc.ChangePassword(ctx, 7, "WrongPass@1", "NewPass@1"), domainErrors.ErrInvalidCredentials)
	})

	t.Run("success", func(t *testing.T) {
		f := newProfileFixture(t)
		f.userRepo.EXPECT().FindByID(ctx, uint(7)).Return(&models.User{ID: 7, Password: &hashStr}, nil).Once()
		f.userRepo.EXPECT().UpdatePassword(ctx, uint(7), mock.MatchedBy(func(h string) bool {
			return bcrypt.CompareHashAndPassword([]byte(h), []byte("NewPass@1")) == nil
		})).Return(nil).Once()

		require.NoError(t, f.uc.ChangePassword(ctx, 7, "OldPass@1", "NewPass@1"))
	})
}

func TestProfileUseCase_EmailChangeOTP(t *testing.T) {
	ctx := context.Background()

	t.Run("SendEmailChangeOTP", func(t *testing.T) {
		f := newProfileFixture(t)
		assert.ErrorIs(t, f.uc.SendEmailChangeOTP(ctx, 0, "new@example.com"), domainErrors.ErrInvalidUserID)
		assert.Error(t, f.uc.SendEmailChangeOTP(ctx, 7, "   "))

		f.userRepo.EXPECT().FindByID(ctx, uint(7)).Return(&models.User{ID: 7, Email: "old@example.com"}, nil).Once()
		assert.Error(t, f.uc.SendEmailChangeOTP(ctx, 7, "OLD@example.com"))

		f.userRepo.EXPECT().FindByID(ctx, uint(7)).Return(&models.User{ID: 7, Email: "old@example.com"}, nil).Once()
		f.userRepo.EXPECT().FindByEmail(ctx, "taken@example.com").Return(&models.User{ID: 9}, nil).Once()
		assert.ErrorIs(t, f.uc.SendEmailChangeOTP(ctx, 7, "taken@example.com"), domainErrors.ErrEmailAlreadyExists)

		f.userRepo.EXPECT().FindByID(ctx, uint(7)).Return(&models.User{ID: 7, Email: "old@example.com"}, nil).Once()
		f.userRepo.EXPECT().FindByEmail(ctx, "new@example.com").Return(nil, domainErrors.ErrUserNotFound).Once()
		f.otpStore.EXPECT().Set(ctx, emailChangeOTPKey(7), mock.AnythingOfType("string"), mock.MatchedBy(func(d time.Duration) bool {
			return d > 0
		})).Return(nil).Once()
		f.otpStore.EXPECT().Set(ctx, emailChangePendingEmailKey(7), "new@example.com", mock.MatchedBy(func(d time.Duration) bool {
			return d > 0
		})).Return(nil).Once()
		f.emailSender.EXPECT().SendVerificationOTP(ctx, "new@example.com", mock.AnythingOfType("string")).Return(nil).Once()

		require.NoError(t, f.uc.SendEmailChangeOTP(ctx, 7, "NEW@example.com"))
	})

	t.Run("VerifyEmailChangeOTP", func(t *testing.T) {
		f := newProfileFixture(t)
		assert.ErrorIs(t, f.uc.VerifyEmailChangeOTP(ctx, 0, "123456"), domainErrors.ErrInvalidCredentials)
		assert.Error(t, f.uc.VerifyEmailChangeOTP(ctx, 7, ""))

		f.otpStore.EXPECT().Get(ctx, emailChangeOTPKey(7)).Return("012345", nil).Once()
		assert.ErrorIs(t, f.uc.VerifyEmailChangeOTP(ctx, 7, "999999"), domainErrors.ErrInvalidOTP)

		f.otpStore.EXPECT().Get(ctx, emailChangeOTPKey(7)).Return("012345", nil).Once()
		f.otpStore.EXPECT().Get(ctx, emailChangePendingEmailKey(7)).Return("new@example.com", nil).Once()
		f.userRepo.EXPECT().FindByEmail(ctx, "new@example.com").Return(nil, domainErrors.ErrUserNotFound).Once()
		f.userRepo.EXPECT().UpdateEmail(ctx, uint(7), "new@example.com").Return(nil).Once()
		f.userRepo.EXPECT().UpdateEmailVerifiedAt(ctx, uint(7), mock.AnythingOfType("time.Time")).Return(nil).Once()
		f.otpStore.EXPECT().Delete(ctx, emailChangeOTPKey(7)).Return(nil).Once()
		f.otpStore.EXPECT().Delete(ctx, emailChangePendingEmailKey(7)).Return(nil).Once()

		require.NoError(t, f.uc.VerifyEmailChangeOTP(ctx, 7, "012345"))
	})
}

func TestProfileUseCase_ProfileImage(t *testing.T) {
	ctx := context.Background()

	t.Run("UploadProfileImage validation and success", func(t *testing.T) {
		f := newProfileFixture(t)
		_, err := f.uc.UploadProfileImage(ctx, 0, bytes.NewReader([]byte("x")), 1, "image/png")
		assert.ErrorIs(t, err, domainErrors.ErrInvalidUserID)

		_, err = f.uc.UploadProfileImage(ctx, 7, bytes.NewReader([]byte("x")), 1, "application/pdf")
		assert.ErrorIs(t, err, domainErrors.ErrInvalidProfileImage)

		oldImg := "https://cdn.example.com/old.png"
		f.userRepo.EXPECT().FindByID(ctx, uint(7)).Return(&models.User{ID: 7, ProfileImage: oldImg}, nil).Once()
		f.storage.EXPECT().Upload(ctx, mock.AnythingOfType("string"), mock.Anything, int64(4), "image/png").
			Return("https://cdn.example.com/new.png", nil).Once()
		f.userRepo.EXPECT().UpdateProfileImage(ctx, uint(7), "https://cdn.example.com/new.png").Return(nil).Once()
		f.storage.EXPECT().Delete(ctx, mock.AnythingOfType("string")).Return(nil).Maybe()

		gotURL, err := f.uc.UploadProfileImage(ctx, 7, bytes.NewReader([]byte("data")), 4, "image/png")
		require.NoError(t, err)
		assert.Equal(t, "https://cdn.example.com/new.png", gotURL)
	})

	t.Run("DeleteProfileImage", func(t *testing.T) {
		f := newProfileFixture(t)
		assert.ErrorIs(t, f.uc.DeleteProfileImage(ctx, 0), domainErrors.ErrInvalidUserID)

		img := "https://cdn.example.com/avatar.png"
		f.userRepo.EXPECT().FindByID(ctx, uint(7)).Return(&models.User{ID: 7, ProfileImage: img}, nil).Once()
		f.storage.EXPECT().Delete(ctx, "profiles/7/profile").Return(nil).Once()
		f.userRepo.EXPECT().UpdateProfileImage(ctx, uint(7), "").Return(nil).Once()

		require.NoError(t, f.uc.DeleteProfileImage(ctx, 7))
	})

	t.Run("UpdateProfileImage failure", func(t *testing.T) {
		f := newProfileFixture(t)
		dbErr := errors.New("db failure")
		f.userRepo.EXPECT().FindByID(ctx, uint(7)).Return(&models.User{ID: 7}, nil).Once()
		f.storage.EXPECT().Upload(ctx, mock.AnythingOfType("string"), mock.Anything, int64(4), "image/jpeg").
			Return("https://cdn.example.com/img.jpg", nil).Once()
		f.userRepo.EXPECT().UpdateProfileImage(ctx, uint(7), "https://cdn.example.com/img.jpg").Return(dbErr).Once()
		f.storage.EXPECT().Delete(ctx, "profiles/7/profile").Return(nil).Once()

		_, err := f.uc.UploadProfileImage(ctx, 7, bytes.NewReader([]byte("data")), 4, "image/jpeg")
		assert.ErrorIs(t, err, dbErr)
	})
}
