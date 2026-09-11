package auth

import (
	"os"
	"testing"
	"time"

	"github.com/KyleBing/portal-go/internal/models"
	"github.com/golang-jwt/jwt/v5"
)

func TestIssueAndParse(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-for-unit")
	if err := Init(); err != nil {
		t.Fatal(err)
	}
	token, err := Issue(&models.User{UID: 42, GroupID: 1})
	if err != nil {
		t.Fatal(err)
	}
	claims, err := Parse(token)
	if err != nil {
		t.Fatal(err)
	}
	if claims.UID != 42 || claims.GroupID != 1 {
		t.Fatalf("unexpected claims: %+v", claims)
	}
	if ShouldRenew(claims) {
		t.Fatal("fresh token should not renew")
	}
}

func TestShouldRenew(t *testing.T) {
	now := time.Now()
	near := &Claims{RegisteredClaims: jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(now.Add(3 * 24 * time.Hour)),
	}}
	if !ShouldRenew(near) {
		t.Fatal("token within renew window should renew")
	}
	far := &Claims{RegisteredClaims: jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(now.Add(20 * 24 * time.Hour)),
	}}
	if ShouldRenew(far) {
		t.Fatal("token outside renew window should not renew")
	}
}

func TestBearerToken(t *testing.T) {
	if got := BearerToken("Bearer abc.def.ghi"); got != "abc.def.ghi" {
		t.Fatalf("got %q", got)
	}
	if got := BearerToken("bearer xyz"); got != "xyz" {
		t.Fatalf("got %q", got)
	}
	if got := BearerToken(""); got != "" {
		t.Fatalf("got %q", got)
	}
}
