package resource

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func Put(request PutRequest, manager Github, inputDir string) (*PutResponse, error) {
	if err := request.Params.Validate(); err != nil {
		return nil, fmt.Errorf("invalid parameters: %s", err)
	}
	prNumber := request.Params.Number
	if prNumber == 0 {
		prNumber = request.Source.Number
	}
	if prNumber == 0 {
		return nil, fmt.Errorf("must set source.number or params.number")
	}
	path := filepath.Join(inputDir, request.Params.Path, ".git")

	refBytes, err := ioutil.ReadFile(filepath.Join(path, "ref"))
	if err != nil {
		return nil, fmt.Errorf("failed to read ref from path: %s", err)
	}
	ref := strings.TrimSpace(string(refBytes))

	// Set status if specified
	if p := request.Params; p.Status != "" {
		description := p.Description

		if err := manager.UpdateCommitStatus(ref, p.BaseContext, safeExpandEnv(p.Context), p.Status, safeExpandEnv(p.TargetURL), description); err != nil {
			return nil, fmt.Errorf("failed to set status: %s", err)
		}
	}

	// Delete previous comments if specified
	if request.Params.DeletePreviousComments {
		err = manager.DeletePreviousComments(prNumber)
		if err != nil {
			return nil, fmt.Errorf("failed to delete previous comments: %s", err)
		}
	}

	// Set comment if specified
	if p := request.Params; p.Comment != "" {
		err = manager.PostComment(prNumber, safeExpandEnv(p.Comment))
		if err != nil {
			return nil, fmt.Errorf("failed to post comment: %s", err)
		}
	}

	return &PutResponse{
		Version: Version{},
	}, nil
}

// PutRequest ...
type PutRequest struct {
	Source Source        `json:"source"`
	Params PutParameters `json:"params"`
}

// PutResponse ...
type PutResponse struct {
	Version Version `json:"version"`
}

// PutParameters for the resource.
type PutParameters struct {
	Path                   string `json:"path"`
	Number                 int    `json:"number"`
	BaseContext            string `json:"base_context"`
	Context                string `json:"context"`
	TargetURL              string `json:"target_url"`
	Description            string `json:"description"`
	Status                 string `json:"status"`
	Comment                string `json:"comment"`
	DeletePreviousComments bool   `json:"delete_previous_comments"`
}

// Validate the put parameters.
func (p *PutParameters) Validate() error {
	if p.Status == "" {
		return nil
	}
	// Make sure we are setting an allowed status
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
