package node

import (
	"slices"
	"sort"
	"strings"

	"github.com/warewulf/warewulf/internal/pkg/hostlist"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

// AllGroup is the reserved built-in group name that expands to every node
// defined in the configuration, regardless of any user-defined groups.
const AllGroup = "all"

// GroupMembers returns the sorted, deduplicated list of node IDs that belong
// to the named nodegroup. Membership is the union of:
//
//   - node IDs listed in the top-level `nodegroups:` stanza (with hostlist
//     ranges expanded), and
//   - nodes whose merged per-node `nodegroups:` field contains the name.
//
// The reserved name "all" returns every defined node ID and ignores both
// the top-level stanza and per-node declarations.
//
// This method also satisfies hostlist.GroupResolver, so hostlist.Expand can
// transparently resolve "@nodegroup" tokens once node.New has registered the
// configuration via hostlist.SetGroupResolver.
func (config *NodesYaml) GroupMembers(name string) []string {
	if name == AllGroup {
		all := config.ListAllNodes()
		// Honor per-node opt-out: a node carrying literal "~all" in its
		// `nodegroups:` field (or inherited from a profile) is excluded from
		// @all expansion. We can't read the merged Node.NodeGroups for this
		// because cleanList strips standalone "~"-prefixed entries; consult
		// the literals instead.
		filtered := make([]string, 0, len(all))
		for _, id := range all {
			if config.hasLiteralNodegroup(id, "~"+AllGroup) {
				continue
			}
			filtered = append(filtered, id)
		}
		return filtered
	}

	members := make(map[string]struct{})

	for _, entry := range hostlist.Expand(config.NodeGroups[name]) {
		if _, ok := config.Nodes[entry]; ok {
			members[entry] = struct{}{}
		} else {
			wwlog.Warn("nodegroup %q references unknown node: %s", name, entry)
		}
	}

	for id := range config.Nodes {
		merged, err := config.GetNode(id)
		if err != nil {
			continue
		}
		if slices.Contains(merged.NodeGroups, name) {
			members[id] = struct{}{}
		}
	}

	if len(members) == 0 {
		// Only warn if neither source mentioned the nodegroup; otherwise the
		// caller asked for a defined-but-empty nodegroup, which is fine.
		if _, defined := config.NodeGroups[name]; !defined {
			wwlog.Warn("unknown nodegroup: %s", name)
		}
	}

	result := make([]string, 0, len(members))
	for id := range members {
		result = append(result, id)
	}
	sort.Strings(result)
	return result
}

// hasLiteralNodegroup reports whether the given token appears verbatim in
// the literal `nodegroups:` field of the named node or of any profile the
// node inherits (recursively). Unlike merged Node.NodeGroups, this bypasses
// the cleanList step that strips standalone `~`-prefixed entries, so it can
// be used to detect opt-out markers like `~all` that would otherwise
// evaporate during merge.
func (config *NodesYaml) hasLiteralNodegroup(nodeID, token string) bool {
	if n, ok := config.Nodes[nodeID]; ok {
		if slices.Contains(n.NodeGroups, token) {
			return true
		}
	}
	for _, profileID := range config.getNodeProfiles(nodeID) {
		if strings.HasPrefix(profileID, "~") {
			continue
		}
		if profile, err := config.GetProfile(profileID); err == nil {
			if slices.Contains(profile.NodeGroups, token) {
				return true
			}
		}
	}
	return false
}

// ListNodesUsingNodegroup is the merged-node counterpart of GroupMembers,
// mirroring ListNodesUsingProfile / ListNodesUsingOverlay. It returns the
// full Node objects (with profiles merged in) for every member of the named
// nodegroup.
func (config *NodesYaml) ListNodesUsingNodegroup(name string) ([]Node, error) {
	members := config.GroupMembers(name)
	if len(members) == 0 {
		return nil, nil
	}
	return config.FindAllNodes(members...)
}

// ListAllNodegroups returns a sorted, deduplicated list of every nodegroup
// name referenced anywhere in the configuration: keys of the top-level
// `nodegroups:` stanza plus any name appearing in a node's or profile's
// per-entity `nodegroups:` field.
//
// The reserved "all" nodegroup is not included unless the user has defined
// it explicitly.
func (config *NodesYaml) ListAllNodegroups() []string {
	seen := make(map[string]struct{})
	for name := range config.NodeGroups {
		seen[name] = struct{}{}
	}
	for _, node := range config.Nodes {
		for _, g := range node.NodeGroups {
			if strings.HasPrefix(g, "~") {
				continue
			}
			seen[g] = struct{}{}
		}
	}
	for _, profile := range config.NodeProfiles {
		for _, g := range profile.NodeGroups {
			if strings.HasPrefix(g, "~") {
				continue
			}
			seen[g] = struct{}{}
		}
	}
	out := make([]string, 0, len(seen))
	for g := range seen {
		out = append(out, g)
	}
	sort.Strings(out)
	return out
}
