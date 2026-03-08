---
description: This workflow helps the AI agent systematically identify, analyze, and fix software bugs or unexpected behaviors. It emphasizes reproducing the issue, tracing the root cause, and applying the safest possible fix while avoiding unnecessary changes.
---

# Debugging Workflow

## Step 1: Identify the Problem

- Understand the reported issue
- Determine expected behavior
- Determine actual behavior

## Step 2: Reproduce the Issue

Try to reproduce the error locally.

Check:

- Logs
- Stack traces
- API responses

## Step 3: Locate the Source

Inspect:

- Related functions
- Database queries
- API calls
- State changes

## Step 4: Analyze Root Cause

Common causes:

- Incorrect logic
- Missing validation
- Race conditions
- Incorrect API usage
- Dependency issues

## Step 5: Fix the Issue

Implement the smallest safe fix.

Avoid:

- Large refactors
- Unrelated changes

## Step 6: Validate the Fix

Test:

- Original failing scenario
- Edge cases
- Regression issues
