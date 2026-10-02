package config

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	cases := []struct {
		name     string
		setup    func(t *testing.T, dir string) string
		wantErr  bool
		notExist bool
		wantPort int
	}{
		{
			name: "successful read",
			setup: func(t *testing.T, dir string) string {
				p := filepath.Join(dir, "ok.json")
				cfg := Config{ServerPort: 8080, Environment: "development", DebugMode: true}
				if err := SaveConfig(p, cfg); err != nil {
					t.Fatalf("setup save: %v", err)
				}
				return p
			},
			wantPort: 8080,
		},
		{
			name: "file not found",
			setup: func(t *testing.T, dir string) string {
				return filepath.Join(dir, "missing.json")
			},
			wantErr:  true,
			notExist: true,
		},
		{
			name: "corrupted JSON",
			setup: func(t *testing.T, dir string) string {
				p := filepath.Join(dir, "bad.json")
				if err := os.WriteFile(p, []byte("{not json"), 0o644); err != nil {
					t.Fatalf("setup write: %v", err)
				}
				return p
			},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			path := tc.setup(t, dir)
			got, err := LoadConfig(path)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error")
					return
				}
				if tc.notExist && !errors.Is(err, os.ErrNotExist) {
					t.Errorf("want os.ErrNotExist, got %v", err)
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if got.ServerPort != tc.wantPort {
				t.Errorf("ServerPort = %d, want %d", got.ServerPort, tc.wantPort)
			}
			if got.AdminPassword != "" {
				t.Errorf("AdminPassword should not be loaded from json, got %q", got.AdminPassword)
			}
		})
	}
}

func TestSaveConfig_OmitsSecretsAndEmptyURL(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cfg.json")
	cfg := Config{ServerPort: 3000, Environment: "staging", AdminPassword: "secret"}
	if err := SaveConfig(path, cfg); err != nil {
		t.Fatalf("save: %v", err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	s := string(raw)
	if strings.Contains(s, "secret") || strings.Contains(s, "AdminPassword") {
		t.Errorf("password leaked in json: %s", s)
	}
	if strings.Contains(s, "database_url") {
		t.Errorf("empty database_url should be omitted: %s", s)
	}
}

// AI generated most of these cases from the signature only.
// I already had 1024 and 65535 in mind. AI added 0, 80, 99999 and empty env.
// AI also claimed port 1023 should pass — that is wrong, 1024 is the min.
func TestValidateConfig(t *testing.T) {
	cases := []struct {
		name string
		cfg  Config
		ok   bool
	}{
		{"ok production 8080", Config{ServerPort: 8080, Environment: "production"}, true},
		{"min port 1024", Config{ServerPort: 1024, Environment: "development"}, true},
		{"max port 65535", Config{ServerPort: 65535, Environment: "staging"}, true},
		{"port 0", Config{ServerPort: 0, Environment: "development"}, false},
		{"port 80", Config{ServerPort: 80, Environment: "development"}, false},
		{"port 99999", Config{ServerPort: 99999, Environment: "development"}, false},
		{"empty env", Config{ServerPort: 8080, Environment: ""}, false},
		{"weird env", Config{ServerPort: 8080, Environment: "prod!"}, false},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateConfig(tc.cfg)
			if tc.ok && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tc.ok && err == nil {
				t.Errorf("expected error")
			}
		})
	}
}

type mockStorage struct {
	createErr error
	openErr   error
}

type errWriter struct{ err error }

func (e errWriter) Write([]byte) (int, error) { return 0, e.err }
func (e errWriter) Close() error              { return nil }

type nopReader struct{}

func (nopReader) Read([]byte) (int, error) { return 0, io.EOF }
func (nopReader) Close() error             { return nil }

func (m mockStorage) Create(path string) (io.WriteCloser, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	return errWriter{err: m.createErr}, nil
}

func (m mockStorage) Open(path string) (io.ReadCloser, error) {
	if m.openErr != nil {
		return nil, m.openErr
	}
	return nopReader{}, nil
}

func TestSaveConfig_WrapsDiskErrors(t *testing.T) {
	err := saveConfig(mockStorage{createErr: os.ErrPermission}, "cfg.json", Config{ServerPort: 8080, Environment: "development"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, os.ErrPermission) {
		t.Errorf("want permission wrap, got %v", err)
	}
}

func TestLoadConfig_WrapsDiskFull(t *testing.T) {
	diskFull := errors.Join(os.ErrInvalid, syscall.ENOSPC)
	err := saveConfig(mockStorage{createErr: diskFull}, "cfg.json", Config{ServerPort: 8080, Environment: "development"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "save config") {
		t.Errorf("error not wrapped: %v", err)
	}
}
