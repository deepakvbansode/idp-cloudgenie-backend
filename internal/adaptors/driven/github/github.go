package github

import (
	"context"
	"fmt"

	"github.com/deepakvbansode/idp-cloudgenie-backend/internal/common/constants"
	"github.com/deepakvbansode/idp-cloudgenie-backend/internal/config"
	"github.com/deepakvbansode/idp-cloudgenie-backend/internal/core/ports"
	"github.com/google/go-github/v55/github"
	"golang.org/x/oauth2"
)

type GithubAdaptor struct {
	logger ports.Logger
	config config.GithubConfig
}

func NewGithubAdaptor(logger ports.Logger,config config.GithubConfig) *GithubAdaptor {
	return &GithubAdaptor{
		logger: logger,
		config: config,
	}
}

func (g *GithubAdaptor) PushXRDToRepo( ctx context.Context, xrd string, repo string, path string) error {
	log := g.logger.WithField("tradeId", ctx.Value(constants.TraceIDKey))
	log.Info("Pushing XRD to GitHub repo:", repo)
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: g.config.Token})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	owner := g.config.Owner
	branch := g.config.Branch

	// Get the current file SHA if it exists (for update)
	var sha *string
	fileContent, _, resp, err := client.Repositories.GetContents(ctx, owner, repo, path, &github.RepositoryContentGetOptions{Ref: branch})
	if err == nil && fileContent != nil {
		sha = fileContent.SHA
	} else if resp != nil && resp.StatusCode != 404 && err != nil {
		log.Error("Failed to check file existence:", err)
		return err
	}

	// Prepare commit message and content
	commitMsg := fmt.Sprintf("Upload XRD file to %s", path)
	// content := base64.StdEncoding.EncodeToString([]byte(xrd))

	opts := &github.RepositoryContentFileOptions{
		Message:   github.String(commitMsg),
		Content:   []byte(xrd),
		Branch:    github.String(branch),
		SHA:       sha, // nil for create, sha for update
		Committer: &github.CommitAuthor{Name: github.String("cloudgenie-bot"), Email: github.String("cloudgenie-bot@example.com")},
	}

	_, _, err = client.Repositories.CreateFile(ctx, owner, repo, path, opts)
	if err != nil {
		// If file exists, try UpdateFile
		if sha != nil {
			_, _, err = client.Repositories.UpdateFile(ctx, owner, repo, path, opts)
			if err != nil {
				g.logger.Error("Failed to update file in GitHub repo:", err)
				return err
			}
		} else {
			log.Error("Failed to create file in GitHub repo:", err)
			return err
		}
	}
	log.Info("XRD file pushed to repo successfully at", path)
	return nil
}