// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// Package ticket computes and applies ticket-source synchronization plans:
// comparing a requirement's rollup CANARY status against its owning
// non-flatfile source's remote status (proposing "transition" actions), and
// codifying flatfile-to-ticket promotion as paired "create_issue" + "remap"
// actions. Plan computation (ComputePlan) is pure and side-effect free;
// JiraClient (jira.go) is the only piece that talks to the network, and only
// when the CLI layer (pkg/cmds/ticket) is told to --apply.
// CANARY: REQ=CBIN-306; FEATURE="TicketSync"; ASPECT=Engine; STATUS=TESTED; TEST=TestCANARY_CBIN_306_ComputePlan_FlatfileCreateAndRemapPairing,TestCANARY_CBIN_306_ComputePlan_FlatfileNoNonFlatfileSource_NoAction,TestCANARY_CBIN_306_ComputePlan_JiraStatusMismatch_Transition,TestCANARY_CBIN_306_ComputePlan_MatchingStatus_NoAction,TestCANARY_CBIN_306_ComputePlan_EmptyRemoteStatus_AllTransitionsProposed,TestCANARY_CBIN_306_ComputePlan_StatusMapOverrideHonored,TestCANARY_CBIN_306_ComputePlan_RollupIsWorstOfTokens,TestCANARY_CBIN_306_ComputePlan_UnresolvedPrefixSkipped,TestCANARY_CBIN_306_ComputePlan_DeterministicOrdering,TestCANARY_CBIN_306_RollupStatus; UPDATED=2026-08-29
package ticket

import (
	"fmt"
	"sort"
	"strings"

	"devnw.dev/canary/pkg/sources"
	"devnw.dev/canary/pkg/storage"
)

// Action is one proposed (or, once applied, completed) ticket-sync
// operation.
type Action struct {
	Type        string `json:"type"`              // create_issue | transition | remap
	ReqID       string `json:"req_id,omitempty"`  // the CANARY requirement ID this action concerns
	Issue       string `json:"issue,omitempty"`   // remote issue key; "" on a pending remap until applied
	To          string `json:"to,omitempty"`      // JIRA status name (transition target)
	Summary     string `json:"summary,omitempty"` // for create_issue
	Description string `json:"description,omitempty"`
	Source      string `json:"source"` // owning source's Name
}

// statusRank orders CANARY statuses from least to most advanced.
var statusRank = map[string]int{
	"STUB":    0,
	"IMPL":    1,
	"TESTED":  2,
	"BENCHED": 3,
}

// DefaultStatusMap is the CANARY-status -> remote-status-name mapping used
// when a source's StatusMap doesn't override a given status.
var DefaultStatusMap = map[string]string{
	"STUB":    "To Do",
	"IMPL":    "In Progress",
	"TESTED":  "Done",
	"BENCHED": "Done",
}

// RollupStatus returns the worst (least advanced) status among tokens per
// STUB < IMPL < TESTED < BENCHED. Tokens with an unrecognized status are
// ignored. An empty slice, or a slice with no recognized statuses, returns
// "".
func RollupStatus(tokens []*storage.Token) string {
	worst := ""
	worstRank := -1
	for _, t := range tokens {
		if t == nil {
			continue
		}
		r, ok := statusRank[t.Status]
		if !ok {
			continue
		}
		if worstRank == -1 || r < worstRank {
			worstRank = r
			worst = t.Status
		}
	}
	return worst
}

// mappedStatus resolves a CANARY status through statusMap, falling back to
// DefaultStatusMap when statusMap doesn't override it.
func mappedStatus(status string, statusMap map[string]string) string {
	if v, ok := statusMap[status]; ok {
		return v
	}
	return DefaultStatusMap[status]
}

// descLimit bounds create_issue's Description to a handful of lines — small
// by default, per project convention.
const descLimit = 10

// buildDescription renders a bounded feature/aspect/status/file summary for
// a create_issue action's Description field.
func buildDescription(tokens []*storage.Token) string {
	sorted := make([]*storage.Token, len(tokens))
	copy(sorted, tokens)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Aspect != sorted[j].Aspect {
			return sorted[i].Aspect < sorted[j].Aspect
		}
		return sorted[i].FilePath < sorted[j].FilePath
	})

	var b strings.Builder
	shown := sorted
	if len(shown) > descLimit {
		shown = shown[:descLimit]
	}
	for _, t := range shown {
		fmt.Fprintf(&b, "%s/%s: %s (%s)\n", t.Feature, t.Aspect, t.Status, t.FilePath)
	}
	if len(sorted) > len(shown) {
		fmt.Fprintf(&b, "… +%d more\n", len(sorted)-len(shown))
	}
	return strings.TrimRight(b.String(), "\n")
}

// primaryFeature picks a deterministic representative Feature name for a
// requirement's tokens (its lowest-Aspect token, alphabetically) — used in
// create_issue's Summary.
func primaryFeature(tokens []*storage.Token) string {
	best := tokens[0]
	for _, t := range tokens[1:] {
		if t.Aspect < best.Aspect {
			best = t
		}
	}
	return best.Feature
}

// ComputePlan computes the deterministic, pure set of proposed sync actions.
//
// tokens are grouped by ReqID; reg resolves each requirement's owning
// source; remoteStatus maps a non-flatfile issue key to its current remote
// status name (nil or empty means "unknown" — every eligible transition is
// then proposed, since no remote state contradicts it).
//
// Rules:
//   - A requirement owned by a non-flatfile source gets a "transition"
//     action when its rollup status (RollupStatus: worst of its tokens,
//     STUB<IMPL<TESTED<BENCHED), mapped through the source's StatusMap
//     (falling back to DefaultStatusMap), differs from remoteStatus[reqID].
//   - A requirement owned by a flatfile source gets a paired "create_issue"
//   - "remap" action, but only when the registry configures at least one
//     non-flatfile source — otherwise there's nothing to promote it to. The
//     create_issue's Summary is "<ReqID>: <primary feature>"; Description is
//     a bounded feature/aspect/status/file list. The paired remap's Issue is
//     left "" — the apply step fills in the created key.
//   - A requirement whose prefix resolves to no configured source is
//     skipped.
//
// Ordering is deterministic: requirements are visited in sorted ReqID order;
// a flatfile requirement's create_issue action is immediately followed by
// its paired remap action.
func ComputePlan(tokens []*storage.Token, reg *sources.Registry, remoteStatus map[string]string) ([]Action, error) {
	if reg == nil {
		return nil, fmt.Errorf("ticket: registry is required")
	}

	byReq := map[string][]*storage.Token{}
	for _, t := range tokens {
		if t == nil || t.ReqID == "" {
			continue
		}
		byReq[t.ReqID] = append(byReq[t.ReqID], t)
	}
	reqIDs := make([]string, 0, len(byReq))
	for id := range byReq {
		reqIDs = append(reqIDs, id)
	}
	sort.Strings(reqIDs)

	hasNonFlatfile := false
	for _, s := range reg.Sources() {
		if s.Type != "flatfile" {
			hasNonFlatfile = true
			break
		}
	}

	var actions []Action
	for _, reqID := range reqIDs {
		toks := byReq[reqID]
		src, ok := reg.Resolve(reqID)
		if !ok {
			continue
		}

		if src.Type == "flatfile" {
			if !hasNonFlatfile {
				continue
			}
			actions = append(actions, Action{
				Type:        "create_issue",
				ReqID:       reqID,
				Summary:     fmt.Sprintf("%s: %s", reqID, primaryFeature(toks)),
				Description: buildDescription(toks),
				Source:      src.Name,
			})
			actions = append(actions, Action{
				Type:   "remap",
				ReqID:  reqID,
				Source: src.Name,
			})
			continue
		}

		rollup := RollupStatus(toks)
		if rollup == "" {
			continue
		}
		want := mappedStatus(rollup, src.StatusMap)
		if want == "" || remoteStatus[reqID] == want {
			continue
		}
		actions = append(actions, Action{
			Type:   "transition",
			ReqID:  reqID,
			Issue:  reqID,
			To:     want,
			Source: src.Name,
		})
	}
	return actions, nil
}
