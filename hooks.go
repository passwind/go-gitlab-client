package gogitlab

import (
	"encoding/json"
	"net/url"
)

const (
	project_url_hooks = "/projects/:id/hooks"          // Get list of project hooks
	project_url_hook  = "/projects/:id/hooks/:hook_id" // Get single project hook
)

type Hook struct {
	Id           int    `json:"id,omitempty"`
	Url          string `json:"url,omitempty"`
	CreatedAtRaw string `json:"created_at,omitempty"`
	HookFlags
}

/*
Get list of project hooks.

    GET /projects/:id/hooks

Parameters:

    id The ID of a project

*/
func (g *Gitlab) ProjectHooks(id string) ([]*Hook, error) {

	url, opaque := g.ResourceUrlRaw(project_url_hooks, map[string]string{":id": id})

	var err error
	var hooks []*Hook

	contents, err := g.buildAndExecRequestRaw("GET", url, opaque, nil)
	if err != nil {
		return hooks, err
	}

	err = json.Unmarshal(contents, &hooks)

	return hooks, err
}

/*
Get single project hook.

    GET /projects/:id/hooks/:hook_id

Parameters:

    id      The ID of a project
    hook_id The ID of a hook

*/
func (g *Gitlab) ProjectHook(id, hook_id string) (*Hook, error) {

	url, opaque := g.ResourceUrlRaw(project_url_hook, map[string]string{
		":id":      id,
		":hook_id": hook_id,
	})

	var err error
	hook := new(Hook)

	contents, err := g.buildAndExecRequestRaw("GET", url, opaque, nil)
	if err != nil {
		return hook, err
	}

	err = json.Unmarshal(contents, &hook)

	return hook, err
}

type HookFlags struct {
	PushEvents            bool `json:"push_events"`
	IssuesEvents          bool `json:"issues_events"`
	MergeRequestsEvents   bool `json:"merge_requests_events"`
	TagPushEvents         bool `json:"tag_push_events"`
	NoteEvents            bool `json:"note_events"`
	JobEvents             bool `json:"job_events"`
	PipelineEvents        bool `json:"pipeline_events"`
	WikiEvents            bool `json:"wiki_events"`
	EnableSSLVerification bool `json:"enable_ssl_verification"`
}

func DefaultHookFlags() HookFlags {
	return HookFlags{
		PushEvents: true,
		EnableSSLVerification: true,
	}
}

func (g *Gitlab) AddProjectHookWithFlags(id, hook_url string, hookFlags HookFlags) error {
	url, opaque := g.ResourceUrlRaw(project_url_hooks, map[string]string{":id": id})
	var err error
	body := buildHookQueryWithFlags(hook_url, hookFlags)
	_, err = g.buildAndExecRequestRaw("POST", url, opaque, body)

	return err
}

/*
Add new project hook.

    POST /projects/:id/hooks

Parameters:

    id                    The ID or NAMESPACE/PROJECT_NAME of a project
    hook_url              The hook URL
    push_events           Trigger hook on push events
    issues_events         Trigger hook on issues events
    merge_requests_events Trigger hook on merge_requests events

*/
func (g *Gitlab) AddProjectHook(id, hook_url string, push_events, issues_events, merge_requests_events bool) error {

	url, opaque := g.ResourceUrlRaw(project_url_hooks, map[string]string{":id": id})

	var err error

	body := buildHookQuery(hook_url, push_events, issues_events, merge_requests_events)
	_, err = g.buildAndExecRequestRaw("POST", url, opaque, []byte(body))

	return err
}

/*
Edit existing project hook.

    PUT /projects/:id/hooks/:hook_id

Parameters:

    id                    The ID or NAMESPACE/PROJECT_NAME of a project
    hook_id               The ID of a project hook
    hook_url              The hook URL
    push_events           Trigger hook on push events
    issues_events         Trigger hook on issues events
    merge_requests_events Trigger hook on merge_requests events

*/
func (g *Gitlab) EditProjectHook(id, hook_id, hook_url string, push_events, issues_events, merge_requests_events bool) error {

	url, opaque := g.ResourceUrlRaw(project_url_hook, map[string]string{
		":id":      id,
		":hook_id": hook_id,
	})

	var err error

	body := buildHookQuery(hook_url, push_events, issues_events, merge_requests_events)
	_, err = g.buildAndExecRequestRaw("PUT", url, opaque, []byte(body))

	return err
}

/*
Remove hook from project.

    DELETE /projects/:id/hooks/:hook_id

Parameters:

    id      The ID or NAMESPACE/PROJECT_NAME of a project
    hook_id The ID of hook to delete

*/
func (g *Gitlab) RemoveProjectHook(id, hook_id string) error {

	url, opaque := g.ResourceUrlRaw(project_url_hook, map[string]string{
		":id":      id,
		":hook_id": hook_id,
	})

	var err error

	_, err = g.buildAndExecRequestRaw("DELETE", url, opaque, nil)

	return err
}

/*
Build HTTP query to add or edit hook
*/
func buildHookQuery(hook_url string, push_events, issues_events, merge_requests_events bool) string {

	v := url.Values{}
	v.Set("url", hook_url)

	if push_events {
		v.Set("push_events", "true")
	} else {
		v.Set("push_events", "false")
	}
	if issues_events {
		v.Set("issues_events", "true")
	} else {
		v.Set("issues_events", "false")
	}
	if merge_requests_events {
		v.Set("merge_requests_events", "true")
	} else {
		v.Set("merge_requests_events", "false")
	}

	return v.Encode()
}

func buildHookQueryWithFlags(hook_url string, hookFlags HookFlags) []byte {
	b, _ := json.Marshal(Hook{
		Url: hook_url,
		HookFlags: hookFlags,
	})

	return b
}
