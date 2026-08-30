// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// Package ticket wires pkg/ticket into `canary ticket sync`: computing a
// codified ticket-source synchronization plan from indexed tokens and the
// configured `sources:` registry, and — only when JIRA credentials are
// present and --apply is set — applying it via the JIRA REST client and
// writing a completed plan plus a remap map for `canary upgrade --map`.
// CANARY: REQ=CP-279; FEATURE="TicketSync"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_306_Sync_NoCredsPlanOnly,TestCANARY_CBIN_306_Sync_NoCredsApplyDegradesGracefully,TestCANARY_CBIN_306_Sync_ApplyWithCreds_EndToEnd,TestCANARY_CBIN_306_Sync_ApplyNoCreds_ProjectRequired,TestCANARY_CBIN_306_Sync_PartialProgress,TestCANARY_CBIN_306_PrintActions_Bounded; UPDATED=2026-08-29
// CANARY: REQ=ENG-3958; FEATURE="TicketDestination"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_ENG_3958_Sync_ApplyWithDestinationSource_NoProjectFlag,TestCANARY_ENG_3958_RemoteStatusForSources_MultiSourceMerge,TestCANARY_ENG_3958_Sync_ApplyNoProject_BlindTransitionRefused,TestCANARY_ENG_3958_Sync_ApplyWithDestinationSource_TransitionsUnaffected,TestCANARY_ENG_3958_Status_Refresh_WithCredsNoProject_PreservesExistingCache,TestCANARY_ENG_3958_Status_Refresh_ProjectWithZeroIssues_CacheSaved; UPDATED=2026-08-29
package ticket

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"devnw.dev/canary/pkg/cmds/internal/utils"
	"devnw.dev/canary/pkg/config"
	"devnw.dev/canary/pkg/external"
	"devnw.dev/canary/pkg/sources"
	"devnw.dev/canary/pkg/storage"
	"devnw.dev/canary/pkg/ticket"
)

// DefaultPlanLimit bounds the printed action table by default; agents raise
// it with --limit when they need more (the full plan is always in the JSON
// file at --plan).
const DefaultPlanLimit = 20

// CreateTicketCommand returns the `canary ticket` parent command with its
// subcommands attached (mirrors CreateDriftCommand / CreateViewCommand: a
// fresh command tree per call, so both the CLI entry point and tests get
// isolated flag state).
func CreateTicketCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ticket <subcommand>",
		Short: "Synchronize CANARY requirements with configured ticket sources (e.g. JIRA)",
		Long: `Compute (and, with credentials, apply) a plan that reconciles CANARY
token state against external ticket sources configured in
.canary/project.yaml's sources: list.

Subcommands:
  sync    Compute a ticket-sync plan; apply it against JIRA with --apply
  status  Show the cached remote-status snapshot, or refresh it (--refresh)
          without computing or applying a sync plan`,
	}
	cmd.AddCommand(createTicketSyncCommand())
	cmd.AddCommand(createTicketStatusCommand())
	return cmd
}

func createTicketSyncCommand() *cobra.Command {
	var dbPath, planPath, project, issueType string
	var apply bool
	var limit int

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Compute (and optionally apply) a ticket-sync plan from indexed tokens",
		Long: `Reads tokens from the CANARY index database and the configured sources
registry, then computes:

  transition     a non-flatfile requirement (e.g. JIRA) whose rollup status
                 no longer matches its remote issue's status.
  create_issue   a flatfile requirement, paired with a "remap" action, when
                 at least one non-flatfile source is configured — promoting
                 it out of the flatfile series.

Without JIRA credentials (JIRA_BASE_URL, JIRA_EMAIL, JIRA_API_TOKEN all set),
this command NEVER errors and NEVER touches the network: it always writes
the plan to --plan and prints a plan-only summary, even with --apply.

JIRA_BASE_URL may be omitted from the environment when the configured jira
source in .canary/project.yaml carries an "api:" field — it is used as the
BaseURL fallback. Precedence is env > source.API: an explicit JIRA_BASE_URL
always wins. JIRA_EMAIL and JIRA_API_TOKEN have no config-file fallback and
must always come from the environment.

With credentials AND --apply: the effective project for creating new
issues is --project when set, otherwise the configured destination
source's "project:" field (the source with "destination: true", or — when
none is marked — the first non-flatfile source; see "sources:" in
.canary/project.yaml). --project is only required when a create_issue
action exists and neither resolves to a project. Remote status is fetched
and merged across every configured jira-type source that has its own
"project:" set (falling back to --project for sources that don't), so a
single sync can cover multiple JIRA projects. Applies create_issue and
transition actions via the JIRA REST API, fills created keys into their
paired remap actions, and writes the completed plan to --plan plus a remap
map ({"OLD-ID":"NEW-ID"}) to <plan>.map.json — ready for 'canary upgrade
--map <plan>.map.json --write'.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTicketSync(cmd, dbPath, planPath, project, issueType, apply, limit)
		},
	}

	cmd.Flags().StringVar(&dbPath, "db", ".canary/canary.db", "path to the CANARY index database")
	cmd.Flags().StringVar(&planPath, "plan", ".canary/ticket-plan.json", "path to write the JSON sync plan")
	cmd.Flags().BoolVar(&apply, "apply", false, "apply the plan against JIRA (requires JIRA_BASE_URL/JIRA_EMAIL/JIRA_API_TOKEN)")
	cmd.Flags().StringVar(&project, "project", "", "JIRA project key; overrides the configured destination source's project (required under --apply only when neither is set and the plan would create issues)")
	cmd.Flags().StringVar(&issueType, "issue-type", "Story", "JIRA issue type for created issues")
	cmd.Flags().IntVar(&limit, "limit", DefaultPlanLimit, "max actions shown in the table (raise when you need more; the full plan is always written to --plan)")
	return cmd
}

// currentTime returns CANARY_TEST_TIMESTAMP (RFC3339) when set and valid,
// otherwise the current time in UTC — the fetched_at stamp written to the
// remote-status cache, and the reference point pkg/external compares it
// against for staleness. Mirrors pkg/external.refTime's convention.
func currentTime() time.Time {
	if ts := os.Getenv("CANARY_TEST_TIMESTAMP"); ts != "" {
		if t, err := time.Parse(time.RFC3339, ts); err == nil {
			return t.UTC()
		}
	}
	return time.Now().UTC()
}

// createTicketStatusCommand returns the `canary ticket status` subcommand:
// plain, it reports the on-disk remote-status cache's age and entry count
// without any network access; --refresh fetches current status from every
// configured jira-type source and overwrites the cache, without computing
// or applying a sync plan.
// CANARY: REQ=ENG-3959; FEATURE="ExternalResolve"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_ENG_3959_Status_Refresh_NoCreds,TestCANARY_ENG_3959_Status_Refresh_WithCreds,TestCANARY_ENG_3959_Status_Plain_ReportsCache,TestCANARY_ENG_3959_Status_Plain_NoCache; UPDATED=2026-08-29
func createTicketStatusCommand() *cobra.Command {
	var refresh bool
	var project string

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show the cached remote ticket status, or refresh it with --refresh",
		Long: `Without --refresh, reports the on-disk remote-status cache
(.canary/remote-status.json) — its age and entry count — without touching
the network. Missing cache is reported, not an error.

With --refresh, fetches current status from every configured jira-type
source (the same multi-project fetch 'ticket sync --apply' uses) and
overwrites the cache. Without JIRA credentials (JIRA_BASE_URL, JIRA_EMAIL,
JIRA_API_TOKEN all set, or a source api: field as the BaseURL fallback),
--refresh degrades gracefully: it never errors, never touches the network,
and never touches the existing cache file.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTicketStatus(cmd, refresh, project)
		},
	}

	cmd.Flags().BoolVar(&refresh, "refresh", false, "fetch current remote status from every configured jira-type source and overwrite the cache")
	cmd.Flags().StringVar(&project, "project", "", "JIRA project key fallback for sources without their own project: (same semantics as 'ticket sync --project')")
	return cmd
}

func runTicketStatus(cmd *cobra.Command, refresh bool, project string) error {
	cmd.SilenceUsage = true
	root := "."

	if !refresh {
		return reportCachedStatus(cmd, root)
	}

	reg, err := sources.LoadFromRoot(root)
	if err != nil {
		return fmt.Errorf("load .canary/project.yaml: %w", err)
	}
	creds := credsFromEnv(reg)
	if !creds.present() {
		fmt.Fprintln(cmd.OutOrStdout(), "CANARY_TICKET_STATUS cached=0 reason=no_credentials")
		return nil
	}

	client := &ticket.JiraClient{BaseURL: creds.BaseURL, Email: creds.Email, Token: creds.Token}
	statuses, fetchedProjects, err := remoteStatusForSources(client, reg, project)
	if err != nil {
		return err
	}

	// A zero-project fetch (no source resolved a project, and --project was
	// not given as a fallback) means nothing was actually fetched -- never
	// overwrite an existing cache with that non-result. This is distinct
	// from fetching a project that legitimately has zero issues: that is
	// still a real fetch (fetchedProjects > 0) and is saved below.
	if fetchedProjects == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "CANARY_TICKET_STATUS cached=0 reason=no_project")
		return nil
	}

	fetchedAt := currentTime()
	if err := external.SaveCache(root, statuses, fetchedAt); err != nil {
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "CANARY_TICKET_STATUS cached=%d fetched_at=%s\n", len(statuses), fetchedAt.Format(time.RFC3339))
	return nil
}

// reportCachedStatus prints the cache's entry count, fetched_at, and age
// without any network access. A missing cache is reported as cached=0
// reason=no_cache — never an error.
func reportCachedStatus(cmd *cobra.Command, root string) error {
	cache, err := external.LoadCache(root)
	if err != nil {
		return err
	}
	if cache == nil {
		fmt.Fprintln(cmd.OutOrStdout(), "CANARY_TICKET_STATUS cached=0 reason=no_cache")
		return nil
	}

	age := "unknown"
	if fetchedAt, perr := time.Parse(time.RFC3339, cache.FetchedAt); perr == nil {
		age = currentTime().Sub(fetchedAt).Round(time.Second).String()
	}
	fmt.Fprintf(cmd.OutOrStdout(), "CANARY_TICKET_STATUS cached=%d fetched_at=%s age=%s\n", len(cache.Statuses), cache.FetchedAt, age)
	return nil
}

// jiraCreds is the three-env-var credential bundle `canary ticket sync`
// looks for. All three must be set or the command degrades to plan-only —
// it must never error for their absence.
type jiraCreds struct {
	BaseURL, Email, Token string
}

func (c jiraCreds) present() bool {
	return c.BaseURL != "" && c.Email != "" && c.Token != ""
}

// credsFromEnv reads JIRA credentials from the environment. Email and Token
// always come from JIRA_EMAIL/JIRA_API_TOKEN — there is no config-file
// fallback for secrets. BaseURL prefers JIRA_BASE_URL when set; when it is
// unset, it falls back to the API field of the first configured jira-type
// source in reg (reg may be nil). Precedence is env > source.API, so an
// explicit JIRA_BASE_URL always wins even if a source also carries API.
func credsFromEnv(reg *sources.Registry) jiraCreds {
	c := jiraCreds{
		BaseURL: os.Getenv("JIRA_BASE_URL"),
		Email:   os.Getenv("JIRA_EMAIL"),
		Token:   os.Getenv("JIRA_API_TOKEN"),
	}
	if c.BaseURL == "" && reg != nil {
		for _, s := range reg.Sources() {
			if s.Type == "jira" && s.API != "" {
				c.BaseURL = s.API
				break
			}
		}
	}
	return c
}

func runTicketSync(cmd *cobra.Command, dbPath, planPath, project, issueType string, apply bool, limit int) error {
	cmd.SilenceUsage = true

	reg, err := sources.LoadFromRoot(".")
	if err != nil {
		return fmt.Errorf("load .canary/project.yaml: %w", err)
	}

	// Read-only: ticket sync consumes the index, it never builds one.
	db, err := utils.OpenIndexRO(cmd, dbPath)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	// Honor the project id_pattern (like list/next) so fixture/example
	// tokens in the index never leak into ticket plans.
	cfg, err := config.Load(".")
	if err != nil {
		return fmt.Errorf("load .canary/project.yaml: %w", err)
	}
	idPattern := cfg.Requirements.IDPattern
	// Deliberately unscoped: `ticket sync --project` already means the JIRA
	// project key, so there is no room here for canary's own project scope.
	// An index holding more than one canary project makes this ambiguous,
	// and ListTokens is the wrong place to notice -- ticket sync gains its
	// own scope selector when the MCP/ticket surface is reworked.
	tokens, err := db.ListTokens("", nil, idPattern, "req_asc", 0)
	if err != nil {
		return utils.GuardContract(cmd, err)
	}

	creds := credsFromEnv(reg)

	// Real apply path: credentials present AND --apply. Everything else
	// (no --apply, or --apply without credentials) is plan-only and never
	// touches the network — degradation is sacred.
	if apply && creds.present() {
		return applyAndReport(cmd, tokens, reg, creds, planPath, project, issueType, limit)
	}

	actions, err := ticket.ComputePlan(tokens, reg, nil)
	if err != nil {
		return err
	}
	if err := writePlan(planPath, actions); err != nil {
		return err
	}
	printActions(cmd, actions, limit)

	reason := "plan_only"
	if apply {
		reason = "no_credentials"
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Plan written to %s\n", planPath)
	fmt.Fprintf(cmd.OutOrStdout(), "CANARY_TICKET_PLAN actions=%d applied=0 reason=%s\n", len(actions), reason)
	return nil
}

// applyAndReport resolves the effective creation project (--project, else
// the configured destination source's project), fetches and merges remote
// status across every jira-type source, computes the plan against it,
// applies create_issue + transition actions via client, writes the
// completed plan and its remap map, and prints the CANARY_TICKET_SYNC
// summary.
func applyAndReport(cmd *cobra.Command, tokens []*storage.Token, reg *sources.Registry, creds jiraCreds, planPath, project, issueType string, limit int) error {
	// Require an effective project when the plan contains create_issue
	// actions that would need one to apply successfully. Transition
	// actions target an already-existing issue by key and never need a
	// project.
	actions, err := ticket.ComputePlan(tokens, reg, nil)
	if err != nil {
		return err
	}

	effProject := effectiveCreationProject(project, reg)
	if effProject == "" && planContainsCreate(actions) {
		return fmt.Errorf("--project is required with --apply (plan contains create_issue actions and no destination source has a configured project)")
	}

	client := &ticket.JiraClient{BaseURL: creds.BaseURL, Email: creds.Email, Token: creds.Token}

	remoteStatus, fetchedProjects, err := remoteStatusForSources(client, reg, project)
	if err != nil {
		return err
	}

	actions, err = ticket.ComputePlan(tokens, reg, remoteStatus)
	if err != nil {
		return err
	}

	// A transition action is applied blind (as "every eligible transition
	// proposed") whenever remote status is unknown for its issue -- which is
	// exactly what a zero-project fetch produces. Applying blind against
	// JIRA risks flipping an issue to a status that was already correct
	// remotely, so refuse before touching the network at all. A create_issue
	// action with a resolvable destination project is unaffected: it never
	// depends on remote status.
	if fetchedProjects == 0 && planContainsTransition(actions) {
		return fmt.Errorf("--project is required with --apply (plan contains transition actions and no remote status could be fetched)")
	}

	// Cache the fetch as soon as it succeeds — pkg/external and a plain
	// `canary ticket status` read this snapshot without ever touching the
	// network, so it must reflect the freshest fetch regardless of whether
	// applying the plan below fully succeeds. A zero-project fetch (nothing
	// resolved to fetch) must never overwrite an existing cache -- that's
	// not a fresher snapshot, it's a non-result.
	if fetchedProjects > 0 {
		if err := external.SaveCache(".", remoteStatus, currentTime()); err != nil {
			return err
		}
	}

	created, transitioned, applyErrors := applyActions(client, actions, effProject, issueType)

	if err := writePlan(planPath, actions); err != nil {
		return err
	}
	if err := writeRemapMap(planPath, actions); err != nil {
		return err
	}

	printActions(cmd, actions, limit)
	fmt.Fprintf(cmd.OutOrStdout(), "Plan written to %s (remap map: %s.map.json)\n", planPath, planPath)
	summary := fmt.Sprintf("CANARY_TICKET_SYNC created=%d transitioned=%d remap_pending=%d",
		created, transitioned, countPendingRemap(actions))
	if len(applyErrors) > 0 {
		summary += fmt.Sprintf(" errors=%d", len(applyErrors))
		fmt.Fprintln(cmd.OutOrStdout(), summary)
		for _, errMsg := range applyErrors {
			fmt.Fprintf(cmd.OutOrStderr(), "error: %s\n", errMsg)
		}
		return fmt.Errorf("apply: %d action(s) failed", len(applyErrors))
	}
	fmt.Fprintln(cmd.OutOrStdout(), summary)
	return nil
}

// effectiveCreationProject resolves the JIRA project used to create new
// issues: the --project flag when set, otherwise the configured
// destination source's Project (see sources.Registry.DestinationSource).
// Empty when neither resolves.
// CANARY: REQ=ENG-3958; FEATURE="TicketDestination"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_ENG_3958_Sync_ApplyWithDestinationSource_NoProjectFlag; UPDATED=2026-08-29
func effectiveCreationProject(flagProject string, reg *sources.Registry) string {
	if flagProject != "" {
		return flagProject
	}
	if reg == nil {
		return ""
	}
	dest, ok := reg.DestinationSource()
	if !ok {
		return ""
	}
	return dest.Project
}

// planContainsCreate reports whether the plan contains a create_issue
// action — the only action type that requires a resolvable project to
// apply. Transition actions target an already-existing issue by key and
// never need a creation project (see planContainsTransition for the
// separate zero-fetch guard that applies to them). Callers combine this
// with an empty effective project to decide the plan is unresolved.
func planContainsCreate(actions []ticket.Action) bool {
	for _, a := range actions {
		if a.Type == "create_issue" {
			return true
		}
	}
	return false
}

// planContainsTransition reports whether the plan contains a transition
// action — the action type that is applied blind (proposed unconditionally)
// when remote status is unknown for its issue. Callers combine this with a
// zero-project fetch to decide the plan is unresolved.
func planContainsTransition(actions []ticket.Action) bool {
	for _, a := range actions {
		if a.Type == "transition" {
			return true
		}
	}
	return false
}

// remoteStatusForSources fetches and merges remote status across every
// jira-type source in reg: each source's own Project is queried first;
// sources with no Project configured fall back to fallbackProject (the
// --project flag), preserving single-project configs' historical behavior.
// A given project is only fetched once even when shared by multiple
// sources. Sources (and the fallback) that resolve to no project at all are
// skipped — they contribute nothing to merge.
//
// Merge semantics (last-write-wins): when the same issue key appears in
// results from multiple sources, the last source in registry order's status
// value overwrites earlier ones. This allows a canonical source for a key to
// appear later in the registry and take precedence.
//
// The returned fetchedProjects count is the number of distinct projects
// actually queried — zero means nothing resolved to a project at all (no
// source's own Project, and no usable fallbackProject), as opposed to a
// project that was queried and legitimately returned zero issues. Callers
// use this to distinguish "nothing to fetch" from "fetched, found nothing"
// — see applyAndReport and runTicketStatus.
// CANARY: REQ=ENG-3958; FEATURE="TicketDestination"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_ENG_3958_RemoteStatusForSources_MultiSourceMerge,TestCANARY_ENG_3958_RemoteStatusForSources_SharedProjectSingleFetch; UPDATED=2026-08-29
func remoteStatusForSources(client *ticket.JiraClient, reg *sources.Registry, fallbackProject string) (merged map[string]string, fetchedProjects int, err error) {
	if reg == nil {
		return nil, 0, nil
	}
	merged = map[string]string{}
	fetched := map[string]bool{}
	for _, s := range reg.Sources() {
		if s.Type != "jira" {
			continue
		}
		p := s.Project
		if p == "" {
			p = fallbackProject
		}
		if p == "" || fetched[p] {
			continue
		}
		fetched[p] = true
		rs, ferr := ticket.FetchRemoteStatus(client, p)
		if ferr != nil {
			return nil, 0, fmt.Errorf("fetch remote status for project %s: %w", p, ferr)
		}
		for k, v := range rs {
			merged[k] = v
		}
	}
	return merged, len(fetched), nil
}

// applyActions applies create_issue actions first (filling each created key
// into its paired remap action), then transition actions, mutating actions
// in place. It collects errors per action instead of aborting on first error,
// allowing partial progress. Returns counts of successful actions and a list
// of error messages. Errors are formatted as "action_type issue/req: message".
func applyActions(client *ticket.JiraClient, actions []ticket.Action, project, issueType string) (created, transitioned int, errs []string) {
	for i := range actions {
		if actions[i].Type != "create_issue" {
			continue
		}
		key, cerr := client.CreateIssue(project, issueType, actions[i].Summary, actions[i].Description)
		if cerr != nil {
			errs = append(errs, fmt.Sprintf("create_issue %s: %v", actions[i].ReqID, cerr))
			continue
		}
		created++
		for j := i + 1; j < len(actions); j++ {
			if actions[j].Type == "remap" && actions[j].ReqID == actions[i].ReqID {
				actions[j].Issue = key
				break
			}
		}
	}

	for i := range actions {
		if actions[i].Type != "transition" {
			continue
		}
		if terr := client.TransitionIssue(actions[i].Issue, actions[i].To); terr != nil {
			errs = append(errs, fmt.Sprintf("transition %s: %v", actions[i].Issue, terr))
			continue
		}
		transitioned++
	}

	return created, transitioned, errs
}

// countPendingRemap counts remap actions whose Issue is still unfilled
// (e.g. because their paired create_issue never ran or failed before it).
func countPendingRemap(actions []ticket.Action) int {
	n := 0
	for _, a := range actions {
		if a.Type == "remap" && a.Issue == "" {
			n++
		}
	}
	return n
}

func writePlan(path string, actions []ticket.Action) error {
	if path == "" {
		return nil
	}
	if dir := filepath.Dir(path); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return fmt.Errorf("create plan directory: %w", err)
		}
	}
	if actions == nil {
		actions = []ticket.Action{} // always write a valid JSON array, never null
	}
	data, err := json.MarshalIndent(actions, "", "  ")
	if err != nil {
		return fmt.Errorf("encode plan: %w", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write plan %s: %w", path, err)
	}
	return nil
}

// remapMapFromActions extracts the {oldID: newID} map from completed remap
// actions, the shape `canary upgrade --map` expects.
func remapMapFromActions(actions []ticket.Action) map[string]string {
	m := map[string]string{}
	for _, a := range actions {
		if a.Type == "remap" && a.Issue != "" {
			m[a.ReqID] = a.Issue
		}
	}
	return m
}

func writeRemapMap(planPath string, actions []ticket.Action) error {
	if planPath == "" {
		return nil
	}
	m := remapMapFromActions(actions)
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("encode remap map: %w", err)
	}
	if err := os.WriteFile(planPath+".map.json", data, 0o600); err != nil {
		return fmt.Errorf("write remap map %s.map.json: %w", planPath, err)
	}
	return nil
}

// printActions prints a small-by-default table of actions, bounded by
// limit, with a "… +N more" hint on truncation — the full plan is always in
// the JSON file at --plan.
func printActions(cmd *cobra.Command, actions []ticket.Action, limit int) {
	out := cmd.OutOrStdout()
	if len(actions) == 0 {
		fmt.Fprintln(out, "No ticket-sync actions proposed")
		return
	}
	if limit <= 0 {
		limit = DefaultPlanLimit
	}
	shown := actions
	if len(shown) > limit {
		shown = shown[:limit]
	}
	for _, a := range shown {
		switch a.Type {
		case "create_issue":
			fmt.Fprintf(out, "[create_issue] %s: %s (source=%s)\n", a.ReqID, a.Summary, a.Source)
		case "transition":
			fmt.Fprintf(out, "[transition]   %s -> %s (source=%s)\n", a.Issue, a.To, a.Source)
		case "remap":
			issue := a.Issue
			if issue == "" {
				issue = "<pending>"
			}
			fmt.Fprintf(out, "[remap]        %s -> %s (source=%s)\n", a.ReqID, issue, a.Source)
		default:
			fmt.Fprintf(out, "[%s] %s (source=%s)\n", a.Type, a.ReqID, a.Source)
		}
	}
	if len(actions) > len(shown) {
		fmt.Fprintf(out, "… +%d more action(s) (use --limit %d)\n", len(actions)-len(shown), len(actions))
	}
}
