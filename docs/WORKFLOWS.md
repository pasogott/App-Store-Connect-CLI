# Workflows (Fastlane-Style Lanes)

`asc workflow` lets you define named, multi-step automation sequences in a repo-local file: `.asc/workflow.json`.

This is designed as a Fastlane-style "lanes" replacement: a single, versioned workflow file that composes existing `asc` commands and normal shell commands.

## Quick Start

1. Create `.asc/workflow.json` in your repo.
2. Validate the file:

```bash
asc workflow validate
```

3. Run a workflow:

```bash
asc workflow run beta
asc workflow run beta BUILD_ID:123456789 GROUP_ID:abcdef
```

## Example `.asc/workflow.json`

Notes:
- The file supports JSONC comments (`//` and `/* */`).
- Output is JSON on stdout; step and hook command output streams to stderr.
- On failures, stdout still remains JSON-only and includes a top-level `error` plus `hooks` results.

```json
{
  "env": {
    "APP_ID": "123456789",
    "VERSION": "1.0.0"
  },
  "before_all": "asc auth status",
  "after_all": "echo workflow_done",
  "error": "echo workflow_failed",
  "workflows": {
    "beta": {
      "description": "Distribute a build to a TestFlight group",
      "env": {
        "GROUP_ID": ""
      },
      "steps": [
        {
          "name": "list_builds",
          "run": "asc builds list --app $APP_ID --sort -uploadedDate --limit 5"
        },
        {
          "name": "list_groups",
          "run": "asc testflight beta-groups list --app $APP_ID --limit 20"
        },
        {
          "name": "add_build_to_group",
          "if": "BUILD_ID",
          "run": "asc builds add-groups --build $BUILD_ID --group $GROUP_ID"
        }
      ]
    },
    "release": {
      "description": "Submit a version for App Store review",
      "steps": [
        {
          "workflow": "sync-metadata",
          "with": {
            "FASTLANE_DIR": "./fastlane/metadata"
          }
        },
        {
          "name": "submit",
          "run": "asc submit create --app $APP_ID --version $VERSION --build $BUILD_ID --confirm"
        }
      ]
    },
    "sync-metadata": {
      "private": true,
      "description": "Private helper workflow (callable only via workflow steps)",
      "steps": [
        {
          "name": "migrate_validate",
          "run": "asc migrate validate --fastlane-dir $FASTLANE_DIR"
        }
      ]
    }
  }
}
```

## Semantics

### Environment Merging

- Entry workflow env: `definition.env` -> `workflow.env` -> CLI params (`KEY:VALUE` / `KEY=VALUE`)
- Sub-workflow env: `sub_workflow.env` provides defaults, caller env overrides, and step `with` overrides win over everything.

### Conditionals

Add `"if": "VAR_NAME"` to a step to skip it when the variable is falsy.

Truthy values (case-insensitive): `1`, `true`, `yes`, `y`, `on`.

### Hooks

Hooks are definition-level commands:
- `before_all` runs once before any steps
- `after_all` runs once after all steps (only if steps succeeded)
- `error` runs on any failure (step failure, hook failure, max call depth, etc.)

Hooks are recorded in the structured JSON output as `hooks.before_all`, `hooks.after_all`, and `hooks.error`.

### Output Contract

- stdout: JSON-only (`asc workflow run` prints a structured result)
- stderr: step/hook command output, plus dry-run previews

This makes it safe to do:

```bash
asc workflow run beta BUILD_ID:123 GROUP_ID:xyz | jq -e '.status == "ok"'
```

