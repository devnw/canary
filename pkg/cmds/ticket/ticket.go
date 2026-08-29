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
// CANARY: REQ=ENG-3958; FEATURE="TicketDestination"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_ENG_3958_Sync_ApplyWithDestinationSource_NoProjectFlag,TestCANARY_ENG_3958_RemoteStatusForSources_MultiSourceMerge; UPDATED=2026-08-29
package ticket

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"devnw.dev/canary/pkg/config"
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
  sync   Compute a ticket-sync plan; apply it against JIRA with --apply`,
	}
	cmd.AddCommand(createTicketSyncCommand())
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

	reg := sources.LoadFromRoot(".")

	db, err := storage.Open(dbPath)
	if err != nil {
		return fmt.Errorf("open index db (run 'canary index' first): %w", err)
	}
	defer func() { _ = db.Close() }()

	// Honor the project id_pattern (like list/next) so fixture/example
	// tokens in the index never leak into ticket plans.
	idPattern := ""
	if cfg, cfgErr := config.Load("."); cfgErr == nil && cfg != nil {
		idPattern = cfg.Requirements.IDPattern
	}
	tokens, err := db.ListTokens(nil, idPattern, "req_id ASC", 0)
	if err != nil {
		return fmt.Errorf("list tokens: %w", err)
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
	if effProject == "" && hasUnresolvedCreateProject(actions) {
		return fmt.Errorf("--project is required with --apply (plan contains create_issue actions and no destination source has a configured project)")
	}

	client := &ticket.JiraClient{BaseURL: creds.BaseURL, Email: creds.Email, Token: creds.Token}

	remoteStatus, err := remoteStatusForSources(client, reg, project)
	if err != nil {
		return err
	}

	actions, err = ticket.ComputePlan(tokens, reg, remoteStatus)
	if err != nil {
		return err
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

// hasUnresolvedCreateProject reports whether the plan contains a
// create_issue action — the only action type that requires a resolvable
// project to apply. Transition actions target an already-existing issue by
// key and never need one.
func hasUnresolvedCreateProject(actions []ticket.Action) bool {
	for _, a := range actions {
		if a.Type == "create_issue" {
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
// CANARY: REQ=ENG-3958; FEATURE="TicketDestination"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_ENG_3958_RemoteStatusForSources_MultiSourceMerge; UPDATED=2026-08-29
func remoteStatusForSources(client *ticket.JiraClient, reg *sources.Registry, fallbackProject string) (map[string]string, error) {
	if reg == nil {
		return nil, nil
	}
	merged := map[string]string{}
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
		rs, err := ticket.FetchRemoteStatus(client, p)
		if err != nil {
			return nil, fmt.Errorf("fetch remote status for project %s: %w", p, err)
		}
		for k, v := range rs {
			merged[k] = v
		}
	}
	return merged, nil
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
