package gogitlab

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGroupSearch(t *testing.T) {
	ts, gitlab := Stub("stubs/groups/show.json")
	groups, err := gitlab.GroupSearch("ws")

	assert.Equal(t, err, nil)
	assert.Equal(t, len(groups), 1)
	defer ts.Close()
}

func TestCreateGroup(t *testing.T) {
	ts, gitlab := Stub("stubs/groups/show.json")
	group := Group{
		Name:       "ws8000",
		Path:       "ws8000",
		ParentId:   35,
		Visibility: "internal",
	}

	_, err := gitlab.CreateGroup(&group)
	assert.Equal(t, err, nil)
	defer ts.Close()
}
