# Agent gap-analysis policy (for AI agents and maintainers)

This document records **Won't do** and **out-of-scope** decisions for gap tracking so future agents do not reopen them.

## CI/CD Hash, Docker tag pinning, pre-commit (Closed — Won't do)

The following are **closed in gap tracking as Won't do** for this repo; handled via different controls (org/CI, deployment, or local tooling). **For future agents:** do not reopen gaps solely about these topics.

- **CI/CD hash pinning** – Handled via different controls (e.g. org/CI pipeline, not repo-level gap work).
- **Docker tag pinning** – Handled via different controls; not tracked as a gap here.
- **Pre-commit hooks** – Handled via different controls; not a requirement for closing gaps in `*GAP_ANALYSIS.md`.

When updating gap tracking:

- Do **not** open or leave open gaps that only request add CI, CI/CD hash pinning, pin Docker tags, or pre-commit hooks as the resolution.
- If a gap was closed as **Won't do** with reference to this policy, do not reopen it unless the policy is explicitly changed.

## GitLab as source of truth

Requirements and gap status are authoritative in **GitLab issues**. Markdown `*GAP_ANALYSIS.md` files are tracking mirrors; keep them in sync with GitLab when closing or updating gaps.
