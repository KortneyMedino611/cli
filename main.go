package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/zalando/go-keyring"
)

const (
	keyringService = "gh"
	keyringAccount = "token"
)

var (
	keyringGet      = keyring.Get
	timeoutDuration = 3 * time.Second
)

func main() {
	token, err := getAuthToken()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if token == "" {
		fmt.Println("Proceeding as unauthenticated.")
	} else {
		fmt.Println("Authenticated successfully.")
	}
}

func getAuthToken() (string, error) {
	// 1. Check GH_TOKEN or GITHUB_TOKEN environment variables first
	token := os.Getenv("GH_TOKEN")
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
	}
	if token != "" {
		return token, nil
	}

	// 2. Attempt to retrieve from keyring with timeout
	token, err := getKeyringTokenWithTimeout(timeoutDuration)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			// Safe error: credentials not found. Fall back to unauthenticated state.
			return "", nil
		}
		// System error or timeout
		return "", fmt.Errorf("keyring lookup failed: %w", err)
	}

	return token, nil
}

func getKeyringTokenWithTimeout(timeout time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	type result struct {
		token string
		err   error
	}
	ch := make(chan result, 1)

	go func() {
		token, err := keyringGet(keyringService, keyringAccount)
		ch <- result{token: token, err: err}
	}()

	select {
	case res := <-ch:
		return res.token, res.err
	case <-ctx.Done():
		return "", fmt.Errorf("keyring lookup timed out after %v. Please check your system credential helper", timeout)
	}
}
