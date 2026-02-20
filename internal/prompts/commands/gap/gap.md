# Gap Command Prompt

## Purpose
Mark gap analysis claims as helpful or unhelpful.

## Task
Implement `canary gap mark` to track gap analysis verification results.

## Expected Behavior
```bash
# Mark a claim as helpful (correctly identified gap)
canary gap mark CBIN-001 --helpful

# Mark a claim as unhelpful (false positive)
canary gap mark CBIN-042 --unhelpful --reason "Already implemented"
```

## Output Format
```
Marked CBIN-001 as HELPFUL

Gap Analysis Claims for CBIN-001:
  Status: Correctly identified implementation gap
  Claim: "User authentication not fully tested"
  Actual: STATUS=IMPL (needs tests)
  Marked: HELPFUL by user on 2025-11-01

This helps improve future gap analysis accuracy.
```

## Use Cases
- Track accuracy of automated gap detection
- Identify false positives in gap analysis
- Improve gap detection algorithms
- Provide feedback on GAP_ANALYSIS.md claims

## Standards
- Store feedback in database
- Track who marked and when
- Optional reason for marking
- Show statistics on claim accuracy
- Feed into gap analysis improvements
