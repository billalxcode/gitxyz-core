package git

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/spf13/viper"
)

type Git struct {
	executable string
}

func NewCommand() *Git {
	executable := viper.GetString("executable")

	return &Git{
		executable: executable,
	}
}

// Executable returns the configured git binary path.
func (g *Git) Executable() string {
	return g.executable
}

func (g *Git) IsBareRepository(path string) (bool, error) {
	out, err := exec.Command(
		g.executable,
		"-C", path,
		"rev-parse",
		"--is-bare-repository",
	).CombinedOutput()

	if err != nil {
		return false, fmt.Errorf("%v: %s", err, out)
	}

	return strings.TrimSpace(string(out)) == "true", nil
}

func (g *Git) InitBare(repo string) ([]byte, error) {
	return exec.Command(
		g.executable,
		"init",
		"--bare",
		repo,
	).CombinedOutput()
}

func (g *Git) ReceivePack(repo string) ([]byte, error) {
	return exec.Command(
		g.executable,
		"receive-pack",
		"--stateless-rpc",
		"--advertise-refs",
		repo,
	).CombinedOutput()
}

func (g *Git) ReceivePackRPC(
	repo string,
	in io.Reader,
	out io.Writer,
) error {
	cmd := exec.Command(
		g.executable,
		"receive-pack",
		"--stateless-rpc",
		repo,
	)

	cmd.Stdin = in
	cmd.Stdout = out

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf(
			"git receive-pack: %w: %s",
			err,
			stderr.String(),
		)
	}

	return nil
}

func (g *Git) UploadPackRPC(
	repo string,
	in io.Reader,
	out io.Writer,
) error {
	cmd := exec.Command(
		g.executable,
		"upload-pack",
		"--stateless-rpc",
		repo,
	)

	cmd.Stdin = in
	cmd.Stdout = out

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf(
			"git upload-pack: %w: %s",
			err,
			stderr.String(),
		)
	}

	return nil
}
