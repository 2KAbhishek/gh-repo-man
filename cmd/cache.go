package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const DefaultCacheDir = "~/.cache/gh-repo-man"

func GetCacheDir() (string, error) {
	cacheDir, err := expandPath(DefaultCacheDir)
	if err != nil {
		return "", fmt.Errorf("failed to expand cache directory path: %w", err)
	}

	if err := os.MkdirAll(cacheDir, 0o750); err != nil {
		return "", fmt.Errorf("failed to create cache directory: %w", err)
	}

	readmeDir := filepath.Join(cacheDir, "readmes")
	if err := os.MkdirAll(readmeDir, 0o750); err != nil {
		return "", fmt.Errorf("failed to create readme cache directory: %w", err)
	}

	return cacheDir, nil
}

func ParseTTL(duration string) (time.Duration, error) {
	if duration == "" {
		return 24 * time.Hour, nil
	}

	duration = strings.TrimSpace(duration)
	if len(duration) < 2 {
		return 0, fmt.Errorf("invalid duration format: %s", duration)
	}

	unitStr := duration[len(duration)-1:]
	valueStr := duration[:len(duration)-1]

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return 0, fmt.Errorf("invalid duration value: %s", valueStr)
	}

	switch unitStr {
	case "m":
		return time.Duration(value) * time.Minute, nil
	case "h":
		return time.Duration(value) * time.Hour, nil
	case "d":
		return time.Duration(value) * 24 * time.Hour, nil
	default:
		return 0, fmt.Errorf("invalid duration unit: %s (supported: m, h, d)", unitStr)
	}
}

func IsCacheValid(filePath string, ttl time.Duration) bool {
	info, err := os.Stat(filePath)
	if err != nil {
		return false
	}

	return time.Since(info.ModTime()) < ttl
}

func LoadReposFromCache(user string) ([]Repo, error) {
	cacheDir, err := GetCacheDir()
	if err != nil {
		return nil, err
	}

	actualUser := user
	if user == "" {
		cachedUser, err := GetCachedCurrentUsername()
		if err != nil {
			return nil, fmt.Errorf("failed to get current username: %w", err)
		}
		actualUser = cachedUser
	}

	filename := fmt.Sprintf("%s_repos.json", actualUser)
	filePath := filepath.Join(cacheDir, filename)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var repos []Repo
	if err := json.Unmarshal(data, &repos); err != nil {
		return nil, fmt.Errorf("failed to parse cached repos: %w", err)
	}

	return repos, nil
}

func atomicWriteFile(filePath string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(filePath)
	tmpFile, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmpFile.Name()

	defer func() {
		_ = tmpFile.Close()
		_ = os.Remove(tmpName)
	}()

	if err := tmpFile.Chmod(perm); err != nil {
		return err
	}
	if _, err := tmpFile.Write(data); err != nil {
		return err
	}
	if err := tmpFile.Sync(); err != nil {
		return err
	}
	if err := tmpFile.Close(); err != nil {
		return err
	}

	return os.Rename(tmpName, filePath)
}

func SaveReposToCache(user string, repos []Repo) error {
	cacheDir, err := GetCacheDir()
	if err != nil {
		return err
	}

	actualUser := user
	if user == "" {
		currentUser, err := GetCachedCurrentUsername()
		if err != nil {
			return fmt.Errorf("failed to get current username: %w", err)
		}
		actualUser = currentUser
	}

	filename := fmt.Sprintf("%s_repos.json", actualUser)
	filePath := filepath.Join(cacheDir, filename)
	data, err := json.MarshalIndent(repos, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal repos: %w", err)
	}

	if err := atomicWriteFile(filePath, data, 0o600); err != nil {
		return fmt.Errorf("failed to write repos cache: %w", err)
	}

	return nil
}

// GetCachedCurrentUsername gets the current username with caching (fast fallback + background rehydrate)
func GetCachedCurrentUsername() (string, error) {
	cacheDir, err := GetCacheDir()
	if err != nil {
		return "", err
	}

	usernameCachePath := filepath.Join(cacheDir, "current_username.txt")
	usernameCacheTTL, err := ParseTTL(config.Performance.Cache.Username)
	if err != nil {
		usernameCacheTTL = 90 * 24 * time.Hour
	}

	data, err := os.ReadFile(usernameCachePath)
	if err == nil && len(data) > 0 {
		cachedUsername := strings.TrimSpace(string(data))
		if cachedUsername != "" {
			if !IsCacheValid(usernameCachePath, usernameCacheTTL) {
				go func() {
					if freshUser, err := GetCurrentUsername(); err == nil && freshUser != "" {
						_ = atomicWriteFile(usernameCachePath, []byte(freshUser), 0o600)
					}
				}()
			}
			return cachedUsername, nil
		}
	}

	username, err := GetCurrentUsername()
	if err != nil {
		return "", err
	}

	if err := atomicWriteFile(usernameCachePath, []byte(username), 0o600); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to cache username: %v\n", err)
	}

	return username, nil
}

func LoadReadmeFromCache(user, repoName string) (string, error) {
	cacheDir, err := GetCacheDir()
	if err != nil {
		return "", err
	}

	filename := fmt.Sprintf("%s_%s.md", user, repoName)
	filePath := filepath.Join(cacheDir, "readmes", filename)

	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func SaveReadmeToCache(user, repoName, content string) error {
	cacheDir, err := GetCacheDir()
	if err != nil {
		return err
	}

	filename := fmt.Sprintf("%s_%s.md", user, repoName)
	filePath := filepath.Join(cacheDir, "readmes", filename)

	if err := atomicWriteFile(filePath, []byte(content), 0o600); err != nil {
		return fmt.Errorf("failed to write readme cache: %w", err)
	}

	return nil
}

func getReadmeCachePath(user, repoName string) (string, error) {
	cacheDir, err := GetCacheDir()
	if err != nil {
		return "", err
	}

	filename := fmt.Sprintf("%s_%s.md", user, repoName)
	return filepath.Join(cacheDir, "readmes", filename), nil
}
