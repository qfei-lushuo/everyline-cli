//go:build windows

package invocation

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Runs only inside the existing time-bounded helper. Environment evidence never
// overrides an ancestry match or a failed identity verification.
func platformEnvironmentFallback(result Result) Result {
	return windowsEnvironmentFallback(result, os.Getenv)
}

func windowsEnvironmentFallback(result Result, getenv func(string) string) Result {
	if result.AgentSourceType != "unknown" || result.EvidenceType != "none" {
		return result
	}
	workbuddy := workbuddyEnvironment(getenv("BASH_ENV"))
	doubao := doubaoEnvironment(getenv("CUA_BASE_PACK_DIR"), getenv("CUA_DLC_DIR"))
	if workbuddy && doubao != "" {
		result.Warnings = append(result.Warnings, "conflicting Windows runtime environment markers")
		return result
	}
	source, rule := doubao, ""
	if workbuddy {
		source = "workbuddy"
	}
	switch source {
	case "workbuddy":
		rule = "client.workbuddy.environment"
	case "doubao":
		rule = "client.doubao.environment"
	case "doubaoWork":
		rule = "client.doubao_work.environment"
	default:
		return result
	}
	result.AgentSourceType = source
	result.EvidenceType = "windows_runtime_environment"
	result.Confidence = "low"
	result.RuleID = rule
	result.Reason = "matched Windows runtime environment for " + source + "; process ancestry did not identify a client"
	return result
}

// Accept native Windows and Git Bash drive paths, including custom installation
// roots with spaces or non-ASCII characters. Never search a default install dir.
func runtimePath(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, "\\", "/")
	if len(value) >= 3 && value[0] == '/' && value[2] == '/' &&
		((value[1] >= 'a' && value[1] <= 'z') || (value[1] >= 'A' && value[1] <= 'Z')) {
		value = string(value[1]) + ":" + value[2:]
	}
	value = filepath.Clean(filepath.FromSlash(value))
	if !filepath.IsAbs(value) {
		return ""
	}
	return value
}

func workbuddyEnvironment(value string) bool {
	path := runtimePath(value)
	if path == "" || !strings.EqualFold(filepath.Base(path), "shell-runtime-bash-env.sh") {
		return false
	}
	shim := filepath.Dir(path)
	vendor := filepath.Dir(shim)
	if !strings.EqualFold(filepath.Base(shim), "shim") || !strings.EqualFold(filepath.Base(vendor), "vendor") {
		return false
	}
	info, err := os.Stat(path)
	if err != nil || !info.Mode().IsRegular() {
		return false
	}
	file, err := os.Open(filepath.Join(filepath.Dir(vendor), "product.json"))
	if err != nil {
		return false
	}
	defer file.Close()
	// The installed product file includes extensive UI config (~384 KiB).
	// Bound the read while allowing that real runtime metadata.
	data, err := io.ReadAll(io.LimitReader(file, 1024*1024+1))
	if err != nil || len(data) > 1024*1024 {
		return false
	}
	var product struct {
		ProductName    string `json:"productName"`
		Authentication struct {
			ID string `json:"id"`
		} `json:"authentication"`
	}
	if json.Unmarshal(data, &product) != nil {
		return false
	}
	return product.ProductName == "WorkBuddy" && product.Authentication.ID == "workbuddy-desktop"
}

func doubaoEnvironment(baseValue, dlcValue string) string {
	base, dlc := runtimePath(baseValue), runtimePath(dlcValue)
	if base == "" || dlc == "" {
		return ""
	}
	bases := filepath.Dir(base)
	runtimeDir := filepath.Dir(bases)
	root := filepath.Dir(runtimeDir)
	envs := filepath.Dir(dlc)
	envsDir := filepath.Dir(envs)
	profile := filepath.Dir(envsDir)
	if !strings.EqualFold(filepath.Base(bases), "bases") ||
		!strings.EqualFold(filepath.Base(runtimeDir), "sandbox_runtime") ||
		!strings.EqualFold(filepath.Base(root), "User Data") ||
		!strings.EqualFold(filepath.Base(envs), "envs") ||
		!strings.EqualFold(filepath.Base(envsDir), "sandbox_envs_dir") ||
		!strings.EqualFold(filepath.Clean(filepath.Dir(profile)), root) {
		return ""
	}
	for _, path := range []string{base, dlc} {
		info, err := os.Stat(path)
		if err != nil || !info.IsDir() {
			return ""
		}
	}
	switch strings.ToLower(filepath.Base(filepath.Dir(root))) {
	case "doubao":
		return "doubao"
	case "doubaowork":
		return "doubaoWork"
	}
	return ""
}
