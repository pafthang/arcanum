package mini

import "testing"

func TestPathTreeSharedParamNames(t *testing.T) {
	tree := NewPathTree[string]()
	p1, err := ParsePathPattern("/api/spaces/{spaceId}/projects")
	if err != nil {
		t.Fatal(err)
	}
	p2, err := ParsePathPattern("/api/spaces/{id}/members")
	if err != nil {
		t.Fatal(err)
	}
	tree.Add(p1, "projects")
	tree.Add(p2, "members")

	v, params, ok := tree.Match("/api/spaces/abc/members")
	if !ok || v != "members" {
		t.Fatalf("members match: ok=%v v=%v", ok, v)
	}
	// Must expose the param name from the winning pattern, not the first shared edge.
	if params["id"] != "abc" {
		t.Fatalf("expected params[id]=abc, got %v", params)
	}

	v, params, ok = tree.Match("/api/spaces/abc/projects")
	if !ok || v != "projects" {
		t.Fatalf("projects match: ok=%v v=%v", ok, v)
	}
	if params["spaceId"] != "abc" {
		t.Fatalf("expected params[spaceId]=abc, got %v", params)
	}
}
