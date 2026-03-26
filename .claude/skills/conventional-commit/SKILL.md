---
name: conventional-commit
description: Creates git commits following the conventional commit specification (v1.0.0). Use when committing changes or creating commit messages.
---

> **IMPORTANT:** Use the `but commit` command. **NEVER** use `git commit`.

# Conventional Commit Skill

## Git Butler CLI (`but`)

This skill is designed for use with the Git Butler CLI tool called `but`. When creating commits, use the `but commit` command which supports conventional commit messages natively.

> **ALWAYS** use `but commit` to create commits.
> **NEVER** use the `-c` or `--create` flag with `but commit`. Always commit to an existing branch.
> **NEVER** use the `git commit` or any other `git` command to modify commits or amend them.

## Attribution

> **NEVER** include Claude/AI attribution in commit messages. Do not add:
> - "Generated with Claude Code" or similar
> - "Co-Authored-By: Claude" or any AI co-author lines
> - Any mention of AI assistance in commit messages

Generate commit messages that follow the [Conventional Commits v1.0.0](https://www.conventionalcommits.org/en/v1.0.0/) specification.

## Commit Message Format

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

## Commit Types

| Type | Description | SemVer |
|------|-------------|--------|
| `feat` | A new feature | MINOR |
| `fix` | A bug fix | PATCH |
| `docs` | Documentation only changes | - |
| `style` | Code style changes (formatting, semicolons, etc.) | - |
| `refactor` | Code change that neither fixes a bug nor adds a feature | - |
| `perf` | Performance improvement | - |
| `test` | Adding or correcting tests | - |
| `build` | Changes to build system or external dependencies | - |
| `ci` | Changes to CI configuration files and scripts | - |
| `chore` | Other changes that don't modify src or test files | - |
| `revert` | Reverts a previous commit | - |

## Rules

1. **Type is required** - Must be one of the types listed above
2. **Description is required** - Brief summary in imperative mood ("add" not "added")
3. **Scope is optional** - Noun describing section of codebase in parentheses
4. **Body is optional** - Detailed explanation of what and why (not how)
5. **Footer is optional** - For metadata like `BREAKING CHANGE:`, `Refs:`, `Reviewed-by:`

## Breaking Changes

Indicate breaking changes using either:
- Add `BREAKING CHANGE:` footer with description

Breaking changes trigger a MAJOR version bump.

## Examples

### Simple commit
```
feat: add user authentication
```

### With scope
```
feat(auth): add OAuth2 login support
```

### With body
```
fix(api): handle null response from server

The API was crashing when the server returned null instead of an empty array.
Added null check and default to empty array.
```

### With breaking change
```
BREAKING CHANGE: Response structure changed from {data: []} to {data: [], meta: {}, links: {}}
```

### With footer references
```
fix(parser): resolve memory leak in token handler

Refs: #123
Reviewed-by: Jane Doe
```

## Checklist

Before committing, verify:
- [ ] Type accurately reflects the change
- [ ] Description is concise (<50 chars ideal, <72 max)
- [ ] Description uses imperative mood ("add" not "added")
- [ ] Scope matches affected component (if used)
- [ ] Breaking changes are marked with `!` or footer
- [ ] Body explains why, not just what (if included)
