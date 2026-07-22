package services

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"gitxyz/internal/helper"
	"gitxyz/pkg/git"

	"gorm.io/gorm"
)

// ErrMergeConflict indicates a patch cannot be merged due to conflicts.
var ErrMergeConflict = errors.New("merge conflict")

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

	// Patch plumbing helpers (used by PatchService).
	BranchExists(repoID, branch string) (bool, error)
	BranchHead(repoID, branch string) (string, error)
	CreateRef(repoID, ref, sha string) error
	SnapshotCommits(repoID, baseSHA, headSHA string) ([]CommitInfo, error)
	SnapshotFiles(repoID, baseSHA, headSHA string) ([]PatchFileSnapshot, error)
	CheckMergeable(repoID, targetSHA, sourceSHA string) (bool, error)
	PerformMerge(repoID, targetBranch, targetSHA, sourceSHA, message, authorName, authorEmail string) (string, error)
}

// PatchFileSnapshot is a single changed file in a patch snapshot.
type PatchFileSnapshot struct {
	FilePath string
	Status   string
	Diff     string
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

// BranchExists reports whether refs/heads/<branch> exists in the repo.
func (s *GitQueryServiceImpl) BranchExists(repoID, branch string) (bool, error) {
	_, err := s.runGit(repoID, "show-ref", "--verify", "--quiet", "refs/heads/"+branch)
	if err != nil {
		// git exits non-zero when the ref is missing; treat that as "not exists".
		return false, nil
	}
	return true, nil
}

// BranchHead returns the commit SHA at the tip of refs/heads/<branch>.
func (s *GitQueryServiceImpl) BranchHead(repoID, branch string) (string, error) {
	out, err := s.runGit(repoID, "rev-parse", "refs/heads/"+branch)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// CreateRef creates/updates a ref to point at sha (used to pin source commits
// so they are not garbage collected).
func (s *GitQueryServiceImpl) CreateRef(repoID, ref, sha string) error {
	_, err := s.runGit(repoID, "update-ref", ref, sha)
	return err
}

// SnapshotCommits returns the commits in baseSHA..headSHA.
func (s *GitQueryServiceImpl) SnapshotCommits(repoID, baseSHA, headSHA string) ([]CommitInfo, error) {
	out, err := s.runGit(repoID, "log", baseSHA+".."+headSHA,
		"--pretty=format:%H|%an|%ae|%ad|%s", "--date=iso-strict")
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	result := make([]CommitInfo, 0, len(lines))
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l == "" {
			continue
		}
		parts := strings.Split(l, "|")
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
	return result, nil
}

// SnapshotFiles returns the changed files between baseSHA and headSHA, including
// per-file diff text.
func (s *GitQueryServiceImpl) SnapshotFiles(repoID, baseSHA, headSHA string) ([]PatchFileSnapshot, error) {
	out, err := s.runGit(repoID, "diff", "--name-status", baseSHA+"..."+headSHA)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	result := make([]PatchFileSnapshot, 0, len(lines))
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l == "" {
			continue
		}
		fields := strings.Fields(l)
		if len(fields) < 2 {
			continue
		}
		status := fields[0]
		filePath := fields[len(fields)-1]
		// Normalize status prefix (e.g. "M", "M100") -> "modified".
		normalized := normalizeFileStatus(status)
		diff, derr := s.runGit(repoID, "diff", "--no-color", baseSHA+"..."+headSHA, "--", filePath)
		if derr != nil {
			diff = ""
		}
		result = append(result, PatchFileSnapshot{
			FilePath: filePath,
			Status:   normalized,
			Diff:     diff,
		})
	}
	return result, nil
}

func normalizeFileStatus(status string) string {
	switch {
	case strings.HasPrefix(status, "A"):
		return "added"
	case strings.HasPrefix(status, "D"):
		return "deleted"
	case strings.HasPrefix(status, "R"):
		return "renamed"
	default:
		return "modified"
	}
}

// CheckMergeable reports whether targetSHA can be merged into sourceSHA cleanly.
// Uses git merge-tree --write-tree; exit 0 = clean, exit 1 = conflict.
func (s *GitQueryServiceImpl) CheckMergeable(repoID, targetSHA, sourceSHA string) (bool, error) {
	_, err := s.runGit(repoID, "merge-tree", "--write-tree", targetSHA, sourceSHA)
	if err != nil {
		// merge-tree exits 1 on conflict; any other error is a real failure.
		if isExitCode(err, 1) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// PerformMerge merges sourceSHA into targetBranch using plumbing:
// merge-tree --write-tree -> commit-tree -> update-ref (CAS). Retries up to 3
// times if the target branch moved concurrently. Returns the new merge commit SHA.
func (s *GitQueryServiceImpl) PerformMerge(repoID, targetBranch, targetSHA, sourceSHA, message, authorName, authorEmail string) (string, error) {
	path := s.storagePath(repoID)
	gitBin := git.NewCommand().Executable()

	for attempt := 0; attempt < 3; attempt++ {
		// 1. Compute the merge tree.
		treeOut, err := s.runGit(repoID, "merge-tree", "--write-tree", targetSHA, sourceSHA)
		if err != nil {
			if isExitCode(err, 1) {
				return "", ErrMergeConflict
			}
			return "", err
		}
		treeSHA := strings.TrimSpace(treeOut)

		// 2. Create the merge commit.
		commitCmd := exec.Command(gitBin, "-C", path, "commit-tree", treeSHA,
			"-p", targetSHA, "-p", sourceSHA, "-m", message)
		commitCmd.Env = append(commitCmd.Environ(),
			"GIT_AUTHOR_NAME="+authorName,
			"GIT_AUTHOR_EMAIL="+authorEmail,
			"GIT_COMMITTER_NAME="+authorName,
			"GIT_COMMITTER_EMAIL="+authorEmail,
		)
		commitOut, err := commitCmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("git commit-tree: %w: %s", err, string(commitOut))
		}
		mergeCommit := strings.TrimSpace(string(commitOut))

		// 3. Atomically update the ref (CAS against targetSHA).
		_, err = s.runGit(repoID, "update-ref", "refs/heads/"+targetBranch, mergeCommit, targetSHA)
		if err != nil {
			// Branch moved concurrently; refresh targetSHA and retry.
			newTarget, hErr := s.BranchHead(repoID, targetBranch)
			if hErr != nil {
				return "", hErr
			}
			targetSHA = newTarget
			continue
		}
		return mergeCommit, nil
	}
	return "", fmt.Errorf("merge failed after retries: target branch moved concurrently")
}

// isExitCode reports whether err is an *exec.ExitError with the given code.
func isExitCode(err error, code int) bool {
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode() == code
	}
	return false
}

// ensure GitQueryServiceImpl satisfies the interface (compile-time check).
var _ GitQueryService = (*GitQueryServiceImpl)(nil)
