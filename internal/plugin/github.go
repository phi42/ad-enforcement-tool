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
// a GitHub release of moduleURL and writes it to dst. It returns the release
// tag that was actually downloaded.
//
// moduleURL may include an optional "@version" suffix to pin a specific
// release tag (e.g. "github.com/owner/repo@v1.2.3"). When no version is
// given the latest release is used.
func FetchRelease(moduleURL, dst string) (tagName string, err error) {
	base, version := SplitVersion(moduleURL)
	owner, repo, parseErr := ParseModuleURL(base)
	if parseErr != nil {
		return "", parseErr
	}
	var release *githubRelease
	if version == "" {
		release, err = fetchLatestRelease(owner, repo)
	} else {
		release, err = fetchReleaseByTag(owner, repo, version)
	}
	if err != nil {
		return "", err
	}
	asset, err := matchAsset(release)
	if err != nil {
		return "", err
	}
	if err := downloadAsset(owner, repo, asset.ID, dst); err != nil {
		return "", err
	}
	return release.TagName, nil
}

type githubRelease struct {
	TagName string        `json:"tag_name"`
	Assets  []githubAsset `json:"assets"`
}

type githubAsset struct {
	ID                 int64  `json:"id"`
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// FetchLatestTag returns the tag name of the latest GitHub release for
// moduleURL (e.g. "github.com/owner/repo") without downloading the binary.
func FetchLatestTag(moduleURL string) (string, error) {
	owner, repo, err := ParseModuleURL(moduleURL)
	if err != nil {
		return "", err
	}
	release, err := fetchLatestRelease(owner, repo)
	if err != nil {
		return "", err
	}
	return release.TagName, nil
}

// fetchReleaseByTag GETs /repos/<owner>/<repo>/releases/tags/<tag> and decodes it.
func fetchReleaseByTag(owner, repo, tag string) (*githubRelease, error) {
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/tags/%s", owner, repo, tag)
	return doGitHubReleaseRequest(apiURL, owner, repo)
}

// fetchLatestRelease GETs /repos/<owner>/<repo>/releases/latest and decodes it.
func fetchLatestRelease(owner, repo string) (*githubRelease, error) {
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)
	return doGitHubReleaseRequest(apiURL, owner, repo)
}

// doGitHubReleaseRequest performs a GET against a GitHub releases API URL and
// decodes the JSON response into a githubRelease.
func doGitHubReleaseRequest(apiURL, owner, repo string) (*githubRelease, error) {
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
		// A 403 with X-RateLimit-Remaining: 0 means the unauthenticated rate
		// limit (60 req/h) has been exhausted. Give an actionable message.
		if resp.StatusCode == http.StatusForbidden &&
			resp.Header.Get("X-RateLimit-Remaining") == "0" {
			return nil, fmt.Errorf(
				"GitHub API rate limit exceeded for %s/%s; "+
					"set the GITHUB_TOKEN environment variable to raise the limit",
				owner, repo,
			)
		}
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
