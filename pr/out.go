package pr

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudfoundry-community/github-pr-instances-resource/models"
)

func Put(request PutRequest, github models.Github, inputDir string) (*PutResponse, error) {
	if err := request.Params.Validate(); err != nil {
		return nil, fmt.Errorf("invalid parameters: %s", err)
	}

	path := filepath.Join(inputDir, request.Params.Path, ".git", "resource")

	// Version available after a GET step.
	var version Version
	content, err := ioutil.ReadFile(filepath.Join(path, "version.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to read version from path: %v", err)
	}
	if err := json.Unmarshal(content, &version); err != nil {
		return nil, fmt.Errorf("failed to unmarshal version from file: %v", err)
	}

	// Metadata available after a GET step.
	var metadata models.Metadata
	content, err = ioutil.ReadFile(filepath.Join(path, "metadata.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata from path: %v", err)
	}
	if err := json.Unmarshal(content, &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata from file: %v", err)
	}

	// Set status if specified
	if p := request.Params; p.Status != "" {
		description := p.Description

		if err := github.UpdateCommitStatus(version.Ref, p.BaseContext, safeExpandEnv(p.Context), p.Status, safeExpandEnv(p.TargetURL), description); err != nil {
			return nil, fmt.Errorf("failed to set status: %v", err)
		}
	}

	prNumber := request.Source.Number

	// Delete previous comments if specified
	if request.Params.DeletePreviousComments {
		err = github.DeletePreviousComments(prNumber)
		if err != nil {
			return nil, fmt.Errorf("failed to delete previous comments: %v", err)
		}
	}

	// Set comment if specified
	if p := request.Params; p.Comment != "" {
		err = github.PostComment(prNumber, safeExpandEnv(p.Comment))
		if err != nil {
			return nil, fmt.Errorf("failed to post comment: %v", err)
		}
	}

	// Set comment from a file if specfied
	if p := request.Params; p.CommentFile != "" {
		content, err := ioutil.ReadFile(filepath.Join(inputDir, p.CommentFile))
		if err != nil {
			return nil, fmt.Errorf("failed to read comment file: %s", err)
		}
		comment := string(content)
		if comment != "" {
			err = github.PostComment(prNumber, safeExpandEnv(comment))
			if err != nil {
				return nil, fmt.Errorf("failed to post comment: %s", err)
			}
		}
	}

	return &PutResponse{
		Version:  version,
		Metadata: metadata,
	}, nil
}

type PutRequest struct {
	Source Source        `json:"source"`
	Params PutParameters `json:"params"`
}

type PutResponse struct {
	Version  Version         `json:"version"`
	Metadata models.Metadata `json:"metadata,omitempty"`
}

type PutParameters struct {
	Path                   string `json:"path"`
	BaseContext            string `json:"base_context"`
	Context                string `json:"context"`
	TargetURL              string `json:"target_url"`
	Description            string `json:"description"`
	Status                 string `json:"status"`
	Comment                string `json:"comment"`
	CommentFile            string `json:"comment_file"`
	DeletePreviousComments bool   `json:"delete_previous_comments"`
}

func (p *PutParameters) Validate() error {
	if p.Status == "" {
		return nil
	}
	var allowedStatus bool

	status := strings.ToLower(p.Status)
	allowed := []string{"success", "pending", "failure", "error"}

	for _, a := range allowed {
		if status == a {
			allowedStatus = true
		}
	}

	if !allowedStatus {
		return fmt.Errorf("unknown status: %s", p.Status)
	}

	return nil
}

func safeExpandEnv(s string) string {
	return os.Expand(s, func(v string) string {
		switch v {
		case "BUILD_ID", "BUILD_NAME", "BUILD_JOB_NAME", "BUILD_PIPELINE_NAME", "BUILD_TEAM_NAME", "ATC_EXTERNAL_URL":
			return os.Getenv(v)
		}
		return "$" + v
	})
}
