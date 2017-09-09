package gogitlab

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strconv"
	"time"
)

var (
	pipelineUrl  = path.Join(project_url, "pipeline")
	pipelinesUrl = path.Join(project_url, "pipelines")
)

type Pipeline struct {
	Id          int        `json:"id"`
	SHA         string     `json:"sha"`
	Ref         string     `json:"ref"`
	Status      string     `json:"status"`
	BeforeSHA   string     `json:"before_sha"`
	Tag         bool       `json:"tag"`
	User        User       `json:"user"`
	CreatedAt   *time.Time `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
	StartedAt   *time.Time `json:"started_at"`
	FinishedAt  *time.Time `json:"finished_at"`
	CommittedAt *time.Time `json:"committed_at"`
	Duration    *int64     `json:"duration"`
}

func (p *Pipeline) Finished() bool {
	return p.Status != "" && p.Status != "running" && p.Status != "pending"
}

// Create pipeline for specified project
func (g *Gitlab) CreatePipeline(pid, ref string) (*Pipeline, error) {
	var pl Pipeline
	data, err := g.buildAndExecRequest(
		http.MethodPost,
		g.ResourceUrlWithQuery(
			pipelineUrl,
			map[string]string{":id": pid},
			map[string]string{"ref": ref},
		),
		nil,
	)
	if nil != err {
		return nil, fmt.Errorf("Request create pipeline API error: %v", err)
	}

	if err := json.Unmarshal(data, &pl); nil != err {
		return nil, fmt.Errorf("Decode response error: %v", err)
	}

	return &pl, nil
}

func (g *Gitlab) ListPipelines(pid string, opts *ListPipelinesOpts) ([]*Pipeline, error) {
	query, err := opts.toQuery()
	if nil != err {
		return nil, fmt.Errorf("Check list pipelines parameters error: %v", err)
	}

	data, err := g.buildAndExecRequest(
		http.MethodGet,
		g.ResourceUrlWithQuery(
			pipelinesUrl,
			map[string]string{":id": pid},
			query,
		),
		nil,
	)
	if nil != err {
		return nil, fmt.Errorf("Request list pipelines API error: %v", err)
	}

	var ps []*Pipeline
	if err := json.Unmarshal(data, &ps); nil != err {
		return nil, fmt.Errorf("Decode response error: %v", err)
	}

	return ps, nil
}

func (g *Gitlab) GetPipeline(projId string, pipelineId int) (*Pipeline, error) {
	data, err := g.buildAndExecRequest(
		http.MethodGet,
		g.ResourceUrl(
			path.Join(pipelinesUrl, strconv.Itoa(pipelineId)),
			map[string]string{":id": projId},
		),
		nil,
	)
	if nil != err {
		return nil, fmt.Errorf("Request get pipeline API error: %v", err)
	}

	var p *Pipeline
	if err := json.Unmarshal(data, &p); nil != err {
		return nil, fmt.Errorf("Decode response error: %v", err)
	}

	return p, nil
}

type ListPipelinesOpts struct {
	Scope      string
	Status     string
	Ref        string
	YamlErrors bool
	Name       string
	UserName   string
	OrderBy    string
	SortAsc    bool
	Pagination
}

type Pagination struct {
	Page    int
	PerPage int
}

var validPipelineScope, validPipelineStatus, validPipelineOrder map[string]bool

func init() {
	validPipelineScope = map[string]bool{
		"running":  true,
		"pending":  true,
		"finished": true,
		"branches": true,
		"tags":     true,
	}

	validPipelineStatus = map[string]bool{
		"running":  true,
		"pending":  true,
		"success":  true,
		"failed":   true,
		"canceled": true,
		"skipped":  true,
	}

	validPipelineOrder = map[string]bool{
		"id":      true,
		"status":  true,
		"ref":     true,
		"user_id": true,
	}
}

func (opts *ListPipelinesOpts) toQuery() (map[string]string, error) {
	if nil == opts {
		return nil, nil
	}

	if err := opts.check(); nil != err {
		return nil, err
	}

	query := make(map[string]string)
	if "" != opts.Scope {
		query["scope"] = opts.Scope
	}
	if "" != opts.Status {
		query["status"] = opts.Status
	}
	if "" != opts.Ref {
		query["ref"] = opts.Ref
	}
	if opts.YamlErrors {
		query["yaml_errors"] = "true"
	}
	if "" != opts.Name {
		query["name"] = opts.Name
	}
	if "" != opts.UserName {
		query["username"] = opts.UserName
	}
	if "" != opts.OrderBy {
		query["order_by"] = opts.OrderBy
	}
	if opts.SortAsc {
		query["sort"] = "asc"
	}
	if opts.Page > 0 {
		query["page"] = strconv.Itoa(opts.Page)
	}
	if opts.PerPage > 0 {
		query["per_page"] = strconv.Itoa(opts.PerPage)
	}
	return query, nil
}

func (opts *ListPipelinesOpts) check() error {
	if nil == opts {
		return nil
	}

	if "" != opts.Scope && !validPipelineScope[opts.Scope] {
		return fmt.Errorf("Invalid scope '%s'", opts.Scope)
	}

	if "" != opts.Status && !validPipelineStatus[opts.Status] {
		return fmt.Errorf("Invalid status '%s'", opts.Status)
	}

	if "" != opts.OrderBy && !validPipelineOrder[opts.OrderBy] {
		return fmt.Errorf("Invalid order_by '%s'", opts.OrderBy)
	}

	if opts.Page < 0 {
		return fmt.Errorf("Invalid page '%d'", opts.Page)
	}

	if opts.PerPage < 0 || opts.PerPage > 100 {
		return fmt.Errorf("Invalid per_page '%d'", opts.PerPage)
	}

	return nil
}
