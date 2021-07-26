package prlist

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
)

func Get(request GetRequest, outputDir string) (*GetResponse, error) {
	path := filepath.Join(outputDir, "prs.json")
	var prNumbers []int
	if err := json.Unmarshal([]byte(request.Version.PRs), &prNumbers); err != nil {
		return nil, err
	}

	prs := make([]PRData, 0, len(prNumbers))
	for _, prNumber := range prNumbers {
		prs = append(prs, PRData{Number: prNumber})
	}

	payload, err := json.Marshal(prs)
	if err != nil {
		return nil, err
	}

	if err := ioutil.WriteFile(path, payload, 0644); err != nil {
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
