// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package next

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"strings"

	"devnw.dev/canary/pkg/canaryscan"
	"devnw.dev/canary/pkg/evidence"
	"devnw.dev/canary/pkg/external"
	"devnw.dev/canary/pkg/sources"
	"devnw.dev/canary/pkg/storage"
)

// declaredFn answers "what does this project declare for reqID?" as the
// feature/aspect keys that would have to be proven for it to be complete. An
// id with no local declarations returns an empty slice, which is what routes
// it to the external/peer resolver instead.
//
// The two selection sources supply it differently -- the index by query, a
// filesystem scan from the report it just produced -- and that is the ONLY
// difference between them: both then run the same completion rule below.
type declaredFn func(reqID string) ([]evidence.FeatureKey, error)

// depGate decides whether a candidate's dependencies block it. There is one
// definition of "complete" in `canary next`, and it lives here:
//
//   - A dependency this project declares is complete iff pkg/evidence.Complete
//     passes over its declared features with the local evidence store at the
//     current commit and project. A declaration (STATUS=TESTED) is a claim,
//     not proof, and `next` no longer accepts one as proof.
//   - A dependency this project does not declare is resolved as external:
//     satisfied clears it, unsatisfied blocks it, and unknown blocks it too
//     unless the caller has explicitly accepted that risk with
//     --allow-unknown-external.
//   - A dependency that is neither declared here nor external at all (an
//     unconfigured prefix, or this project's own flatfile series with no
//     token) blocks: it names something that does not exist.
type depGate struct {
	root string
	// evidenceProjectID scopes evidence lookups. It is the configured
	// project.key unless --project overrode it; the index scope (which may be
	// empty, meaning "every project") is a different question and lives on the
	// query, not here.
	evidenceProjectID string
	// commit is the current HEAD, or "" when it cannot be read. Empty is not a
	// wildcard: no record can match it, so every dependency fails closed.
	commit               string
	recs                 []evidence.Record
	reg                  *sources.Registry
	allowUnknownExternal bool
	// warned dedups the per-dependency stderr note to once per run.
	warned map[string]bool
	stderr io.Writer
	// blockedCount counts the candidates this gate turned away, so an empty
	// answer can say whether the tree is finished or merely stuck. Reporting
	// "all requirements completed" over a pile of blocked work is the same
	// class of lie as reporting it over an index that was never built.
	blockedCount int
}

// blocked reports whether any id in a comma-separated DEPENDS_ON list is
// incomplete. An error is returned only for questions canary refuses to
// answer (storage.ErrProjectRequired), never for a dependency that is merely
// missing.
func (g *depGate) blocked(dependsOn string, declared declaredFn) (blocked bool, err error) {
	if strings.TrimSpace(dependsOn) == "" {
		return false, nil
	}
	defer func() {
		if blocked {
			g.blockedCount++
		}
	}()
	for _, dep := range strings.Split(dependsOn, ",") {
		dep = strings.TrimSpace(dep)
		if dep == "" {
			continue
		}
		keys, err := declared(dep)
		if err != nil {
			return false, err
		}
		if len(keys) == 0 {
			deny, derr := g.externalBlocks(dep)
			if derr != nil {
				return false, derr
			}
			if deny {
				return true, nil
			}
			continue
		}
		if !g.complete(dep, keys) {
			return true, nil
		}
	}
	return false, nil
}

// complete is the local half of the one completion definition: every declared
// feature of reqID must have a PASS evidence record for this project at this
// commit.
func (g *depGate) complete(reqID string, keys []evidence.FeatureKey) bool {
	return evidence.CompleteReq(reqID, keys, g.recs, g.evidenceProjectID, g.commit)
}

// externalBlocks resolves a dependency with no local declarations against the
// external/peer sources and reports whether it blocks selection.
//
// Unknown blocks. It used to pass -- "degradation is sacred" -- but the thing
// being degraded there was the answer to "may this work start?", and handing
// an agent a requirement whose prerequisite might not exist is not a
// degradation, it is a wrong answer. The risk is still available on request:
// --allow-unknown-external.
func (g *depGate) externalBlocks(dep string) (bool, error) {
	res := external.Resolve(dep, g.reg, g.root)
	if !res.IsExternal() {
		return true, nil // names nothing this project or any source knows
	}
	switch res.State {
	case external.StateSatisfied:
		return false, nil
	case external.StateUnsatisfied:
		return true, nil
	default:
		g.note(dep, res)
		return !g.allowUnknownExternal, nil
	}
}

// note prints the one-line explanation for an unresolvable dependency, once
// per id per run.
func (g *depGate) note(dep string, res external.Resolution) {
	if g.warned == nil || g.warned[dep] {
		return
	}
	g.warned[dep] = true
	w := g.stderr
	if w == nil {
		return
	}
	verb := "blocks selection"
	if g.allowUnknownExternal {
		verb = "is unresolved (allowed)"
	}
	fmt.Fprintf(w, "note: external dependency %s %s: %s\n", dep, verb, res.ShortDetail())
}

// dbDeclared answers declaredFn from the token index. A contract refusal --
// the same id under two projects with no --project to disambiguate -- is a
// question canary cannot answer, not a dependency that happens to be missing,
// so it is propagated rather than swallowed into "no declarations".
func dbDeclared(db *storage.DB, projectID string) declaredFn {
	return func(reqID string) ([]evidence.FeatureKey, error) {
		toks, err := db.GetTokensByReqID(projectID, reqID)
		if err != nil {
			if errors.Is(err, storage.ErrProjectRequired) {
				return nil, err
			}
			return nil, nil
		}
		keys := make([]evidence.FeatureKey, 0, len(toks))
		seen := make(map[evidence.FeatureKey]struct{}, len(toks))
		for _, tok := range toks {
			key := evidence.FeatureKey{Feature: tok.Feature, Aspect: tok.Aspect}
			if _, dup := seen[key]; dup {
				continue
			}
			seen[key] = struct{}{}
			keys = append(keys, key)
		}
		return keys, nil
	}
}

// scanDeclared answers declaredFn from a filesystem scan's report, built once
// so a candidate list of any length costs one pass over the requirements.
func scanDeclared(rep canaryscan.Report) declaredFn {
	declared := make(map[string][]evidence.FeatureKey, len(rep.Requirements))
	for _, r := range rep.Requirements {
		declared[r.ID] = canaryscan.DeclaredFeatures(rep, r.ID)
	}
	return func(reqID string) ([]evidence.FeatureKey, error) {
		return declared[reqID], nil
	}
}

// loadEvidenceRecords reads root's evidence store. A missing store is not an
// error -- nothing has been proven yet, so nothing is complete. A malformed
// one IS an error: treating unparseable evidence as absent would hide
// tampering behind a "dependency incomplete" that reads like ordinary
// progress.
func loadEvidenceRecords(root string) ([]evidence.Record, error) {
	path := filepath.Join(root, filepath.FromSlash(canaryscan.EvidenceFile))
	f, err := evidence.Load(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	return f.Records, nil
}
