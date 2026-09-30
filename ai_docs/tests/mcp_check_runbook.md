# CLI E2E Runbook: `skillshare mcp check`

## Scope

Verify the read-only MCP check against a global config whose servers cover every
finding: an unset `fromEnv` variable, a command missing from `PATH`, a `.invalid` host,
unsynced targets, a client rule refusal, `targets: []`, and one `mcp.projects` root.
Covers human output, the `--json` shape including `project`, `--no-dns`, name
filtering, exit codes 0 and 1, and `GET /api/mcp/check` with and without `?dns=0`.
No server is launched, no command is run, and nothing is synced.

## Environment

Run inside the devcontainer with a fresh ssenv HOME:

```bash
ssenv create mcp-check-e2e --init
cd /tmp
ssenv enter mcp-check-e2e -- mdproof --report json /workspace/ai_docs/tests/mcp_check_runbook.md
```

Human-output steps set `NO_COLOR=1` so the status marks can be matched as plain text.
Every step works from `$HOME` and points `SKILLSHARE_CONFIG` at
`$HOME/mcp-check/config.yaml`. Never run these steps with the repository root as the
working directory: the CLI would pick up the repository's `.skillshare/` as project
mode. The check needs DNS only for Step 4; `.invalid` never resolves, online or not.
Step 7 starts `skillshare ui` on port 47931; pick another free port if it is taken.

## Steps

### Step 1: Write the global config

The skills directory must exist before `skillshare ui` starts. `mcp list` would refuse
this config because of the `workspace` server, so the step reads it back through a check
of the one server that has no problems.

```bash
set -eu
cd "$HOME"
CASE="$HOME/mcp-check"
rm -rf "$CASE"
mkdir -p "$CASE/skills" "$CASE/app"
export SKILLSHARE_CONFIG="$CASE/config.yaml"
cat > "$SKILLSHARE_CONFIG" <<YAML
source: $CASE/skills
targets: {}
mcp:
  targets: [claude, cursor]
  servers:
    docs:
      url: https://example.com/docs
    envless:
      command: sh
      env:
        API_KEY: {fromEnv: SS_E2E_CHECK_TOKEN}
    ghost:
      command: no-such-mcp-binary
    offline:
      url: https://mcp.skillshare.invalid/mcp
      targets: [cursor]
    parked:
      url: https://example.com/parked
      targets: []
    workspace:
      url: https://example.com/workspace
  projects:
    ~/mcp-check/app:
      targets: [claude]
      servers:
        docs:
          command: no-such-project-binary
YAML
ss mcp check parked --json -g
```

Expected:
- exit_code: 0
- jq: [.servers[].name] == ["parked"]
- jq: .summary == {"errors": 0, "warnings": 0}

### Step 2: Human output lists every server and exits 1

```bash
cd "$HOME"
export SKILLSHARE_CONFIG="$HOME/mcp-check/config.yaml"
unset SS_E2E_CHECK_TOKEN
NO_COLOR=1 ss mcp check --no-dns -g
```

Expected:
- exit_code: 1
- ✗ envless
- ✗ env API_KEY reads SS_E2E_CHECK_TOKEN, which is not set
- ✗ command no-such-mcp-binary was not found on PATH
- ✗ claude: Claude MCP workspace: Claude Code reserves this name
- ! cursor: not synced yet; run skillshare sync mcp
- · kept in Skillshare only; no Agent receives it
- ✗ docs  (project ~/mcp-check/app)
- ✗ command no-such-project-binary was not found on PATH
- 7 server(s) checked: 4 error(s), 9 warning(s)

### Step 3: JSON shape and the project field

```bash
cd "$HOME"
export SKILLSHARE_CONFIG="$HOME/mcp-check/config.yaml"
unset SS_E2E_CHECK_TOKEN
ss mcp check --json --no-dns -g
```

Expected:
- exit_code: 1
- jq: [.servers[].name] == ["docs", "envless", "ghost", "offline", "parked", "workspace", "docs"]
- jq: [.servers[] | select(has("project"))] | length == 1
- jq: .servers[6].project | startswith("/") and endswith("/mcp-check/app")
- jq: .servers[6].ok == false and .servers[6].findings[0].check == "command" and .servers[6].findings[0].subject == "no-such-project-binary"
- jq: .servers[6].findings[1] == {"level": "warning", "check": "sync", "target": "claude", "message": "not synced yet; run skillshare sync mcp"}
- jq: .servers[1].findings[0] | .check == "env" and .subject == "SS_E2E_CHECK_TOKEN" and .target == ""
- jq: .servers[5].findings[0] | .check == "client-rule" and .target == "claude"
- jq: .servers[4] | .ok == true and .findings == [{"level": "info", "check": "targets", "target": "", "message": "kept in Skillshare only; no Agent receives it"}]
- jq: [.servers[].findings[] | select(.check == "dns")] == []
- jq: .summary == {"errors": 4, "warnings": 9}

### Step 4: DNS warning, and --no-dns skips it

A DNS warning never fails the check, so both runs exit 0.

```bash
set -eu
cd "$HOME"
export SKILLSHARE_CONFIG="$HOME/mcp-check/config.yaml"
ss mcp check offline --json -g > "$HOME/mcp-check/dns.json"
ss mcp check offline --json --no-dns -g > "$HOME/mcp-check/no-dns.json"
jq -n --slurpfile dns "$HOME/mcp-check/dns.json" --slurpfile skip "$HOME/mcp-check/no-dns.json" '{dns: $dns[0], skip: $skip[0]}'
```

Expected:
- exit_code: 0
- jq: .dns.servers[0].findings[0] == {"level": "warning", "check": "dns", "target": "", "message": "host mcp.skillshare.invalid did not resolve", "subject": "mcp.skillshare.invalid"}
- jq: .dns.servers[0].ok == true and .dns.summary == {"errors": 0, "warnings": 2}
- jq: [.skip.servers[0].findings[].check] == ["sync"]

### Step 5: Name filtering, unknown names, and exit 0

```bash
set -eu
cd "$HOME"
export SKILLSHARE_CONFIG="$HOME/mcp-check/config.yaml"
ss mcp check docs --json --no-dns -g > "$HOME/mcp-check/docs.json" || test $? -eq 1
if ss mcp check nope -g > "$HOME/mcp-check/unknown.txt" 2>&1; then echo "unknown name must fail"; exit 1; fi
ss mcp check parked -g > "$HOME/mcp-check/parked.txt"
jq -c '[.servers[] | {name, project}]' "$HOME/mcp-check/docs.json"
cat "$HOME/mcp-check/unknown.txt" "$HOME/mcp-check/parked.txt"
```

Expected:
- exit_code: 0
- regex: \[\{"name":"docs","project":null\},\{"name":"docs","project":"[^"]+/mcp-check/app"\}\]
- unknown MCP server "nope"; known servers: docs, envless, ghost, offline, parked, workspace
- 1 server(s) checked: 0 error(s), 0 warning(s)

### Step 6: A set variable passes and its value is never printed

```bash
set -eu
cd "$HOME"
export SKILLSHARE_CONFIG="$HOME/mcp-check/config.yaml"
SS_E2E_CHECK_TOKEN=s3cr3t-e2e-value NO_COLOR=1 ss mcp check envless --no-dns -g > "$HOME/mcp-check/envless.txt"
SS_E2E_CHECK_TOKEN=s3cr3t-e2e-value ss mcp check envless --json --no-dns -g >> "$HOME/mcp-check/envless.txt"
if grep -q s3cr3t-e2e-value "$HOME/mcp-check/envless.txt"; then echo "value leaked"; exit 1; fi
head -n 5 "$HOME/mcp-check/envless.txt"
```

Expected:
- exit_code: 0
- ! envless
- 1 server(s) checked: 0 error(s), 2 warning(s)

### Step 7: The dashboard endpoint, with and without ?dns=0

```bash
set -eu
cd "$HOME"
export SKILLSHARE_CONFIG="$HOME/mcp-check/config.yaml"
unset SS_E2E_CHECK_TOKEN
PORT=47931
ss ui --port "$PORT" --no-open -g > "$HOME/mcp-check/ui.log" 2>&1 &
UI_PID=$!
trap 'kill "$UI_PID" 2>/dev/null || true' EXIT
for _ in $(seq 1 100); do
  curl -fsS "http://127.0.0.1:$PORT/api/mcp/check?dns=0" > "$HOME/mcp-check/api-no-dns.json" 2>/dev/null && break
  sleep 0.2
done
curl -fsS "http://127.0.0.1:$PORT/api/mcp/check" > "$HOME/mcp-check/api-dns.json"
jq -n --slurpfile skip "$HOME/mcp-check/api-no-dns.json" --slurpfile dns "$HOME/mcp-check/api-dns.json" '{skip: $skip[0], dns: $dns[0]}'
```

Expected:
- exit_code: 0
- jq: .skip.summary == {"errors": 4, "warnings": 9}
- jq: [.skip.servers[] | select(.project)] | length == 1
- jq: [.skip.servers[].findings[] | select(.check == "dns")] == []
- jq: .dns.summary == {"errors": 4, "warnings": 10}
- jq: [.dns.servers[].findings[] | select(.check == "dns") | .subject] == ["mcp.skillshare.invalid"]

## Pass Criteria

- Steps 1–7 pass.
- Errors exit 1 and warnings alone exit 0.
- `project` appears only on the `mcp.projects` server and holds the expanded root.
- No step prints the value of `SS_E2E_CHECK_TOKEN`.
