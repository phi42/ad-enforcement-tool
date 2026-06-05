package plugin

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// FetchRelease downloads the binary for the current OS and architecture from
// the latest GitHub release of moduleURL (e.g. "github.com/owner/repo") and
// writes it to dst.
func FetchRelease(moduleURL, dst string) error {
	owner, repo, err := ParseModuleURL(moduleURL)
	if err != nil {
		return err
	}
	release, err := fetchLatestRelease(owner, repo)
	if err != nil {
		return err
	}
	asset, err := matchAsset(release)
	if err != nil {
		return err
	}
	return downloadAsset(owner, repo, asset.ID, dst)
}

type githubRelease struct {
	Assets []githubAsset `json:"assets"`
}

type githubAsset struct {
	ID                 int64  `json:"id"`
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// fetchLatestRelease GETs /repos/<owner>/<repo>/releases/latest and decodes it.
func fetchLatestRelease(owner, repo string) (*githubRelease, error) {
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)
	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("building GitHub API request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "ade-tool")
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("calling GitHub API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned %s for %s/%s", resp.Status, owner, repo)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("decoding GitHub API response: %w", err)
	}
	return &release, nil
}

// matchAsset selects the first release asset whose filename contains the
// current GOOS and GOARCH (case-insensitive).
func matchAsset(release *githubRelease) (*githubAsset, error) {
	goos := runtime.GOOS
	goarch := runtime.GOARCH
	for i := range release.Assets {
		name := strings.ToLower(release.Assets[i].Name)
		if strings.Contains(name, goos) && strings.Contains(name, goarch) {
			return &release.Assets[i], nil
		}
	}
	return nil, fmt.Errorf(
		"no release asset found for %s/%s; asset filenames must include OS and architecture (e.g., my-plugin-%s-%s)",
		goos, goarch, goos, goarch,
	)
}

// downloadAsset streams the binary at /repos/<owner>/<repo>/releases/assets/<id>
// to dst.
//
// The API URL (rather than browser_download_url) is intentional: it ensures
// the Authorization header reaches only api.github.com. The CDN redirect
// receives a pre-signed URL and does not need the token, which is the
// recommended approach for private repository assets.
func downloadAsset(owner, repo string, assetID int64, dst string) error {
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/assets/%d", owner, repo, assetID)

	if err := os.MkdirAll(filepath.Dir(dst), 0750); err != nil {
		return fmt.Errorf("creating plugin directory: %w", err)
	}

	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		return fmt.Errorf("building download request: %w", err)
	}
	req.Header.Set("Accept", "application/octet-stream")
	req.Header.Set("User-Agent", "ade-tool")
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("downloading: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download returned %s", resp.Status)
	}

	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0750)
	if err != nil {
		return fmt.Errorf("creating destination file %s: %w", dst, err)
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("writing downloaded binary: %w", err)
	}
	return nil
}
