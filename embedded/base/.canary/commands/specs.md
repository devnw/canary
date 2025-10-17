---
description: List all requirement specification directories
---

<!-- CANARY: REQ=CBIN-145; FEATURE="SpecsCmd"; ASPECT=Docs; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-17 -->

## User Input

```text
$ARGUMENTS
```

## Outline

List all requirement specification directories in `.canary/specs/`.

1. **Parse arguments**:
   - Check for custom path (--path)
   - Check for JSON output (--json)

2. **Run canary specs command**:
   ```bash
   canary specs [flags]
   ```

   **Available flags:**
   - `--path <directory>`: Path to specs directory (default: `.canary/specs`)
   - `--json`: Output as JSON for programmatic parsing

   **Default behavior:**
   - Lists all subdirectories in `.canary/specs/`
   - Shows requirement ID and feature name (extracted from directory name)
   - Indicates which files exist (spec.md, plan.md)
   - Sorts by requirement ID

3. **Display results**:
   - Requirement ID and feature name
   - Directory path
   - Files present (spec.md, plan.md)
   - Total count of specifications

4. **Common usage patterns**:

   **List all specifications:**
   ```bash
   canary specs
   ```

   **Get JSON output for parsing:**
   ```bash
   canary specs --json
   ```

   **Check specific directory:**
   ```bash
   canary specs --path /path/to/custom/specs
   ```

5. **Use results for**:
   - Discovering available requirements
   - Checking which requirements have specs/plans
   - Validating spec directory structure
   - Programmatic spec directory processing

## Example Output

```markdown
## Specification Directories

Found 14 specification directories:

📁 {{.ReqID}}-132 - Next Priority Command
   .canary/specs/{{.ReqID}}-132-next-priority-command
   Files: spec.md, plan.md

📁 {{.ReqID}}-133 - ImplementCommand
   .canary/specs/{{.ReqID}}-133-ImplementCommand
   Files: spec.md, plan.md

📁 {{.ReqID}}-134 - Spec Modification
   .canary/specs/{{.ReqID}}-134-spec-modification
   Files: spec.md, plan.md

📁 {{.ReqID}}-135 - Priority List
   .canary/specs/{{.ReqID}}-135-priority-list
   Files: spec.md, plan.md

📁 {{.ReqID}}-137 - Requirement History
   .canary/specs/{{.ReqID}}-137-requirement-history
   Files: spec.md
   (no plan file)

Total: 14 specifications
```

## JSON Output Example

```json
[
  {
    "req_id": "{{.ReqID}}-132",
    "feature_name": "Next Priority Command",
    "directory": ".canary/specs/{{.ReqID}}-132-next-priority-command",
    "has_spec": true,
    "has_plan": true
  },
  {
    "req_id": "{{.ReqID}}-133",
    "feature_name": "ImplementCommand",
    "directory": ".canary/specs/{{.ReqID}}-133-ImplementCommand",
    "has_spec": true,
    "has_plan": true
  }
]
```

## Use Cases

**Specification Discovery:**
```bash
# What specifications exist?
canary specs

# Get machine-readable list
canary specs --json | jq -r '.[].req_id'
```

**Validation:**
```bash
# Check for missing plan files
canary specs --json | jq -r '.[] | select(.has_plan == false) | .req_id'

# Check for missing spec files
canary specs --json | jq -r '.[] | select(.has_spec == false) | .req_id'
```

**Integration with workflows:**
```bash
# Count specifications
canary specs --json | jq 'length'

# List incomplete specs
canary specs --json | jq -r '.[] | select(.has_spec == false or .has_plan == false)'
```

## Guidelines

- **Simple Discovery**: Quick way to see all requirements without database
- **No Database Required**: Works directly with filesystem
- **Structured Output**: Both human-readable and JSON formats
- **Validation Tool**: Identify missing spec.md or plan.md files
- **Fast Access**: Agents can quickly discover available specifications
- **Clean Output**: Emoji indicators and organized formatting
- **Sorted Results**: Alphabetically by requirement ID
- **Error Handling**: Clear messages if specs directory doesn't exist
