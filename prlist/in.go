package prlist

import (
	"io/ioutil"
	"path/filepath"
)

func Get(request GetRequest, outputDir string) (*GetResponse, error) {
	path := filepath.Join(outputDir, "prs.json")
	if err := ioutil.WriteFile(path, []byte(request.Version.PRs), 0644); err != nil {
		return nil, err
	}
	return &GetResponse{
		Version: request.Version,
	}, nil
}

type GetRequest struct {
	Source  Source  `json:"source"`
	Version Version `json:"version"`
}

type GetResponse struct {
	Version Version `json:"version"`
}
