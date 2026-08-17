package main

import (
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/zalando/go-keyring"
)

func TestGetAuthToken_EnvVarFallback(t *testing.T) {
	os.Setenv("GH_TOKEN", "env-token")
	defer os.Unsetenv("GH_TOKEN")

	token, err := getAuthToken()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if token != "env-token" {
		t.Errorf("expected token 'env-token', got %q", token)
	}
}

func TestGetAuthToken_KeyringSuccess(t *testing.T) {
	os.Unsetenv("GH_TOKEN")
	os.Unsetenv("GITHUB_TOKEN")

	oldKeyringGet := keyringGet
	defer func() { keyringGet = oldKeyringGet }()

	keyringGet = func(service, user string) (string, error) {
		return "keyring-token", nil
	}

	token, err := getAuthToken()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if token != "keyring-token" {
		t.Errorf("expected token 'keyring-token', got %q", token)
	}
}

func TestGetAuthToken_KeyringNotFound(t *testing.T) {
	os.Unsetenv("GH_TOKEN")
	os.Unsetenv("GITHUB_TOKEN")

	oldKeyringGet := keyringGet
	defer func() { keyringGet = oldKeyringGet }()

	keyringGet = func(service, user string) (string, error) {
		return "", keyring.ErrNotFound
	}

	token, err := getAuthToken()
	if err != nil {
		t.Fatalf("expected no error for ErrNotFound, got %v", err)
	}
	if token != "" {
		t.Errorf("expected empty token, got %q", token)
	}
}

func TestGetAuthToken_KeyringSystemError(t *testing.T) {
	os.Unsetenv("GH_TOKEN")
	os.Unsetenv("GITHUB_TOKEN")

	oldKeyringGet := keyringGet
	defer func() { keyringGet = oldKeyringGet }()

	expectedErr := errors.New("dbus connection failed")
	keyringGet = func(service, user string) (string, error) {
		return "", expectedErr
	}

	_, err := getAuthToken()
	if err == nil {
		t.Fatal("expected error for system error, got nil")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected wrapped error %v, got %v", expectedErr, err)
	}
}

func TestGetAuthToken_KeyringTimeout(t *testing.T) {
	os.Unsetenv("GH_TOKEN")
	os.Unsetenv("GITHUB_TOKEN")

	oldKeyringGet := keyringGet
	defer func() { keyringGet = oldKeyringGet }()

	keyringGet = func(service, user string) (string, error) {
		time.Sleep(100 * time.Millisecond)
		return "slow-token", nil
	}

	oldTimeout := timeoutDuration
	timeoutDuration = 10 * time.Millisecond
	defer func() { timeoutDuration = oldTimeout }()

	_, err := getAuthToken()
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if !strings.Contains(err.Error(), "timed out") {
		t.Errorf("expected timeout error message, got %v", err)
	}
}
