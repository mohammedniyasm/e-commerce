package redis

import (
	"context"
	"testing"
	"time"

	redisClient "github.com/redis/go-redis/v9"
)

func TestRefreshSessionStore_SaveAndGet(t *testing.T) {
	client := redisClient.NewClient(&redisClient.Options{
		Addr: "localhost:6379",
	})
	ctx := context.Background()
	store := NewRefreshSessionStore(client, 1*time.Minute)

	jti := "test-jti-123"

	userID := uint(25)

	err := store.Save(ctx, jti, userID)
	if err != nil {
		t.Fatalf("failed to save refresh session: %v", err)
	}

	// value, err := client.Get(ctx, "refresh_session:"+jti).Result()
	// if err != nil {
	// 	t.Fatalf("failed to read directly from redis: %v", err)
	// }
	
	// t.Logf("stored value: %s", value)
	gotUserID, err := store.Get(ctx, jti)
	if err != nil {
		t.Fatalf("failed to get refresh session: %v", err)
	}
	if gotUserID != userID {
		t.Fatalf("expected user ID %d, got %d", userID, gotUserID)
	}

	// Cleanup
	_ = store.Delete(ctx, jti)
}
func TestRefreshSessionStore_Delete(t *testing.T) {
	client := redisClient.NewClient(&redisClient.Options{
		Addr: "localhost:6379",
	})

	ctx := context.Background()

	store := NewRefreshSessionStore(
		client,
		1*time.Minute,
	)

	jti := "test-delete-jti"
	userID := uint(25)

	err := store.Save(ctx, jti, userID)
	if err != nil {
		t.Fatalf("failed to save refresh session: %v", err)
	}

	err = store.Delete(ctx, jti)
	if err != nil {
		t.Fatalf("failed to delete refresh session: %v", err)
	}

	_, err = store.Get(ctx, jti)
	if err == nil {
		t.Fatal("expected refresh session to be deleted")
	}
}