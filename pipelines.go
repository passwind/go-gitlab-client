package gogitlab

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strconv"
	"time"
	"net/url"
)

var (
	pipelineCreationUrl = path.Join(project_url, "pipeline")
	pipelinesUrl        = path.Join(project_url, "pipelines")
	pipelineUrl			= path.Join(project_url, "pipelines", ":pipeline_id")
)

type Pipeline struct {
	PipelineBrief
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

type PipelineBrief struct {
	Id          int        `json:"id"`
	SHA         string     `json:"sha"`
	Ref         string     `json:"ref"`
	Status      string     `json:"status"`
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
			pipelineCreationUrl,
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

func (g *Gitlab) ListPipelines(pid string, opts *ListPipelinesOpts) ([]*PipelineBrief, error) {
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

	var ps []*PipelineBrief
	if err := json.Unmarshal(data, &ps); nil != err {
		return nil, fmt.Errorf("Decode response error: %v", err)
	}

	return ps, nil
}

func (g *Gitlab) GetPipeline(projId string, pipelineId int) (*Pipeline, error) {
	data, err := g.buildAndExecRequest(
		http.MethodGet,
		g.ResourceUrl(
			pipelineUrl,
			map[string]string{
				":id": projId,
				":pipeline_id": strconv.Itoa(pipelineId),
			},
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
		query["scopes"] = opts.Scope
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
	opts.Pagination.toQuery(query)
	return query, nil
}

func (opts *ListPipelinesOpts) check() error {
	if nil == opts {
		return nil
	}

	if "" != opts.Scope && !validPipelineScope[opts.Scope] {
		return fmt.Errorf("Invalid scopes '%s'", opts.Scope)
	}

	if "" != opts.Status && !validPipelineStatus[opts.Status] {
		return fmt.Errorf("Invalid status '%s'", opts.Status)
	}

	if "" != opts.OrderBy && !validPipelineOrder[opts.OrderBy] {
		return fmt.Errorf("Invalid order_by '%s'", opts.OrderBy)
	}

	return opts.Pagination.check()
}

func (p Pagination) check() error {
	if p.Page < 0 {
		return fmt.Errorf("Invalid page '%d'", p.Page)
	}

	if p.PerPage < 0 || p.PerPage > 100 {
		return fmt.Errorf("Invalid per_page '%d'", p.PerPage)
	}

	return nil
}

func (p Pagination) toQuery(query map[string]string) {
	if p.Page > 0 {
		query["page"] = strconv.Itoa(p.Page)
	}
	if p.PerPage > 0 {
		query["per_page"] = strconv.Itoa(p.PerPage)
	}
}

func (p Pagination) toQueryValues(vals url.Values) {
	if p.Page > 0 {
		vals.Set("page", strconv.Itoa(p.Page))
	}
	if p.PerPage > 0 {
		vals.Set("per_page", strconv.Itoa(p.PerPage))
	}
}