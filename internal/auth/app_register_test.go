package auth

import (
	"testing"
	"time"
)

func TestVerifyAppRegister(t *testing.T) {
	const secret = "test-secret-for-register"
	t.Setenv("APP_REGISTER_SECRET", secret)

	email := "user@example.com"
	ts := time.Now().Unix()
	sign := SignAppRegister(secret, AppClientIOS, email, ts)

	ok, msg := VerifyAppRegister(AppClientIOS, email, sign, ts)
	if !ok {
		t.Fatalf("expected ok, got msg=%q", msg)
	}

	ok, _ = VerifyAppRegister(AppClientAndroid, email, SignAppRegister(secret, AppClientAndroid, email, ts), ts)
	if !ok {
		t.Fatal("android client should be allowed")
	}

	ok, msg = VerifyAppRegister(AppClientIOS, email, "deadbeef", ts)
	if ok || msg == "" {
		t.Fatalf("bad sign should fail, ok=%v msg=%q", ok, msg)
	}

	ok, _ = VerifyAppRegister(AppClientIOS, email, sign, ts-int64(appRegisterSkew/time.Second)-10)
	if ok {
		t.Fatal("expired ts should fail")
	}

	t.Setenv("APP_REGISTER_SECRET", "")
	ok, msg = VerifyAppRegister(AppClientIOS, email, sign, ts)
	if ok || msg == "" {
		t.Fatalf("empty secret should disable channel, ok=%v msg=%q", ok, msg)
	}
}
