package utils

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
	"github.com/zalando/go-keyring"
)

var (
	ErrMissingToken = errors.New("Access token not provided. Supply an access token by running " + Green("supabase login") + " or setting the SUPABASE_ACCESS_TOKEN environment variable.")
)

const namespace = "Supabase CLI"
const AccessTokenKey = "access-token"

func Get(project string) (string, error) {
	if err := assertKeyringSupported(); err != nil {
		return "", err
	}
	return keyring.Get(namespace, project)
}

func GetSupabaseToken(fs afero.Fs) (string, error) {
	// Env takes precedence
	if accessToken := os.Getenv("SUPABASE_ACCESS_TOKEN"); accessToken != "" {
		return accessToken, nil
	}
	// Load from native credentials store
	if accessToken, err := Get(AccessTokenKey); err == nil {
		return accessToken, nil
	}
	// Fallback to token file
	return fallbackLoadToken(fs)
}

func assertKeyringSupported() error {
	// Suggested check: https://github.com/microsoft/WSL/issues/423
	if f, err := os.ReadFile("/proc/sys/kernel/osrelease"); err == nil && bytes.Contains(f, []byte("WSL")) {
		return errors.New("Keyring is not supported on WSL")
	}

	return nil
}

func fallbackLoadToken(fsys afero.Fs) (string, error) {
	path, err := getAccessTokenPath()

	if err != nil {
		return "", err
	}

	accessToken, err := afero.ReadFile(fsys, path)

	if errors.Is(err, os.ErrNotExist) {
		return "", ErrMissingToken
	} else if err != nil {
		return "", err
	}

	return string(accessToken), nil
}

func getAccessTokenPath() (string, error) {
	home, err := os.UserHomeDir()

	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".supabase", AccessTokenKey), nil
}
