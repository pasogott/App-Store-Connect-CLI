# ASC CLI behavior test matrix

Apply the cases relevant to the changed behavior, under the test rules in `AGENTS.md`. Use existing coverage when it proves the contract. Add tests only for distinct observable behavior or demonstrated regressions; do not repeat shared parser and renderer coverage for each command.

## Flags and parsing

- Cover the valid path and one invalid value (stderr and exit code `2`) of a new or changed flag, unless an existing test already does. Add them as rows in an existing table when one exists.
- When parsing or dispatch changes, cover affected flag ordering and values that look like subcommand names.
- Assert required-flag errors on stderr; do not test only for `flag.ErrHelp`.
- Never accept and silently ignore an unsupported flag or value.

## Output and exit behavior

- Parse JSON or XML and assert fields instead of relying only on substring matching.
- Assert representative table or Markdown structure and headers.
- Verify stdout, stderr, and exit code with a built binary for changed CLI behavior.
- Test successful artifact/report writes and write failures.
- Prevent duplicate usage or duplicate error output.

## HTTP and API behavior

- Use `httptest` to assert method, path, query, headers, and request body.
- Cover a realistic non-empty response per response family. Test validation and API errors only where the endpoint handles them differently from the shared client.
- Keep representative response-decoding assertions when consolidating table-driven tests.
- Test pagination and empty responses where applicable.

## Auth and process isolation

- Run repository tests with `ASC_BYPASS_KEYCHAIN=1`.
- Use `t.Setenv` and a temporary `ASC_CONFIG_PATH` for auth-sensitive tests.
- Set or clear `ASC_PROFILE`, `ASC_KEY_ID`, `ASC_ISSUER_ID`, `ASC_PRIVATE_KEY_PATH`, `ASC_PRIVATE_KEY`, `ASC_PRIVATE_KEY_B64`, and `ASC_STRICT_AUTH` locally when relevant.
- Restore exact process state when a test cannot use `t.Setenv`.
- Allow `t.Skip` only for a specific documented and reproducible condition; never use a broad error-string skip.

## Live verification

- Prefer read-only calls first.
- For authorized mutations and cleanup, use a disposable resource, record its ID, and delete or cancel it afterward.
- Assert the externally visible result, not only a successful HTTP status.
- Report any state that could not be cleaned up.
