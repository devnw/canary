#!/usr/bin/env bash
# CANARY: REQ=CBIN-108; FEATURE="RequirementScript"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2025-10-16
# Create a new CANARY requirement specification directory and file

set -euo pipefail

# Parse arguments
REQ_ID=""
FEATURE=""

while [[ $# -gt 0 ]]; do
  case $1 in
    --req-id)
      REQ_ID="$2"
      shift 2
      ;;
    --feature)
      FEATURE="$2"
      shift 2
      ;;
    *)
      echo "Unknown argument: $1" >&2
      exit 1
      ;;
  esac
done

# Validate required arguments
if [[ -z "$REQ_ID" ]]; then
  echo "Error: --req-id required" >&2
  echo "Usage: $0 --req-id {{.ProjectKey}}-XXX --feature \"feature-name\"" >&2
  exit 1
fi

if [[ -z "$FEATURE" ]]; then
  echo "Error: --feature required" >&2
  echo "Usage: $0 --req-id {{.ProjectKey}}-XXX --feature \"feature-name\"" >&2
  exit 1
fi

# Create directory structure
SPEC_DIR=".canary/specs/${REQ_ID}-${FEATURE}"
mkdir -p "$SPEC_DIR"

# Initialize spec file from template
SPEC_FILE="$SPEC_DIR/spec.md"
if [[ -f ".canary/templates/spec-template.md" ]]; then
  cp ".canary/templates/spec-template.md" "$SPEC_FILE"

  # Replace placeholders
  DATE=$(date +%Y-%m-%d)
  sed -i "s/{{.ProjectKey}}-XXX/${REQ_ID}/g" "$SPEC_FILE"
  sed -i "s/\[FEATURE NAME\]/${FEATURE}/g" "$SPEC_FILE"
  sed -i "s/YYYY-MM-DD/${DATE}/g" "$SPEC_FILE"
else
  # Create minimal spec if template doesn't exist
  cat > "$SPEC_FILE" <<EOF
# Feature Specification: ${FEATURE}

**Requirement ID:** ${REQ_ID}
**Status:** STUB
**Created:** $(date +%Y-%m-%d)

## Overview
[Feature description]

## User Stories
[User stories]

## Functional Requirements
[Requirements]

## Success Criteria
[Criteria]
EOF
fi

# Output JSON for parsing by AI agent
cat <<EOF
{
  "req_id": "${REQ_ID}",
  "feature": "${FEATURE}",
  "spec_dir": "${SPEC_DIR}",
  "spec_file": "${SPEC_FILE}",
  "status": "created"
}
EOF
