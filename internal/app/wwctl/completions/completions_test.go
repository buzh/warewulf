package completions

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/warewulf/warewulf/internal/pkg/testenv"
)

func Test_Nodes_IncludesAtNodegroupCandidates(t *testing.T) {
	env := testenv.New(t)
	defer env.RemoveAll()
	env.WriteFile("etc/warewulf/nodes.conf", `
nodeprofiles:
  gpu:
    nodegroups:
      - gpu-nodes
nodes:
  n01:
    profiles:
      - gpu
  n02:
    nodegroups:
      - admin
  n03: {}
nodegroups:
  rack1:
    - n03
`)

	got, _ := Nodes(nil, nil, "")
	assert.Contains(t, got, "n01", "node candidates should still include plain names")
	assert.Contains(t, got, "n02")
	assert.Contains(t, got, "n03")
	assert.Contains(t, got, "@rack1", "top-level stanza names should appear with @ prefix")
	assert.Contains(t, got, "@admin", "per-node nodegroup names should appear with @ prefix")
	assert.Contains(t, got, "@gpu-nodes", "profile-inherited nodegroup names should appear with @ prefix")
	assert.Contains(t, got, "@all", "@all should always be offered")
}

func Test_Nodegroups_PlainNamesPlusAll(t *testing.T) {
	env := testenv.New(t)
	defer env.RemoveAll()
	env.WriteFile("etc/warewulf/nodes.conf", `
nodes:
  n01:
    nodegroups:
      - admin
nodegroups:
  rack1:
    - n01
`)

	got, _ := Nodegroups(nil, nil, "")
	assert.Contains(t, got, "rack1")
	assert.Contains(t, got, "admin")
	assert.Contains(t, got, "all", "@all should always be offered as a completion candidate")
	for _, c := range got {
		assert.NotContains(t, c, "@", "Nodegroups completions are plain names; the leading @ is added by the caller")
	}
}

func Test_Nodegroups_AllNotDuplicatedWhenUserDefined(t *testing.T) {
	env := testenv.New(t)
	defer env.RemoveAll()
	env.WriteFile("etc/warewulf/nodes.conf", `
nodes:
  n01: {}
nodegroups:
  all:
    - n01
`)

	got, _ := Nodegroups(nil, nil, "")
	count := 0
	for _, c := range got {
		if c == "all" {
			count++
		}
	}
	assert.Equal(t, 1, count, "user-defined `all` group must not produce a duplicate entry")
}
