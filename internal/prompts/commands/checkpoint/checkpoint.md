# Checkpoint Command Prompt

## Purpose
Create a snapshot of current requirement tracking state.

## Task
Implement `canary checkpoint` to save point-in-time state for comparison.

## Expected Behavior
```bash
# Create checkpoint
canary checkpoint create "Before refactoring auth module"

# List checkpoints
canary checkpoint list

# Compare current state to checkpoint
canary checkpoint diff checkpoint-2025-11-01

# Restore checkpoint (if supported)
canary checkpoint restore checkpoint-2025-11-01
```

## Output Format
```
Checkpoint created: checkpoint-2025-11-01-123456

Captured state:
  Total requirements: 42
  By status:
    BENCHED: 2
    TESTED: 8
    IMPL: 18
    STUB: 14
  
  Completion: 42%

Checkpoint saved to: .canary/checkpoints/checkpoint-2025-11-01-123456.json
```

## Checkpoint Data
- Timestamp
- Full token database snapshot
- Summary statistics
- Optional description/tag
- Git commit hash (if in git repo)

## Use Cases
- Track progress over time
- Compare before/after refactoring
- Milestone snapshots
- Rollback if needed
- Generate progress reports

## Standards
- Store as JSON in `.canary/checkpoints/`
- Include full token details
- Capture git state if available
- Support diff between checkpoints
- Show what changed (added, modified, removed tokens)
