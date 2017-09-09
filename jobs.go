package gogitlab

import (
	"time"
	"path"
	"net/url"
	"fmt"
	"net/http"
	"strconv"
	"encoding/json"
)

var (
	pipelineJobsUrl = path.Join(pipelineUrl, "jobs")
)

type Job struct {
	Commit Commit `json:"commit"`
	CreatedAt   *time.Time `json:"created_at"`
	FinishedAt  *time.Time `json:"finished_at"`
	StartedAt   *time.Time `json:"started_at"`
	Id int `json:"id"`
	Name string `json:"name"`
	Pipeline PipelineBrief `json:"pipeline"`
	Ref string `json:"ref"`
	Stage string `json:"stage"`
	Status string `json:"status"`
	Tag bool `json:"tag"`
	User User `json:"user"`
}

type ListJobsOpts struct {
	scopes map[JobScope]bool
	Pagination
}

type JobScope int8

func (s JobScope) String() string {
	switch s {
	case JobScopeCreated:
		return "created"
	case JobScopePending:
		return "pending"
	case JobScopeRunning:
		return "running"
	case JobScopeFailed:
		return "failed"
	case JobScopeSuccess:
		return "success"
	case JobScopeCanceled:
		return "canceled"
	case JobScopeSkipped:
		return "skipped"
	case JobScopeManual:
		return "manual"
	}
	return ""
}

const (
	JobScopeCreated JobScope = 1 + iota
	JobScopePending
	JobScopeRunning
	JobScopeFailed
	JobScopeSuccess
	JobScopeCanceled
	JobScopeSkipped
	JobScopeManual
)

func (o *ListJobsOpts) AddScope(scopes ...JobScope) {
	if o.scopes == nil {
		o.scopes = make(map[JobScope]bool)
	}
	for _, s := range scopes {
		o.scopes[s] = true
	}
}

func (o *ListJobsOpts) toQueryValues() (url.Values, error) {
	if nil == o {
		return nil, nil
	}

	if err := o.check(); nil != err {
		return nil, err
	}
	query := make(url.Values)
	for s, _ := range o.scopes {
		query.Add("scope[]", s.String())
	}
	o.Pagination.toQueryValues(query)
	return query, nil
}

func (g *Gitlab) ListPipelineJobs(projId string, pipelineId int, opts *ListJobsOpts) ([]*Job, error) {
	query, err := opts.toQueryValues()
	if nil != err {
		return nil, fmt.Errorf("Check list jobs parameters error: %v", err)
	}

	data, err := g.buildAndExecRequest(
		http.MethodGet,
		g.ResourceUrlWithQueryValues(
			pipelineJobsUrl,
			map[string]string{
				":id": projId,
				":pipeline_id": strconv.Itoa(pipelineId),
			},
			query,
		),
		nil,
	)
	if nil != err {
		return nil, fmt.Errorf("Request list pipeline jobs API error: %v", err)
	}

	var js []*Job
	if err := json.Unmarshal(data, &js); nil != err {
		return nil, fmt.Errorf("Decode response error: %v", err)
	}

	return js, nil
}