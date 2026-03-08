---
description: This workflow helps the AI agent improve the internal structure of the code without changing its external behavior. It focuses on identifying code smells (long functions, duplication, tight coupling), planning the refactor carefully.
---

# Refactoring Workflow

## Step 1: Identify Code Smells

Examples:

- Long functions
- Duplicate code
- Tight coupling
- Poor naming

## Step 2: Understand Current Behavior

Before refactoring:

- Understand all logic
- Identify dependencies

## Step 3: Plan Refactor

Define:

- What will change
- What must remain unchanged

## Step 4: Apply Refactor

Examples:

- Extract functions
- Introduce interfaces
- Remove duplication

## Step 5: Validate Behavior

Ensure:

- Same output
- No regressions
