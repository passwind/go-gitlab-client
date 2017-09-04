package gogitlab

import (
	"path"
	"net/http"
	"encoding/json"
)

type Pipeline struct {
	Id int `json:"id"`
	SHA string `json:"sha"`
	Ref string `json:"ref"`
	Status string `json:"status"`
	BeforeSHA string `json:"before_sha"`
	Tag bool `json:"tag"`
	User User `json:"user"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	StartedAt *string `json:"started_at"`
	FinishedAt *string `json:"finished_at"`
	CommittedAt *string `json:"committed_at"`
	Duration *int64 `json:"duration"`
}

// Create pipeline for specified project
func (g *Gitlab) CreatePipeline(pid, ref string) (*Pipeline, error) {
	var pl *Pipeline
	data, err := g.buildAndExecRequest(
		http.MethodPost,
		g.ResourceUrlWithQuery(
			path.Join(project_url, "pipeline"),
			map[string]string{":id": pid},
			map[string]string{"ref": ref}),
		nil,
	)
	if nil != err {
		return nil, fmt.Errorf("Request API error: %v", err)
	}

	if err := json.Unmarshal(data, *pl); nil != err {
		return nil, fmt.Errorf("Decode response error: %v", err)
	}

	return pl, nil
}
