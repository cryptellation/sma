package main

import (
	"context"
	"strings"

	"github.com/cryptellation/sma/dagger/internal/dagger"
)

// Git provides access to a git repository.
type Git struct {
	container *dagger.Container
}

// NewGitOptions contains the options for creating a new Git container.
type NewGitOptions struct {
	SrcDir *dagger.Directory
	User   *string
	Token  *dagger.Secret
}

// NewGit creates a new Git container with the given source directory and token.
func NewGit(ctx context.Context, opts NewGitOptions) (Git, error) {
	var err error

	// Create container
	container := dag.Container().
		From("alpine/git").
		WithMountedDirectory("/git", opts.SrcDir).
		WithWorkdir("/git").
		WithoutEntrypoint()

	// Set user/token if provided
	if opts.User != nil && opts.Token != nil {
		// Set authentication based on the token
		tokenString, err := opts.Token.Plaintext(ctx)
		if err != nil {
			return Git{}, err
		}

		// Change the url to use the token
		container, err = container.WithExec([]string{
			"git", "remote", "set-url", "origin", "https://" + *opts.User + ":" + tokenString + "@github.com/cryptellation/sma.git",
		}).Sync(ctx)
		if err != nil {
			return Git{}, err
		}
	}

	// Set Git author
	container, err = setGitAuthor(ctx, container)
	if err != nil {
		return Git{}, err
	}

	return Git{
		container: container,
	}, nil
}

// GetLastCommit returns the last commit SHA.
func (g *Git) GetLastCommitShortSHA(ctx context.Context) (string, error) {
	res, err := g.container.
		WithExec([]string{"git", "rev-parse", "--short", "HEAD"}).
		Stdout(ctx)
	if err != nil {
		return "", err
	}

	// Remove potential new line
	res = strings.TrimSuffix(res, "\n")

	return res, nil
}

func (g *Git) GetActualBranch(ctx context.Context) (string, error) {
	res, err := g.container.
		WithExec([]string{"git", "rev-parse", "--abbrev-ref", "HEAD"}).
		Stdout(ctx)
	if err != nil {
		return "", err
	}

	// Remove potential new line
	res = strings.TrimSuffix(res, "\n")

	return res, nil
}

// GetLastCommitTitle returns the title of the last commit.
func (g *Git) GetLastCommitTitle(ctx context.Context) (string, error) {
	res, err := g.container.
		WithExec([]string{"git", "log", "-1", "--pretty=%B"}).
		Stdout(ctx)
	if err != nil {
		return "", err
	}

	// Remove potential new line
	res = strings.TrimSuffix(res, "\n")

	return res, nil
}

// PublishTagFromReleaseTitle creates a new tag based on the last commit title.
// The title should follow the angular convention.
func (g *Git) PublishTagFromReleaseTitle(ctx context.Context) error {
	// Get new semver
	title, err := g.GetLastCommitTitle(ctx)
	if err != nil {
		return err
	}

	// Get last tag
	lastTag, err := g.GetLastTag(ctx)
	if err != nil {
		return err
	}

	// Process newSemVer change
	change, newSemVer, err := ProcessSemVerChange(lastTag, title)
	if err != nil {
		return err
	}
	if change == SemVerChangeNone {
		return nil
	}

	// Tag commit
	g.container, err = g.container.
		WithExec([]string{"git", "tag", "v" + newSemVer}).
		Sync(ctx)
	if err != nil {
		return err
	}

	// Push new tag
	g.container, err = g.container.
		WithExec([]string{"git", "push", "--tags"}).
		Sync(ctx)

	return err
}

// GetLastTag returns the last tag of the repository.
func (g *Git) GetLastTag(ctx context.Context) (string, error) {
	res, err := g.container.
		WithExec([]string{"git", "describe", "--tags", "--abbrev=0"}).
		Stdout(ctx)
	if err != nil {
		return "", err
	}

	// Remove potential new line
	res = strings.TrimSuffix(res, "\n")

	return res, nil
}

func setGitAuthor(
	ctx context.Context,
	container *dagger.Container,
) (*dagger.Container, error) {
	// Add infos on author
	container, err := container.
		WithExec([]string{"git", "config", "--global", "user.email", "louis.fradin+cryptellation-ci@gmail.com"}).
		Sync(ctx)
	if err != nil {
		return nil, err
	}
	container, err = container.
		WithExec([]string{"git", "config", "--global", "user.name", "Cryptellation CI"}).
		Sync(ctx)
	if err != nil {
		return nil, err
	}

	return container, nil
}
