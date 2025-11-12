# KNOT Workflow Examples

## Example 1: New Feature Development

### Initial Setup

```bash
# Create project
knot project create --title "User Authentication System" --description "Implement JWT-based user authentication with login, registration, and password reset"
knot project select --id <project-id>

# Create main tasks
knot task create --title "Research authentication best practices" --description "Research JWT implementation, security standards, and similar projects" --complexity 3
knot task create --title "Design authentication system architecture" --description "Design database schema, API endpoints, and security flows" --complexity 6
knot task create --title "Implement user model and database" --description "Create user table, model, and database migrations" --complexity 5
knot task create --title "Implement JWT token management" --description "Create JWT generation, validation, and refresh logic" --complexity 8
knot task create --title "Implement authentication endpoints" --description "Create login, register, refresh, and logout endpoints" --complexity 9
knot task create --title "Add password reset functionality" --description "Implement secure password reset with email verification" --complexity 7
knot task create --title "Write comprehensive tests" --description "Unit tests, integration tests, and security tests" --complexity 6
knot task create --title "Update documentation" --description "API documentation, setup guides, and security notes" --complexity 4
```

### Setting Dependencies

```bash
# Research must come first
knot dependency add --task-id <design-id> --depends-on <research-id>

# Implementation depends on design
knot dependency add --task-id <user-model-id> --depends-on <design-id>
knot dependency add --task-id <jwt-id> --depends-on <design-id>

# JWT and user model must be complete before endpoints
knot dependency add --task-id <endpoints-id> --depends-on <user-model-id>
knot dependency add --task-id <endpoints-id> --depends-on <jwt-id>

# Password reset depends on endpoints
knot dependency add --task-id <password-reset-id> --depends-on <endpoints-id>

# Testing and documentation depend on implementation
knot dependency add --task-id <testing-id> --depends-on <endpoints-id>
knot dependency add --task-id <documentation-id> --depends-on <endpoints-id>
```

### Breaking Down Complex Tasks

```bash
# Check for tasks needing breakdown
knot breakdown

# Break down authentication endpoints (complexity 9)
knot task create --parent-id <endpoints-id> --title "Implement registration endpoint" --complexity 4
knot task create --parent-id <endpoints-id> --title "Implement login endpoint" --complexity 5
knot task create --parent-id <endpoints-id> --title "Implement token refresh endpoint" --complexity 4
knot task create --parent-id <endpoints-id> --title "Implement logout endpoint" --complexity 3
knot task create --parent-id <endpoints-id> --title "Add middleware for protected routes" --complexity 6

# Break down JWT implementation (complexity 8)
knot task create --parent-id <jwt-id> --title "Design JWT token structure" --complexity 3
knot task create --parent-id <jwt-id> --title "Implement JWT generation" --complexity 4
knot task create --parent-id <jwt-id> --title "Implement JWT validation" --complexity 5
knot task create --parent-id <jwt-id> --title "Implement token refresh logic" --complexity 4
```

### Working Through Tasks

```bash
# Find next actionable task
knot actionable
# Output: "Research authentication best practices"

# Start working
knot task update-state --id <research-id> --state in-progress

# Complete research
knot task update-state --id <research-id> --state completed

# Continue with next task
knot actionable
# Output: "Design authentication system architecture"
```

## Example 2: Bug Fix Workflow

### Bug Discovery and Initial Tasks

```bash
# Create investigation task
knot task create --title "Investigate: User login fails with valid credentials" --description "Users report login failures despite correct username/password" --complexity 4

# Start investigation
knot task update-state --id <investigation-id> --state in-progress
```

### During Investigation - Discovering Root Cause

```bash
# While investigating, discover the issue is in password hashing
knot task create --title "Root cause: Password hashing algorithm mismatch" --description "New registrations use bcrypt, but existing users have SHA-256 hashes" --complexity 5

# Create migration task
knot task create --title "Migrate existing password hashes to bcrypt" --description "Create secure migration script to update all user passwords" --complexity 7

# Create fix tasks
knot task create --title "Fix login to handle both hash types during migration" --description "Update authentication logic to support both algorithms temporarily" --complexity 6
knot task create --title "Update registration to use consistent hashing" --description "Ensure new registrations use correct algorithm" --complexity 3

# Create testing tasks
knot task create --title "Test migration script thoroughly" --description "Verify migration works without data loss" --complexity 8
knot task create --title "Test login with both old and new password formats" --description "Comprehensive testing of authentication during transition" --complexity 6
```

### Setting Dependencies for Bug Fix

```bash
# Root cause must be identified first
knot dependency add --task-id <migration-id> --depends-on <root-cause-id>
knot dependency add --task-id <fix-login-id> --depends-on <root-cause-id>

# Migration must be tested before deployment
knot dependency add --task-id <test-migration-id> --depends-on <migration-id>

# Login fix depends on migration understanding
knot dependency add --task-id <fix-login-id> --depends-on <migration-id>

# Comprehensive testing depends on fixes
knot dependency add --task-id <test-auth-id> --depends-on <fix-login-id>
knot dependency add --task-id <test-auth-id> --depends-on <registration-fix-id>
```

### Breaking Down Complex Testing

```bash
# Testing migration (complexity 8) needs breakdown
knot task create --parent-id <test-migration-id> --title "Create test database with sample users" --complexity 3
knot task create --parent-id <test-migration-id> --title "Test migration on small dataset" --complexity 4
knot task create --parent-id <test-migration-id> --title "Test migration rollback procedure" --complexity 5
knot task create --parent-id <test-migration-id> --title "Performance test migration on full dataset" --complexity 6
knot task create --parent-id <test-migration-id> --title "Verify no data corruption during migration" --complexity 7
```

## Example 3: Emergency Hotfix Workflow

### Critical Bug Discovery

```bash
# Immediate task creation for critical issue
knot task create --title "HOTFIX: Security vulnerability in password reset" --description "Password reset tokens are predictable, allowing account takeover" --complexity 9

# Immediately set to in-progress
knot task update-state --id <hotfix-id> --state in-progress

# Break down immediately (complexity 9)
knot task create --parent-id <hotfix-id> --title "Disable password reset functionality immediately" --complexity 2
knot task create --parent-id <hotfix-id> --title "Implement secure token generation" --complexity 6
knot task create --parent-id <hotfix-id> --title "Invalidate all existing password reset tokens" --complexity 3
knot task create --parent-id <hotfix-id> --title "Test new token generation thoroughly" --complexity 5
knot task create --parent-id <hotfix-id> --title "Deploy hotfix to production" --complexity 4
```

### Emergency Dependencies

```bash
# Critical sequence for security fix
knot dependency add --task-id <disable-id> --depends-on <hotfix-id>
knot dependency add --task-id <implement-id> --depends-on <disable-id>
knot dependency add --task-id <invalidate-id> --depends-on <implement-id>
knot dependency add --task-id <test-id> --depends-on <invalidate-id>
knot dependency add --task-id <deploy-id> --depends-on <test-id>
```

### Working the Hotfix

```bash
# Find first task
knot actionable
# Output: "Disable password reset functionality immediately"

# Work through critical sequence
knot task update-state --id <disable-id> --state in-progress
# ... implement disable ...
knot task update-state --id <disable-id> --state completed

# Continue to next critical task
knot actionable
# Output: "Implement secure token generation"
```

## Example 4: Research and Analysis Project

### Research Project Setup

```bash
knot project create --title "Microservices Architecture Research" --description "Research microservices patterns for potential system migration"
knot project select --id <project-id>

# Create research tasks
knot task create --title "Research current monolithic architecture" --description "Document current system structure and pain points" --complexity 4
knot task create --title "Research microservices patterns" --description "Study common patterns and best practices" --complexity 5
knot task create --title "Analyze competitors' architectures" --description "Research how similar companies handle microservices" --complexity 6
knot task create --title "Evaluate migration strategies" --description "Analyze different approaches to migration" --complexity 7
knot task create --title "Create proof of concept" --description "Implement small proof of concept" --complexity 8
knot task create --title "Write recommendation document" --description "Final architecture recommendation" --complexity 5
```

### Breaking Down Proof of Concept

```bash
# Proof of concept (complexity 8) needs breakdown
knot task create --parent-id <poc-id> --title "Design simple microservice" --complexity 3
knot task create --parent-id <poc-id> --title "Implement service discovery mechanism" --complexity 5
knot task create --parent-id <poc-id> --title "Implement inter-service communication" --complexity 5
knot task create --parent-id <poc-id> --title "Add monitoring and logging" --complexity 4
knot task create --parent-id <poc-id> --title "Test performance and reliability" --complexity 6
```

### Research Dependencies

```bash
# Research sequence
knot dependency add --task-id <patterns-id> --depends-on <current-id>
knot dependency add --task-id <competitors-id> --depends-on <patterns-id>
knot dependency add --task-id <strategies-id> --depends-on <competitors-id>
knot dependency add --task-id <poc-id> --depends-on <strategies-id>
knot dependency add --task-id <recommendation-id> --depends-on <poc-id>
```

## Example 5: Daily Workflow Management

### Morning Routine

```bash
# Check current project context
knot project get-selected

# Find today's tasks
knot actionable

# Review project health
knot breakdown
knot blocked

# Plan work based on actionable tasks
```

### During Development Day

```bash
# Working on task, discover new requirement
knot task create --title "Add data validation to user input" --description "Discovered during API endpoint testing" --complexity 4

# Set dependency
knot dependency add --task-id <validation-id> --depends-on <current-task-id>

# Continue with current task
```

### End of Day Review

```bash
# Update any completed tasks
knot task update-state --id <completed-task-id> --state completed

# Check tomorrow's work
knot actionable

# Review project progress
knot project get --id <project-id>

# Note any blockers or issues
knot blocked
```

These examples demonstrate how KNOT provides structure for complex workflows while maintaining flexibility for real-world development and task planning scenarios.