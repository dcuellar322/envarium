package application

import (
	"errors"
	"testing"
	"time"
)

func TestBootstrapIsOneTimeAndMutationRequiresCSRF(t *testing.T) {
	t.Parallel()

	manager := NewManager()
	bootstrap, err := manager.IssueBootstrap()
	if err != nil {
		t.Fatalf("IssueBootstrap() error = %v", err)
	}
	session, err := manager.Exchange(bootstrap.Token)
	if err != nil {
		t.Fatalf("Exchange() error = %v", err)
	}
	if _, err := manager.Exchange(bootstrap.Token); !errors.Is(err, ErrInvalidBootstrap) {
		t.Fatalf("second Exchange() error = %v", err)
	}
	if _, err := manager.ValidateSession("wrong"); !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("ValidateMutation(wrong) error = %v", err)
	}
	if _, err := manager.ValidateSession(session.CSRFToken); err != nil {
		t.Fatalf("ValidateMutation() error = %v", err)
	}
}

func TestExpiredCredentialsAreRejected(t *testing.T) {
	t.Parallel()

	manager := NewManager()
	now := time.Date(2026, 7, 16, 0, 0, 0, 0, time.UTC)
	manager.now = func() time.Time { return now }
	bootstrap, err := manager.IssueBootstrap()
	if err != nil {
		t.Fatalf("IssueBootstrap() error = %v", err)
	}
	now = now.Add(bootstrapLifetime + time.Second)
	if _, err := manager.Exchange(bootstrap.Token); !errors.Is(err, ErrInvalidBootstrap) {
		t.Fatalf("Exchange(expired) error = %v", err)
	}
}

func TestSessionsExpireWhenIdleButActiveSessionsRetainTheirAbsoluteLimit(t *testing.T) {
	t.Parallel()

	manager := NewManager()
	now := time.Date(2026, 7, 16, 0, 0, 0, 0, time.UTC)
	manager.now = func() time.Time { return now }
	bootstrap, err := manager.IssueBootstrap()
	if err != nil {
		t.Fatal(err)
	}
	session, err := manager.Exchange(bootstrap.Token)
	if err != nil {
		t.Fatal(err)
	}
	now = now.Add(sessionIdleLifetime - time.Second)
	if _, err := manager.ValidateSession(session.CSRFToken); err != nil {
		t.Fatalf("active session rejected: %v", err)
	}
	now = now.Add(sessionIdleLifetime - time.Second)
	if _, err := manager.ValidateSession(session.CSRFToken); err != nil {
		t.Fatalf("refreshed session rejected: %v", err)
	}
	now = session.ExpiresAt.Add(time.Second)
	if _, err := manager.ValidateSession(session.CSRFToken); !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("session beyond absolute limit error = %v", err)
	}

	bootstrap, err = manager.IssueBootstrap()
	if err != nil {
		t.Fatal(err)
	}
	idle, err := manager.Exchange(bootstrap.Token)
	if err != nil {
		t.Fatal(err)
	}
	now = now.Add(sessionIdleLifetime + time.Second)
	if _, err := manager.ValidateSession(idle.CSRFToken); !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("idle session error = %v", err)
	}
}

func TestInvalidProofDoesNotKeepSessionAliveAndTabsAreIndependent(t *testing.T) {
	manager := NewManager()
	now := time.Now()
	manager.now = func() time.Time { return now }
	boot, _ := manager.IssueBootstrap()
	first, _ := manager.Exchange(boot.Token)
	boot, _ = manager.IssueBootstrap()
	second, _ := manager.Exchange(boot.Token)
	now = now.Add(sessionIdleLifetime - time.Second)
	if _, err := manager.ValidateSession(first.ID); !errors.Is(err, ErrInvalidSession) {
		t.Fatal("owner ID accepted as credential")
	}
	if _, err := manager.ValidateSession(second.CSRFToken); err != nil {
		t.Fatal(err)
	}
	now = now.Add(2 * time.Second)
	if _, err := manager.ValidateSession(first.CSRFToken); !errors.Is(err, ErrInvalidSession) {
		t.Fatal("invalid proof extended idle life")
	}
	if _, err := manager.ValidateSession(second.CSRFToken); err != nil {
		t.Fatal("another tab invalidated active session")
	}
}
