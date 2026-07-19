package services

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"gitxyz/internal/helper"
	"gitxyz/pkg/git"

	"gorm.io/gorm"
)

// GitQueryService exposes read-only repository inspection (branches, commits,
// file contents, tree listing) by shelling out to the git executable against the
// on-disk bare repository storage.
type GitQueryService interface {
	ListBranches(repoID string) ([]string, error)
	DeleteBranch(repoID, branch string) error
	ListCommits(repoID, ref string, limit int) ([]CommitInfo, error)
	GetCommit(repoID, sha string) (*CommitInfo, error)
	GetContents(repoID, ref, path string) ([]ContentEntry, error)
	GetFile(repoID, ref, path string) ([]byte, error)
}

type GitQueryServiceImpl struct {
	db *gorm.DB
}

func NewGitQueryService(db *gorm.DB) GitQueryService {
	return &GitQueryServiceImpl{db: db}
}

// storagePath resolves the on-disk location of a repository from its ID.
func (s *GitQueryServiceImpl) storagePath(repoID string) string {
	return helper.RepositoryStoragePath(repoID)
}

func (s *GitQueryServiceImpl) runGit(repoID string, args ...string) (string, error) {
	path := s.storagePath(repoID)
	cmd := exec.Command(git.NewCommand().Executable(), "-C", path)
	cmd.Args = append(cmd.Args, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, string(out))
	}
	return string(out), nil
}

func (s *GitQueryServiceImpl) ListBranches(repoID string) ([]string, error) {
	out, err := s.runGit(repoID, "for-each-ref", "--format=%(refname:short)", "refs/heads")
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	result := make([]string, 0, len(lines))
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l != "" {
			result = append(result, l)
		}
	}
	return result, nil
}

func (s *GitQueryServiceImpl) DeleteBranch(repoID, branch string) error {
	_, err := s.runGit(repoID, "update-ref", "-d", "refs/heads/"+branch)
	return err
}

// CommitInfo is a subset of git commit metadata returned by the API.
type CommitInfo struct {
	SHA     string    `json:"sha"`
	Message string    `json:"message"`
	Author  string    `json:"author"`
	Email   string    `json:"email"`
	Date    time.Time `json:"date"`
}

func (s *GitQueryServiceImpl) ListCommits(repoID, ref string, limit int) ([]CommitInfo, error) {
	if limit <= 0 || limit > 100 {
		limit = 30
	}
	args := []string{"log", fmt.Sprintf("-%d", limit), "--format=%H%x1f%an%x1f%ae%x1f%aI%x1f%s", "--date=iso-strict"}
	if ref != "" {
		args = append(args, ref)
	}
	out, err := s.runGit(repoID, args...)
	if err != nil {
		return nil, err
	}
	return parseCommits(out), nil
}

func (s *GitQueryServiceImpl) GetCommit(repoID, sha string) (*CommitInfo, error) {
	out, err := s.runGit(repoID, "log", "-1", "--format=%H%x1f%an%x1f%ae%x1f%aI%x1f%s", "--date=iso-strict", sha)
	if err != nil {
		return nil, err
	}
	commits := parseCommits(out)
	if len(commits) == 0 {
		return nil, fmt.Errorf("commit not found: %s", sha)
	}
	return &commits[0], nil
}

func parseCommits(out string) []CommitInfo {
	lines := strings.Split(strings.TrimSpace(out), "\n")
	result := make([]CommitInfo, 0, len(lines))
	for _, l := range lines {
		if l == "" {
			continue
		}
		parts := strings.Split(l, "\x1f")
		if len(parts) < 5 {
			continue
		}
		date, _ := time.Parse(time.RFC3339, parts[3])
		result = append(result, CommitInfo{
			SHA:     parts[0],
			Author:  parts[1],
			Email:   parts[2],
			Date:    date,
			Message: parts[4],
		})
	}
	return result
}

// ContentEntry describes a single entry in a repository tree (file or dir).
type ContentEntry struct {
	Type string `json:"type"` // "file" or "dir"
	Name string `json:"name"`
	Path string `json:"path"`
	SHA  string `json:"sha"`
	Size int64  `json:"size"`
}

func (s *GitQueryServiceImpl) GetContents(repoID, ref, path string) ([]ContentEntry, error) {
	// Use ls-tree to list a directory (or show a single file).
	treeRef := ref
	if treeRef == "" {
		treeRef = "HEAD"
	}
	path = cleanPath(path)
	args := []string{"ls-tree", "-l", treeRef}
	if path != "" {
		args = append(args, "--", path)
	}
	out, err := s.runGit(repoID, args...)
	if err != nil {
		return nil, err
	}
	return parseTree(out), nil
}

func parseTree(out string) []ContentEntry {
	lines := strings.Split(strings.TrimSpace(out), "\n")
	result := make([]ContentEntry, 0, len(lines))
	for _, l := range lines {
		if l == "" {
			continue
		}
		// format: <mode> SP <type> SP <sha> SP <size> TAB <name>
		fields := strings.SplitN(l, "\t", 2)
		if len(fields) != 2 {
			continue
		}
		meta := strings.Fields(fields[0])
		if len(meta) < 4 {
			continue
		}
		objType := meta[1]
		sha := meta[2]
		name := fields[1]
		entryType := "file"
		if objType == "tree" {
			entryType = "dir"
		}
		// git ls-tree already returns the full path as the entry name, so use
		// it directly as both Name and Path.
		var size int64
		if entryType == "file" {
			fmt.Sscanf(meta[3], "%d", &size)
		}
		result = append(result, ContentEntry{
			Type: entryType,
			Name: name,
			Path: name,
			SHA:  sha,
			Size: size,
		})
	}
	return result
}

func (s *GitQueryServiceImpl) GetFile(repoID, ref, path string) ([]byte, error) {
	treeRef := ref
	if treeRef == "" {
		treeRef = "HEAD"
	}
	out, err := s.runGit(repoID, "show", treeRef+":"+cleanPath(path))
	if err != nil {
		return nil, err
	}
	return []byte(out), nil
}

// cleanPath removes a leading slash so the path is interpreted as repository
// relative rather than an absolute filesystem path by git.
func cleanPath(p string) string {
	return strings.TrimPrefix(p, "/")
}

// ensure GitQueryServiceImpl satisfies the interface (compile-time check).
var _ GitQueryService = (*GitQueryServiceImpl)(nil)
