package auth

import (
	"context"
	domainerrors "ecommerce/internal/domain/errors"
	"ecommerce/internal/domain/models"
	"ecommerce/internal/mocks"
	"ecommerce/internal/usecase/interfaces"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

// import (
//
//	"context"
//	domainerrors "ecommerce/internal/domain/errors"
//	"ecommerce/internal/domain/models"
//	"ecommerce/internal/mocks"
//	"ecommerce/internal/usecase/interfaces"
//	"errors"
//	"io"
//	"log/slog"
//	"testing"
//	"time"
//	"github.com/stretchr/testify/mock"
//	"github.com/stretchr/testify/require"
//	"golang.org/x/crypto/bcrypt"
//
// )
// func setupAuthTest(t *testing.T) (
//
//	context.Context,
//	*mocks.MockUserRepository,
//	*mocks.MockOTPstore,
//	*mocks.MockJWTService,
//	*mocks.MockRefreshSessionStore,
//	*mocks.MockAccessTokenBlacklist,
//	*mocks.MockEmailSender,
//	*mocks.MockGoogleTokenVerifier,
//	*mocks.MockUnitOfWork,
//	*AuthUseCase,
//
//	) {
//		ctx := context.Background()
//		userRepo := mocks.NewMockUserRepository(t)
//		otpStore := mocks.NewMockOTPstore(t)
//		jwtService := mocks.NewMockJWTService(t)
//		refreshStore := mocks.NewMockRefreshSessionStore(t)
//		blacklist := mocks.NewMockAccessTokenBlacklist(t)
//		emailSender := mocks.NewMockEmailSender(t)
//		googleVerifier := mocks.NewMockGoogleTokenVerifier(t)
//		uow := mocks.NewMockUnitOfWork(t)
//		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
//		uc := NewAuthUseCase(
//			userRepo,
//			otpStore,
//			jwtService,
//			refreshStore,
//			blacklist,
//			emailSender,
//			googleVerifier,
//			uow,
//			logger,
//		)
//		return ctx, userRepo, otpStore, jwtService, refreshStore, blacklist, emailSender, googleVerifier, uow, uc
//	}
//
//	func TestRegister(t *testing.T) {
//		validPassword := "Niyas@123"
//		t.Run("Success Without Referral", func(t *testing.T) {
//			ctx, userRepo, _, _, _, _, _, _, uow, uc := setupAuthTest(t)
//			txUserRepo := mocks.NewMockUserRepository(t)
//			user := &models.User{
//				Name:     "Mohammed Niyas",
//				Phone:    "8848598156",
//				Email:    "test@example.com",
//				Password: &validPassword,
//			}
//			userRepo.EXPECT().FindByEmail(ctx, "test@example.com").Return(nil, domainerrors.ErrUserNotFound).Once()
//			uow.EXPECT().WithinTransaction(ctx, mock.Anything).RunAndReturn(
//				func(ctx context.Context, fn func(repos interfaces.TransactionRepositories) error) error {
//					return fn(interfaces.TransactionRepositories{
//						User: txUserRepo,
//					})
//				},
//			).Once()
//			txUserRepo.EXPECT().Create(ctx, mock.MatchedBy(func(u *models.User) bool {
//				return u.Email == "test@example.com" &&
//					u.Role == models.RoleUser &&
//					len(u.ReferralCode) > 0 &&
//					u.Password != nil &&
//					*u.Password != validPassword // Password must be hashed
//			})).Return(nil).Once()
//			createdUser, err := uc.Register(ctx, user, "")
//			require.NoError(t, err)
//			require.NotNil(t, createdUser)
//			require.Equal(t, models.RoleUser, createdUser.Role)
//		})
//		t.Run("Invalid Name", func(t *testing.T) {
//			ctx, _, _, _, _, _, _, _, _, uc := setupAuthTest(t)
//			user := &models.User{
//				Name:     "123@#$",
//				Phone:    "8848598156",
//				Email:    "test@example.com",
//				Password: &validPassword,
//			}
//			_, err := uc.Register(ctx, user, "")
//			require.ErrorIs(t, err, domainerrors.ErrInvalidName)
//		})
//		t.Run("Invalid Phone", func(t *testing.T) {
//			ctx, _, _, _, _, _, _, _, _, uc := setupAuthTest(t)
//			user := &models.User{
//				Name:     "Mohammed Niyas",
//				Phone:    "123",
//				Email:    "test@example.com",
//				Password: &validPassword,
//			}
//			_, err := uc.Register(ctx, user, "")
//			require.ErrorIs(t, err, domainerrors.ErrInvalidPhone)
//		})
//		t.Run("Weak Password", func(t *testing.T) {
//			ctx, _, _, _, _, _, _, _, _, uc := setupAuthTest(t)
//			weakPass := "123"
//			user := &models.User{
//				Name:     "Mohammed Niyas",
//				Phone:    "8848598156",
//				Email:    "test@example.com",
//				Password: &weakPass,
//			}
//			_, err := uc.Register(ctx, user, "")
//			require.ErrorIs(t, err, domainerrors.ErrWeakPassword)
//		})
//		t.Run("Email Already Exists", func(t *testing.T) {
//			ctx, userRepo, _, _, _, _, _, _, _, uc := setupAuthTest(t)
//			user := &models.User{
//				Name:     "Mohammed Niyas",
//				Phone:    "8848598156",
//				Email:    "existing@example.com",
//				Password: &validPassword,
//			}
//			userRepo.EXPECT().FindByEmail(ctx, "existing@example.com").Return(&models.User{ID: 1}, nil).Once()
//			_, err := uc.Register(ctx, user, "")
//			require.ErrorIs(t, err, domainerrors.ErrEmailAlreadyExists)
//		})
//	}
//
//	func TestLogin(t *testing.T) {
//		rawPassword := "Niyas@123"
//		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(rawPassword), bcrypt.DefaultCost)
//		hashedPassStr := string(hashedPassword)
//		t.Run("Success", func(t *testing.T) {
//			ctx, userRepo, _, jwtService, refreshStore, _, _, _, _, uc := setupAuthTest(t)
//			user := &models.User{
//				ID:        10,
//				Email:     "user@example.com",
//				Password:  &hashedPassStr,
//				Role:      models.RoleUser,
//				IsBlocked: false,
//			}
//			userRepo.EXPECT().FindByEmail(ctx, "user@example.com").Return(user, nil).Once()
//			jwtService.EXPECT().GenerateAccessToken(uint(10), "user").Return("access_token_123", nil).Once()
//			jwtService.EXPECT().GenerateRefreshToken(uint(10)).Return("refresh_token_456", "jti_789", nil).Once()
//			refreshStore.EXPECT().Save(ctx, "jti_789", uint(10)).Return(nil).Once()
//			accessToken, refreshToken, err := uc.Login(ctx, "user@example.com", rawPassword)
//			require.NoError(t, err)
//			require.Equal(t, "access_token_123", accessToken)
//			require.Equal(t, "refresh_token_456", refreshToken)
//		})
//		t.Run("Empty Credentials", func(t *testing.T) {
//			ctx, _, _, _, _, _, _, _, _, uc := setupAuthTest(t)
//			_, _, err := uc.Login(ctx, "", rawPassword)
//			require.ErrorIs(t, err, domainerrors.ErrInvalidCredentials)
//			_, _, err = uc.Login(ctx, "user@example.com", "")
//			require.ErrorIs(t, err, domainerrors.ErrInvalidCredentials)
//		})
//		t.Run("User Not Found", func(t *testing.T) {
//			ctx, userRepo, _, _, _, _, _, _, _, uc := setupAuthTest(t)
//			userRepo.EXPECT().FindByEmail(ctx, "notfound@example.com").Return(nil, domainerrors.ErrUserNotFound).Once()
//			_, _, err := uc.Login(ctx, "notfound@example.com", rawPassword)
//			require.ErrorIs(t, err, domainerrors.ErrInvalidCredentials)
//		})
//		t.Run("Wrong Password", func(t *testing.T) {
//			ctx, userRepo, _, _, _, _, _, _, _, uc := setupAuthTest(t)
//			user := &models.User{
//				ID:       10,
//				Email:    "user@example.com",
//				Password: &hashedPassStr,
//				Role:     models.RoleUser,
//			}
//			userRepo.EXPECT().FindByEmail(ctx, "user@example.com").Return(user, nil).Once()
//			_, _, err := uc.Login(ctx, "user@example.com", "WrongPassword@123")
//			require.ErrorIs(t, err, domainerrors.ErrInvalidCredentials)
//		})
//		t.Run("User Blocked", func(t *testing.T) {
//			ctx, userRepo, _, _, _, _, _, _, _, uc := setupAuthTest(t)
//			user := &models.User{
//				ID:        10,
//				Email:     "user@example.com",
//				Password:  &hashedPassStr,
//				Role:      models.RoleUser,
//				IsBlocked: true,
//			}
//			userRepo.EXPECT().FindByEmail(ctx, "user@example.com").Return(user, nil).Once()
//			_, _, err := uc.Login(ctx, "user@example.com", rawPassword)
//			require.ErrorIs(t, err, domainerrors.ErrUserBlocked)
//		})
//	}
//
//	func TestAdminLogin(t *testing.T) {
//		rawPassword := "Admin@123"
//		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(rawPassword), bcrypt.DefaultCost)
//		hashedPassStr := string(hashedPassword)
//		t.Run("Success Admin Login", func(t *testing.T) {
//			ctx, userRepo, _, jwtService, refreshStore, _, _, _, _, uc := setupAuthTest(t)
//			user := &models.User{
//				ID:       1,
//				Email:    "admin@example.com",
//				Password: &hashedPassStr,
//				Role:     models.RoleAdmin,
//			}
//			userRepo.EXPECT().FindByEmail(ctx, "admin@example.com").Return(user, nil).Once()
//			jwtService.EXPECT().GenerateAccessToken(uint(1), "admin").Return("admin_access", nil).Once()
//			jwtService.EXPECT().GenerateRefreshToken(uint(1)).Return("admin_refresh", "admin_jti", nil).Once()
//			refreshStore.EXPECT().Save(ctx, "admin_jti", uint(1)).Return(nil).Once()
//			acc, ref, err := uc.AdminLogin(ctx, "admin@example.com", rawPassword)
//			require.NoError(t, err)
//			require.Equal(t, "admin_access", acc)
//			require.Equal(t, "admin_refresh", ref)
//		})
//		t.Run("Non-Admin Attempting Admin Login", func(t *testing.T) {
//			ctx, userRepo, _, _, _, _, _, _, _, uc := setupAuthTest(t)
//			user := &models.User{
//				ID:       2,
//				Email:    "normaluser@example.com",
//				Password: &hashedPassStr,
//				Role:     models.RoleUser,
//			}
//			userRepo.EXPECT().FindByEmail(ctx, "normaluser@example.com").Return(user, nil).Once()
//			_, _, err := uc.AdminLogin(ctx, "normaluser@example.com", rawPassword)
//			require.ErrorIs(t, err, domainerrors.ErrInvalidCredentials)
//		})
//	}
//
//	func TestRefreshAccessToken(t *testing.T) {
//		t.Run("Success Token Rotation", func(t *testing.T) {
//			ctx, userRepo, _, jwtService, refreshStore, _, _, _, _, uc := setupAuthTest(t)
//			jwtService.EXPECT().ValidateRefreshToken("valid_refresh_token").Return("10", "old_jti", nil).Once()
//			refreshStore.EXPECT().Consume(ctx, "old_jti").Return(uint(10), true, nil).Once()
//			user := &models.User{
//				ID:        10,
//				Role:      models.RoleUser,
//				IsBlocked: false,
//			}
//			userRepo.EXPECT().FindByID(ctx, uint(10)).Return(user, nil).Once()
//			jwtService.EXPECT().GenerateAccessToken(uint(10), "user").Return("new_access_token", nil).Once()
//			jwtService.EXPECT().GenerateRefreshToken(uint(10)).Return("new_refresh_token", "new_jti", nil).Once()
//			refreshStore.EXPECT().Save(ctx, "new_jti", uint(10)).Return(nil).Once()
//			access, refresh, err := uc.RefreshAccessToken(ctx, "valid_refresh_token")
//			require.NoError(t, err)
//			require.Equal(t, "new_access_token", access)
//			require.Equal(t, "new_refresh_token", refresh)
//		})
//		t.Run("Consumed or Invalid Session", func(t *testing.T) {
//			ctx, _, _, jwtService, refreshStore, _, _, _, _, uc := setupAuthTest(t)
//			jwtService.EXPECT().ValidateRefreshToken("consumed_token").Return("10", "old_jti", nil).Once()
//			refreshStore.EXPECT().Consume(ctx, "old_jti").Return(uint(0), false, nil).Once()
//			_, _, err := uc.RefreshAccessToken(ctx, "consumed_token")
//			require.ErrorIs(t, err, domainerrors.ErrInvalidCredentials)
//		})
//		t.Run("User Blocked During Refresh", func(t *testing.T) {
//			ctx, userRepo, _, jwtService, refreshStore, _, _, _, _, uc := setupAuthTest(t)
//			jwtService.EXPECT().ValidateRefreshToken("token").Return("10", "jti", nil).Once()
//			refreshStore.EXPECT().Consume(ctx, "jti").Return(uint(10), true, nil).Once()
//			user := &models.User{
//				ID:        10,
//				Role:      models.RoleUser,
//				IsBlocked: true,
//			}
//			userRepo.EXPECT().FindByID(ctx, uint(10)).Return(user, nil).Once()
//			_, _, err := uc.RefreshAccessToken(ctx, "token")
//			require.ErrorIs(t, err, domainerrors.ErrUserBlocked)
//		})
//	}
//
//	func TestLogout(t *testing.T) {
//		t.Run("Success", func(t *testing.T) {
//			ctx, _, _, jwtService, refreshStore, blacklist, _, _, _, uc := setupAuthTest(t)
//			jwtService.EXPECT().ValidateRefreshToken("refresh_token").Return("10", "jti_123", nil).Once()
//			refreshStore.EXPECT().Delete(ctx, "jti_123").Return(nil).Once()
//			claims := &interfaces.AccessClaims{
//				UserID:    "10",
//				Role:      "user",
//				JTI:       "access_jti",
//				ExpiresAt: time.Now().Add(15 * time.Minute),
//			}
//			blacklist.EXPECT().Blacklist(ctx, "access_jti", mock.Anything).Return(nil).Once()
//			err := uc.Logout(ctx, "refresh_token", claims)
//			require.NoError(t, err)
//		})
//		t.Run("Empty Refresh Token", func(t *testing.T) {
//			ctx, _, _, _, _, _, _, _, _, uc := setupAuthTest(t)
//			err := uc.Logout(ctx, "", nil)
//			require.ErrorIs(t, err, domainerrors.ErrInvalidCredentials)
//		})
//	}
//
//	func TestSendVerificationOTP(t *testing.T) {
//		t.Run("Success", func(t *testing.T) {
//			ctx, userRepo, otpStore, _, _, _, emailSender, _, _, uc := setupAuthTest(t)
//			user := &models.User{
//				ID:              10,
//				Email:           "user@example.com",
//				EmailVerifiedAt: nil,
//			}
//			userRepo.EXPECT().FindByEmail(ctx, "user@example.com").Return(user, nil).Once()
//			otpStore.EXPECT().Set(ctx, "email_verification_otp:user@example.com", mock.Anything, 5*time.Minute).Return(nil).Once()
//			emailSender.EXPECT().SendVerificationOTP(ctx, "user@example.com", mock.Anything).Return(nil).Once()
//			err := uc.SendVerficationOTP(ctx, "user@example.com")
//			require.NoError(t, err)
//		})
//		t.Run("Already Verified", func(t *testing.T) {
//			ctx, userRepo, _, _, _, _, _, _, _, uc := setupAuthTest(t)
//			now := time.Now()
//			user := &models.User{
//				ID:              10,
//				Email:           "user@example.com",
//				EmailVerifiedAt: &now,
//			}
//			userRepo.EXPECT().FindByEmail(ctx, "user@example.com").Return(user, nil).Once()
//			err := uc.SendVerficationOTP(ctx, "user@example.com")
//			require.ErrorIs(t, err, domainerrors.ErrEmailAlreadyVerified)
//		})
//	}
//
//	func TestVerifyEmail(t *testing.T) {
//		t.Run("Success", func(t *testing.T) {
//			ctx, userRepo, otpStore, _, _, _, _, _, _, uc := setupAuthTest(t)
//			user := &models.User{
//				ID:    10,
//				Email: "test@example.com",
//			}
//			userRepo.EXPECT().FindByID(ctx, uint(10)).Return(user, nil).Once()
//			otpStore.EXPECT().Get(ctx, "email_verification_otp:test@example.com").Return("123456", nil).Once()
//			userRepo.EXPECT().UpdateEmailVerifiedAt(ctx, uint(10), mock.Anything).Return(nil).Once()
//			otpStore.EXPECT().Delete(ctx, "email_verification_otp:test@example.com").Return(nil).Once()
//			err := uc.VerifyEmail(ctx, 10, "123456")
//			require.NoError(t, err)
//		})
//		t.Run("Already Verified", func(t *testing.T) {
//			ctx, userRepo, _, _, _, _, _, _, _, uc := setupAuthTest(t)
//			now := time.Now()
//			user := &models.User{
//				ID:              10,
//				Email:           "test@example.com",
//				EmailVerifiedAt: &now,
//			}
//			userRepo.EXPECT().FindByID(ctx, uint(10)).Return(user, nil).Once()
//			err := uc.VerifyEmail(ctx, 10, "123456")
//			require.ErrorIs(t, err, domainerrors.ErrEmailAlreadyVerified)
//		})
//		t.Run("Invalid OTP", func(t *testing.T) {
//			ctx, userRepo, otpStore, _, _, _, _, _, _, uc := setupAuthTest(t)
//			user := &models.User{
//				ID:    10,
//				Email: "test@example.com",
//			}
//			userRepo.EXPECT().FindByID(ctx, uint(10)).Return(user, nil).Once()
//			otpStore.EXPECT().Get(ctx, "email_verification_otp:test@example.com").Return("654321", nil).Once()
//			err := uc.VerifyEmail(ctx, 10, "123456")
//			require.Error(t, err)
//			require.Contains(t, err.Error(), "Invalid OTP")
//		})
//	}
//
//	func TestForgotPasswordAndVerify(t *testing.T) {
//		t.Run("ForgotPassword Success", func(t *testing.T) {
//			ctx, userRepo, otpStore, _, _, _, emailSender, _, _, uc := setupAuthTest(t)
//			user := &models.User{
//				ID:    10,
//				Email: "user@example.com",
//			}
//			userRepo.EXPECT().FindByEmail(ctx, "user@example.com").Return(user, nil).Once()
//			otpStore.EXPECT().Set(ctx, "password_reset_otp:user@example.com", mock.Anything, 5*time.Minute).Return(nil).Once()
//			emailSender.EXPECT().SendVerificationOTP(ctx, "user@example.com", mock.Anything).Return(nil).Once()
//			err := uc.ForgotPassword(ctx, "user@example.com")
//			require.NoError(t, err)
//		})
//		t.Run("ForgotPassword User Not Found Silent Return", func(t *testing.T) {
//			ctx, userRepo, _, _, _, _, _, _, _, uc := setupAuthTest(t)
//			userRepo.EXPECT().FindByEmail(ctx, "unknown@example.com").Return(nil, domainerrors.ErrUserNotFound).Once()
//		err := uc.ForgotPassword(ctx, "unknown@example.com")
//			require.NoError(t, err) // Silently returns to prevent user enumeration
//		})
//		t.Run("VerifyForgotPasswordOTP Success", func(t *testing.T) {
//			ctx, userRepo, otpStore, _, _, _, _, _, _, uc := setupAuthTest(t)
//			user := &models.User{
//				ID:    10,
//				Email: "user@example.com",
//			}
//			userRepo.EXPECT().FindByEmail(ctx, "user@example.com").Return(user, nil).Once()
//			otpStore.EXPECT().Get(ctx, "password_reset_otp:user@example.com").Return("888999", nil).Once()
//			otpStore.EXPECT().Set(ctx, mock.MatchedBy(func(k string) bool {
//				return len(k) > 20
//			}), "10", 10*time.Minute).Return(nil).Once()
//			otpStore.EXPECT().Delete(ctx, "password_reset_otp:user@example.com").Return(nil).Once()
//			resetToken, err := uc.VerifyForgotPasswordOTP(ctx, "user@example.com", "888999")
//			require.NoError(t, err)
//			require.NotEmpty(t, resetToken)
//		})
//		t.Run("VerifyForgotPasswordOTP Invalid OTP", func(t *testing.T) {
//			ctx, userRepo, otpStore, _, _, _, _, _, _, uc := setupAuthTest(t)
//			user := &models.User{
//				ID:    10,
//				Email: "user@example.com",
//			}
//			userRepo.EXPECT().FindByEmail(ctx, "user@example.com").Return(user, nil).Once()
//			otpStore.EXPECT().Get(ctx, "password_reset_otp:user@example.com").Return("888999", nil).Once()
//			_, err := uc.VerifyForgotPasswordOTP(ctx, "user@example.com", "000111")
//			require.ErrorIs(t, err, domainerrors.ErrInvalidOTP)
//		})
//	}
//
//	func TestResetPassword(t *testing.T) {
//		newPass := "NewSecret@123"
//		t.Run("Success", func(t *testing.T) {
//			ctx, userRepo, otpStore, _, _, _, _, _, _, uc := setupAuthTest(t)
//			user := &models.User{
//				ID:    10,
//				Email: "user@example.com",
//			}
//			otpStore.EXPECT().Get(ctx, "password_reset_token:reset_token_xyz").Return("10", nil).Once()
//			userRepo.EXPECT().FindByID(ctx, uint(10)).Return(user, nil).Once()
//			userRepo.EXPECT().UpdatePassword(ctx, uint(10), mock.MatchedBy(func(hash string) bool {
//				return len(hash) > 20
//			})).Return(nil).Once()
//			otpStore.EXPECT().Delete(ctx, "password_reset_token:reset_token_xyz").Return(nil).Once()
//			err := uc.ResetPassword(ctx, "reset_token_xyz", newPass)
//			require.NoError(t, err)
//		})
//		t.Run("Weak Password", func(t *testing.T) {
//			ctx, _, _, _, _, _, _, _, _, uc := setupAuthTest(t)
//			err := uc.ResetPassword(ctx, "token_123", "weak")
//			require.ErrorIs(t, err, domainerrors.ErrWeakPassword)
//		})
//		t.Run("Empty Reset Token", func(t *testing.T) {
//			ctx, _, _, _, _, _, _, _, _, uc := setupAuthTest(t)
//			err := uc.ResetPassword(ctx, "", newPass)
//			require.ErrorIs(t, err, domainerrors.ErrInvalidCredentials)
//		})
//	}
//
//	func TestGoogleLogin(t *testing.T) {
//		t.Run("Existing Google User Success", func(t *testing.T) {
//			ctx, userRepo, _, jwtService, refreshStore, _, _, googleVerifier, _, uc := setupAuthTest(t)
//			googleVerifier.EXPECT().Verify(ctx, "valid_google_id_token").Return(&interfaces.GoogleUserInfo{
//				SubjectID: "google_sub_123",
//				Email:     "googleuser@example.com",
//				Name:      "Google User",
//			}, nil).Once()
//			user := &models.User{
//				ID:        15,
//				Email:     "googleuser@example.com",
//				Role:      models.RoleUser,
//				IsBlocked: false,
//			}
//			userRepo.EXPECT().FindByGoogleID(ctx, "google_sub_123").Return(user, nil).Once()
//			jwtService.EXPECT().GenerateAccessToken(uint(15), "user").Return("google_access_token", nil).Once()
//			jwtService.EXPECT().GenerateRefreshToken(uint(15)).Return("google_refresh_token", "google_jti", nil).Once()
//			refreshStore.EXPECT().Save(ctx, "google_jti", uint(15)).Return(nil).Once()
//			access, refresh, err := uc.GoogleLogin(ctx, "valid_google_id_token", "")
//			require.NoError(t, err)
//			require.Equal(t, "google_access_token", access)
//			require.Equal(t, "google_refresh_token", refresh)
//		})
//		t.Run("Invalid Google Token", func(t *testing.T) {
//			ctx, _, _, _, _, _, _, googleVerifier, _, uc := setupAuthTest(t)
//			googleVerifier.EXPECT().Verify(ctx, "bad_token").Return(nil, errors.New("invalid token")).Once()
//			_, _, err := uc.GoogleLogin(ctx, "bad_token", "")
//			require.ErrorIs(t, err, domainerrors.ErrInvalidCredentials)
//		})
//	}
type fixture struct {
	uc         *AuthUseCase
	userRepo   *mocks.MockUserRepository
	uow        *mocks.MockUnitOfWork
	jwt        *mocks.MockJWTService
	sessions   *mocks.MockRefreshSessionStore
	txUser     *mocks.MockUserRepository
	txReferral *mocks.MockReferralRepository
	txSettings *mocks.MockReferralSettingsRepository
}

func newFixture(t *testing.T) *fixture {
	f := &fixture{
		userRepo:   mocks.NewMockUserRepository(t),
		uow:        mocks.NewMockUnitOfWork(t),
		jwt:        mocks.NewMockJWTService(t),
		sessions:   mocks.NewMockRefreshSessionStore(t),
		txUser:     mocks.NewMockUserRepository(t),
		txReferral: mocks.NewMockReferralRepository(t),
		txSettings: mocks.NewMockReferralSettingsRepository(t),
	}
	f.uc = &AuthUseCase{
		userRepo:             f.userRepo, // same mock as in fixture
		uow:                  f.uow,      // same mock as in fixture
		log:                  slog.New(slog.NewTextHandler(io.Discard, nil)),
		otpStore:             mocks.NewMockOTPstore(t),
		jwtService:           f.jwt,
		refreshSessionStore:  f.sessions,
		accessTokenBlacklist: mocks.NewMockAccessTokenBlacklist(t),
		emailSender:          mocks.NewMockEmailSender(t),
		googleTokenVerifier:  mocks.NewMockGoogleTokenVerifier(t),
	}
	return f
}
func (f *fixture) expectTX() {
	f.uow.EXPECT().WithinTransaction(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, fn func(repos interfaces.TransactionRepositories) error) error {
		return fn(interfaces.TransactionRepositories{
			User:             f.txUser,
			Referral:         f.txReferral,
			ReferralSettings: f.txSettings,
		})
	})
}
func validUser() *models.User {
	pw := "Str0ng@Pass1"
	return &models.User{Name: "Niyas", Email: "niyas@test.com", Phone: "9876543210", Password: &pw}
}
func TestRegister_Validation(t *testing.T) {
	tests := []struct {
		name      string
		mutate    func(u *models.User)
		wantError error
	}{
		{"invalid name", func(u *models.User) { u.Name = " 1" }, domainerrors.ErrInvalidName},
		{"invalid phone", func(u *models.User) { u.Phone = " 123" }, domainerrors.ErrInvalidPhone},
		{"nil password", func(u *models.User) { u.Password = nil }, domainerrors.ErrWeakPassword},
		{"invalid password", func(u *models.User) { *u.Password = "hyyy" }, domainerrors.ErrWeakPassword},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture(t)
			user := validUser()
			tt.mutate(user)
			got, err := f.uc.Register(context.Background(), user, "")
			assert.ErrorIs(t, err, tt.wantError)
			assert.Nil(t, got)
		})
	}
}
func TestRegister_EmailAlreadyExists(t *testing.T) {
	f := newFixture(t)
	f.userRepo.EXPECT().FindByEmail(mock.Anything, "niyas@test.com").Return(&models.User{ID: 1}, nil)
	got, err := f.uc.Register(context.Background(), validUser(), "")
	assert.ErrorIs(t, err, domainerrors.ErrEmailAlreadyExists)
	assert.Nil(t, got)
}
func TestRegister_FindByEmailUnexpectedError(t *testing.T) {
	f := newFixture(t)
	dbErr := errors.New("db down")
	f.userRepo.EXPECT().FindByEmail(mock.Anything, mock.Anything).Return(nil, dbErr)
	_, err := f.uc.Register(context.Background(), validUser(), "")
	assert.ErrorIs(t, err, dbErr)
}
func TestRegister_Success_WithoutReferral(t *testing.T) {
	f := newFixture(t)
	f.userRepo.EXPECT().FindByEmail(mock.Anything, mock.Anything).Return(nil, domainerrors.ErrUserNotFound)
	f.expectTX()
	f.txUser.EXPECT().Create(mock.Anything, mock.AnythingOfType("*models.User")).Return(nil)
	user := validUser()
	plainpass := *user.Password
	got, err := f.uc.Register(context.Background(), user, "")
	assert.NoError(t, err)
	assert.Equal(t, models.RoleUser, got.Role)
	assert.NotNil(t, got.ReferralCode)
	assert.NotEqual(t, plainpass, *got.Password)
	assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(*got.Password), []byte(plainpass)))
}
func TestRegister_Success_WithReferral(t *testing.T) {
	f := newFixture(t)
	f.userRepo.EXPECT().FindByEmail(mock.Anything, mock.Anything).Return(nil, domainerrors.ErrUserNotFound)
	f.expectTX()
	f.txUser.EXPECT().Create(mock.Anything, mock.AnythingOfType("*models.User")).Run(func(_ context.Context, u *models.User) { u.ID = 2 }).Return(nil)
	f.txUser.EXPECT().FindByReferralCode(mock.Anything, "REF123").Return(&models.User{ID: 1}, nil)
	f.txSettings.EXPECT().Get(mock.Anything).Return(&models.ReferralSettings{
		ReferrerReward:     50,
		ReferredReward:     50,
		MinimumOrderAmount: 100,
		IsActive:           true,
	}, nil)
	f.txReferral.EXPECT().CreateIfNotReferred(mock.Anything, mock.MatchedBy(func(r *models.Referral) bool {
		return r.ReferrerID == 1 && r.ReferredUserID == 2 &&
			r.Status == models.ReferralStatusPending && r.ReferrerReward == 50
	})).Return(true, nil)
	got, err := f.uc.Register(context.Background(), validUser(), "REF123")
	require.NoError(t, err)
	assert.Equal(t, uint(2), got.ID)
}
func TestRegister_InvalidReferralCode_StillRegisters(t *testing.T) {
	f := newFixture(t)
	f.userRepo.EXPECT().FindByEmail(mock.Anything, mock.Anything).Return(nil, domainerrors.ErrUserNotFound)
	f.expectTX()
	f.txUser.EXPECT().Create(mock.Anything, mock.AnythingOfType("*models.User")).Return(nil)
	f.txUser.EXPECT().FindByReferralCode(mock.Anything, "BAD").Return(nil, domainerrors.ErrUserNotFound)
	got, err := f.uc.Register(context.Background(), validUser(), "BAD")
	require.NoError(t, err)
	assert.NotNil(t, got)
}
func TestRegister_TransactionFails(t *testing.T) {
	f := newFixture(t)
	f.userRepo.EXPECT().FindByEmail(mock.Anything, mock.Anything).Return(nil, domainerrors.ErrUserNotFound)
	f.expectTX()
	createErr := errors.New("insert failed")
	f.txUser.EXPECT().Create(mock.Anything, mock.Anything).Return(createErr)
	got, err := f.uc.Register(context.Background(), validUser(), "")
	assert.ErrorIs(t, err, createErr)
	assert.Nil(t, got)
}
func hashedUser(t *testing.T, plain string) *models.User {
	t.Helper()
	h, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.MinCost)
	require.NoError(t, err)
	hs := string(h)
	return &models.User{ID: 1, Email: "niyas@test.com", Password: &hs, Role: models.RoleUser}
}
func TestLogin(t *testing.T) {
	t.Run("Empty input", func(t *testing.T) {
		tests := []struct {
			name     string
			email    string
			password string
		}{
			{"empty email", "", "pass"},
			{"empty password", "a@b.com", ""},
			{"both empty", "", ""},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				f := newFixture(t)
				_, _, err := f.uc.Login(context.Background(), tt.email, tt.password)
				assert.ErrorIs(t, err, domainerrors.ErrInvalidCredentials)
			})
		}
	})
	t.Run("User Not Found", func(t *testing.T) {
		f := newFixture(t)
		f.userRepo.EXPECT().FindByEmail(mock.Anything, "x@test.com").Return(nil, domainerrors.ErrUserNotFound)
		_, _, err := f.uc.login(context.Background(), "x@test.com", "pass", false)
		assert.ErrorIs(t, err, domainerrors.ErrInvalidCredentials)
	})
	t.Run("Repo Error", func(t *testing.T) {
		f := newFixture(t)
		dbErr := errors.New("db down")
		f.userRepo.EXPECT().FindByEmail(mock.Anything, mock.Anything).Return(nil, dbErr)
		_, _, err := f.uc.login(context.Background(), "a@b.com", "pass", false)
		assert.ErrorIs(t, err, dbErr)
	})
	t.Run("nil password", func(t *testing.T) {
		f := newFixture(t)
		f.userRepo.EXPECT().FindByEmail(mock.Anything, mock.Anything).
			Return(&models.User{ID: 1, Password: nil}, nil)

		_, _, err := f.uc.login(context.Background(), "a@b.com", "pass", false)

		assert.ErrorIs(t, err, domainerrors.ErrInvalidCredentials)
	})
	t.Run("WrongPassword", func(t *testing.T) {
		f := newFixture(t)
		f.userRepo.EXPECT().FindByEmail(mock.Anything, mock.Anything).Return(hashedUser(t, "Correct@123"), nil)
		_, _, err := f.uc.login(context.Background(), "a@b.com", "Wrong@123", false)
		assert.ErrorIs(t, err, domainerrors.ErrInvalidCredentials)
	})
	t.Run("Blocked User", func(t *testing.T) {
		f := newFixture(t)
		u := hashedUser(t, "Correct@123")
		u.IsBlocked = true
		f.userRepo.EXPECT().FindByEmail(mock.Anything, mock.Anything).Return(u, nil)
		_, _, err := f.uc.login(context.Background(), "a@b.com", "Correct@123", false)
		assert.ErrorIs(t, err, domainerrors.ErrUserBlocked)
	})
	t.Run("Admin Only", func(t *testing.T) {
		f := newFixture(t)
		f.userRepo.EXPECT().FindByEmail(mock.Anything, mock.Anything).Return(hashedUser(t, "Correct@123"), nil)
		_, _, err := f.uc.login(context.Background(), "a@b.com", "Correct@123", true)
		assert.ErrorIs(t, err, domainerrors.ErrInvalidCredentials)
	})
	t.Run("Access Token Error", func(t *testing.T) {
		f := newFixture(t)
		tokErr := errors.New("sign failed")
		f.userRepo.EXPECT().FindByEmail(mock.Anything, mock.Anything).Return(hashedUser(t, "Correct@123"), nil)
		f.jwt.EXPECT().GenerateAccessToken(uint(1), "user").Return("", tokErr)
		_, _, err := f.uc.login(context.Background(), "a@b.com", "Correct@123", false)
		assert.ErrorIs(t, err, tokErr)
	})
	t.Run("Refresh Token Error", func(t *testing.T) {
		f := newFixture(t)
		tokErr := errors.New("sign failed")
		f.userRepo.EXPECT().FindByEmail(mock.Anything, mock.Anything).Return(hashedUser(t, "Correct@123"), nil)
		f.jwt.EXPECT().GenerateAccessToken(uint(1), "user").Return("access", nil)
		f.jwt.EXPECT().GenerateRefreshToken(uint(1)).Return("", "", tokErr)
		_, _, err := f.uc.login(context.Background(), "a@b.com", "Correct@123", false)
		assert.ErrorIs(t, err, tokErr)
	})
	t.Run("Session Save Error", func(t *testing.T) {
		f := newFixture(t)
		redisErr := errors.New("redis down")
		f.userRepo.EXPECT().FindByEmail(mock.Anything, mock.Anything).Return(hashedUser(t, "Correct@123"), nil)
		f.jwt.EXPECT().GenerateAccessToken(uint(1), "user").Return("access", nil)
		f.jwt.EXPECT().GenerateRefreshToken(uint(1)).Return("refresh", "jti-1", nil)
		f.sessions.EXPECT().Save(mock.Anything, "jti-1", uint(1)).Return(redisErr)
		at, rt, err := f.uc.login(context.Background(), "a@b.com", "Correct@123", false)
		assert.ErrorIs(t, err, redisErr)
		assert.Empty(t, at)
		assert.Empty(t, rt)
	})
	t.Run("Success", func(t *testing.T) {
		f := newFixture(t)
		f.userRepo.EXPECT().FindByEmail(mock.Anything, "a@b.com").Return(hashedUser(t, "Correct@123"), nil)
		f.jwt.EXPECT().GenerateAccessToken(uint(1), "user").Return("access-token", nil)
		f.jwt.EXPECT().GenerateRefreshToken(uint(1)).Return("refresh-token", "jti-1", nil)
		f.sessions.EXPECT().Save(mock.Anything, "jti-1", uint(1)).Return(nil)
		at, rt, err := f.uc.login(context.Background(), "a@b.com", "Correct@123", false)
		require.NoError(t, err)
		assert.Equal(t, "access-token", at)
		assert.Equal(t, "refresh-token", rt)
	})
	t.Run("Admin Success", func(t *testing.T) {
		f := newFixture(t)
		u := hashedUser(t, "Correct@123")
		u.Role = models.RoleAdmin
		f.userRepo.EXPECT().FindByEmail(mock.Anything, "a@b.com").Return(u, nil)
		f.jwt.EXPECT().GenerateAccessToken(uint(1), "admin").Return("access-token", nil)
		f.jwt.EXPECT().GenerateRefreshToken(uint(1)).Return("refresh-token", "jti-1", nil)
		f.sessions.EXPECT().Save(mock.Anything, "jti-1", uint(1)).Return(nil)
		at, rt, err := f.uc.login(context.Background(), "a@b.com", "Correct@123", true)
		require.NoError(t, err)
		assert.Equal(t, "access-token", at)
		assert.Equal(t, "refresh-token", rt)
	})
}
