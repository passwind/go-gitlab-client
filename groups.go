package gogitlab

import (
	"encoding/json"
)

const (
	groups_url         = "/groups"              // Get a list of groups owned by the authenticated user
	groups_search_url  = "/groups?search="      // Search for groups by name
	group_create_url   = "/groups"              // New group [POST]
	group_url_projects = "/groups/:id/projects" // Get group projects
)

type Group struct {
	Id          int    `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Path        string `json:"path,omitempty"`
	Description string `json:"description,omitempty"`
	Visibility  string `json:"description,omitempty"`
	ParentId    int    `json:"parent_id"`
}

func groups(u string, g *Gitlab) ([]*Group, error) {
	url := g.ResourceUrl(u, nil)

	var groups []*Group

	contents, err := g.buildAndExecRequest("GET", url, nil)
	if err == nil {
		err = json.Unmarshal(contents, &groups)
	}

	return groups, err
}

func (g *Gitlab) GroupSearch(search string) ([]*Group, error) {
	url := groups_search_url + search
	return groups(url, g)
}

/*
Create a group,
which is owned by the authentication user.
Namespaced project may be retrieved by specifying the namespace
and its project name like this:

	{"name":"ws8000","path":"ws8000","parent_id":35,"visibility":"internal"}

*/
func (g *Gitlab) CreateGroup(group *Group) (*Group, error) {

	url := g.ResourceUrl(group_create_url, nil)

	encodedRequest, err := json.Marshal(group)
	if err != nil {
		return nil, err
	}
	var result *Group

	contents, err := g.buildAndExecRequest("POST", url, encodedRequest)
	if err == nil {
		err = json.Unmarshal(contents, &result)
	}

	return result, err
}
