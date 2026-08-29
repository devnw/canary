// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package ticket

import (
	"reflect"
	"testing"

	"devnw.dev/canary/pkg/sources"
	"devnw.dev/canary/pkg/storage"
)

// testRegistry mirrors pkg/sources's own test fixture: one flatfile source
// (CBIN) plus one non-flatfile (jira, PLAT) so ComputePlan sees both the
// create_issue+remap path and the transition path.
func testRegistry(t *testing.T, statusMap map[string]string) *sources.Registry {
	t.Helper()
	r, err := sources.NewRegistry([]sources.Source{
		{Name: "core", Type: "flatfile", Key: "CBIN"},
		{Name: "platform", Type: "jira", Key: "PLAT", StatusMap: statusMap},
	})
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}
	return r
}

func flatfileOnlyRegistry(t *testing.T) *sources.Registry {
	t.Helper()
	r, err := sources.NewRegistry([]sources.Source{
		{Name: "core", Type: "flatfile", Key: "CBIN"},
	})
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}
	return r
}

func TestCANARY_CBIN_306_ComputePlan_FlatfileCreateAndRemapPairing(t *testing.T) {
	reg := testRegistry(t, nil)
	tokens := []*storage.Token{
		{ReqID: "CBIN-105", Feature: "Scanner", Aspect: "Engine", Status: "TESTED", FilePath: "scan.go"},
		{ReqID: "CBIN-105", Feature: "ScannerCLI", Aspect: "CLI", Status: "IMPL", FilePath: "cli.go"},
	}

	actions, err := ComputePlan(tokens, reg, nil)
	if err != nil {
		t.Fatalf("ComputePlan: %v", err)
	}
	if len(actions) != 2 {
		t.Fatalf("len(actions) = %d, want 2 (create_issue + remap): %+v", len(actions), actions)
	}
	create := actions[0]
	if create.Type != "create_issue" || create.ReqID != "CBIN-105" {
		t.Errorf("actions[0] = %+v, want create_issue for CBIN-105", create)
	}
	if create.Summary == "" {
		t.Error("create_issue Summary must not be empty")
	}
	if create.Description == "" {
		t.Error("create_issue Description must not be empty")
	}
	remap := actions[1]
	if remap.Type != "remap" || remap.ReqID != "CBIN-105" || remap.Issue != "" {
		t.Errorf("actions[1] = %+v, want paired remap with empty Issue placeholder", remap)
	}
}

func TestCANARY_ENG_3958_ComputePlan_CreateIssueStampsDestinationProject(t *testing.T) {
	r, err := sources.NewRegistry([]sources.Source{
		{Name: "core", Type: "flatfile", Key: "CBIN"},
		{Name: "platform", Type: "jira", Key: "PLAT", Project: "PLATPROJ"},
	})
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}
	tokens := []*storage.Token{
		{ReqID: "CBIN-105", Feature: "Scanner", Aspect: "Engine", Status: "TESTED", FilePath: "scan.go"},
	}
	actions, err := ComputePlan(tokens, r, nil)
	if err != nil {
		t.Fatalf("ComputePlan: %v", err)
	}
	if len(actions) != 2 || actions[0].Type != "create_issue" {
		t.Fatalf("actions = %+v, want create_issue+remap", actions)
	}
	if actions[0].Project != "PLATPROJ" {
		t.Errorf("create_issue.Project = %q, want PLATPROJ (destination source)", actions[0].Project)
	}
	if actions[1].Project != "" {
		t.Errorf("remap.Project = %q, want empty", actions[1].Project)
	}
}

func TestCANARY_ENG_3958_ComputePlan_CreateIssueProjectEmptyWhenDestinationUnset(t *testing.T) {
	// No source in the registry carries a Project — create_issue's Project
	// must be empty, not error or panic.
	reg := testRegistry(t, nil)
	tokens := []*storage.Token{
		{ReqID: "CBIN-105", Feature: "Scanner", Aspect: "Engine", Status: "TESTED", FilePath: "scan.go"},
	}
	actions, err := ComputePlan(tokens, reg, nil)
	if err != nil {
		t.Fatalf("ComputePlan: %v", err)
	}
	if len(actions) != 2 || actions[0].Type != "create_issue" || actions[0].Project != "" {
		t.Fatalf("actions = %+v, want create_issue with empty Project (no destination project configured)", actions)
	}
}

func TestCANARY_CBIN_306_ComputePlan_FlatfileNoNonFlatfileSource_NoAction(t *testing.T) {
	// With only a flatfile source configured, there's nothing to promote a
	// requirement to — no create_issue/remap pair should be proposed.
	reg := flatfileOnlyRegistry(t)
	tokens := []*storage.Token{
		{ReqID: "CBIN-105", Feature: "Scanner", Aspect: "Engine", Status: "STUB", FilePath: "scan.go"},
	}

	actions, err := ComputePlan(tokens, reg, nil)
	if err != nil {
		t.Fatalf("ComputePlan: %v", err)
	}
	if len(actions) != 0 {
		t.Errorf("actions = %+v, want none (no non-flatfile source configured)", actions)
	}
}

func TestCANARY_CBIN_306_ComputePlan_JiraStatusMismatch_Transition(t *testing.T) {
	reg := testRegistry(t, nil)
	tokens := []*storage.Token{
		{ReqID: "PLAT-42", Feature: "Sync", Aspect: "Engine", Status: "IMPL", FilePath: "sync.go"},
	}
	remoteStatus := map[string]string{"PLAT-42": "To Do"} // canary says IMPL -> "In Progress"; remote says "To Do"

	actions, err := ComputePlan(tokens, reg, remoteStatus)
	if err != nil {
		t.Fatalf("ComputePlan: %v", err)
	}
	if len(actions) != 1 {
		t.Fatalf("len(actions) = %d, want 1: %+v", len(actions), actions)
	}
	a := actions[0]
	if a.Type != "transition" || a.Issue != "PLAT-42" || a.To != "In Progress" {
		t.Errorf("action = %+v, want transition PLAT-42 -> In Progress", a)
	}
}

func TestCANARY_CBIN_306_ComputePlan_MatchingStatus_NoAction(t *testing.T) {
	reg := testRegistry(t, nil)
	tokens := []*storage.Token{
		{ReqID: "PLAT-42", Feature: "Sync", Aspect: "Engine", Status: "TESTED", FilePath: "sync.go"},
	}
	remoteStatus := map[string]string{"PLAT-42": "Done"} // TESTED -> "Done" default mapping matches remote

	actions, err := ComputePlan(tokens, reg, remoteStatus)
	if err != nil {
		t.Fatalf("ComputePlan: %v", err)
	}
	if len(actions) != 0 {
		t.Errorf("actions = %+v, want none (status already matches)", actions)
	}
}

func TestCANARY_CBIN_306_ComputePlan_EmptyRemoteStatus_AllTransitionsProposed(t *testing.T) {
	reg := testRegistry(t, nil)
	tokens := []*storage.Token{
		{ReqID: "PLAT-1", Feature: "A", Aspect: "Engine", Status: "STUB", FilePath: "a.go"},
		{ReqID: "PLAT-2", Feature: "B", Aspect: "Engine", Status: "BENCHED", FilePath: "b.go"},
	}

	actions, err := ComputePlan(tokens, reg, map[string]string{})
	if err != nil {
		t.Fatalf("ComputePlan: %v", err)
	}
	if len(actions) != 2 {
		t.Fatalf("len(actions) = %d, want 2 (empty remoteStatus proposes every transition): %+v", len(actions), actions)
	}
	for _, a := range actions {
		if a.Type != "transition" {
			t.Errorf("action %+v, want type transition", a)
		}
	}
}

func TestCANARY_CBIN_306_ComputePlan_StatusMapOverrideHonored(t *testing.T) {
	reg := testRegistry(t, map[string]string{"STUB": "Backlog"})
	tokens := []*storage.Token{
		{ReqID: "PLAT-7", Feature: "X", Aspect: "Engine", Status: "STUB", FilePath: "x.go"},
	}
	// Remote already at the DEFAULT mapping ("To Do") — since the source
	// overrides STUB -> "Backlog", this must still propose a transition.
	remoteStatus := map[string]string{"PLAT-7": "To Do"}

	actions, err := ComputePlan(tokens, reg, remoteStatus)
	if err != nil {
		t.Fatalf("ComputePlan: %v", err)
	}
	if len(actions) != 1 || actions[0].To != "Backlog" {
		t.Fatalf("actions = %+v, want one transition to Backlog (StatusMap override)", actions)
	}

	// Remote already matching the OVERRIDE must produce no action.
	remoteStatus2 := map[string]string{"PLAT-7": "Backlog"}
	actions2, err := ComputePlan(tokens, reg, remoteStatus2)
	if err != nil {
		t.Fatalf("ComputePlan: %v", err)
	}
	if len(actions2) != 0 {
		t.Errorf("actions = %+v, want none (remote already matches override)", actions2)
	}
}

func TestCANARY_CBIN_306_ComputePlan_RollupIsWorstOfTokens(t *testing.T) {
	reg := testRegistry(t, nil)
	// Two tokens for the same requirement: BENCHED and STUB. Worst = STUB.
	tokens := []*storage.Token{
		{ReqID: "PLAT-9", Feature: "X", Aspect: "Bench", Status: "BENCHED", FilePath: "x_test.go"},
		{ReqID: "PLAT-9", Feature: "X", Aspect: "Engine", Status: "STUB", FilePath: "x.go"},
	}
	actions, err := ComputePlan(tokens, reg, nil)
	if err != nil {
		t.Fatalf("ComputePlan: %v", err)
	}
	if len(actions) != 1 || actions[0].To != "To Do" {
		t.Fatalf("actions = %+v, want one transition to To Do (worst-of-tokens = STUB)", actions)
	}
}

func TestCANARY_CBIN_306_ComputePlan_UnresolvedPrefixSkipped(t *testing.T) {
	reg := testRegistry(t, nil)
	tokens := []*storage.Token{
		{ReqID: "OTHER-1", Feature: "X", Aspect: "Engine", Status: "STUB", FilePath: "x.go"},
	}
	actions, err := ComputePlan(tokens, reg, nil)
	if err != nil {
		t.Fatalf("ComputePlan: %v", err)
	}
	if len(actions) != 0 {
		t.Errorf("actions = %+v, want none (unresolved prefix)", actions)
	}
}

func TestCANARY_CBIN_306_ComputePlan_DeterministicOrdering(t *testing.T) {
	reg := testRegistry(t, nil)
	tokens := []*storage.Token{
		{ReqID: "PLAT-2", Feature: "B", Aspect: "Engine", Status: "STUB", FilePath: "b.go"},
		{ReqID: "CBIN-105", Feature: "A", Aspect: "Engine", Status: "STUB", FilePath: "a.go"},
		{ReqID: "PLAT-1", Feature: "C", Aspect: "Engine", Status: "STUB", FilePath: "c.go"},
	}

	actions1, err := ComputePlan(tokens, reg, nil)
	if err != nil {
		t.Fatalf("ComputePlan: %v", err)
	}
	actions2, err := ComputePlan(tokens, reg, nil)
	if err != nil {
		t.Fatalf("ComputePlan: %v", err)
	}
	if !reflect.DeepEqual(actions1, actions2) {
		t.Fatalf("ComputePlan is not deterministic:\n%+v\nvs\n%+v", actions1, actions2)
	}
	wantOrder := []string{"CBIN-105", "CBIN-105", "PLAT-1", "PLAT-2"} // create+remap for CBIN-105 first (sorted req id)
	if len(actions1) != len(wantOrder) {
		t.Fatalf("len(actions) = %d, want %d: %+v", len(actions1), len(wantOrder), actions1)
	}
	for i, want := range wantOrder {
		if actions1[i].ReqID != want {
			t.Errorf("actions[%d].ReqID = %q, want %q", i, actions1[i].ReqID, want)
		}
	}
}

func TestCANARY_CBIN_306_RollupStatus(t *testing.T) {
	cases := []struct {
		name   string
		tokens []*storage.Token
		want   string
	}{
		{"single stub", []*storage.Token{{Status: "STUB"}}, "STUB"},
		{"worst wins", []*storage.Token{{Status: "BENCHED"}, {Status: "IMPL"}}, "IMPL"},
		{"all tested", []*storage.Token{{Status: "TESTED"}, {Status: "TESTED"}}, "TESTED"},
		{"empty", nil, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := RollupStatus(c.tokens); got != c.want {
				t.Errorf("RollupStatus() = %q, want %q", got, c.want)
			}
		})
	}
}
