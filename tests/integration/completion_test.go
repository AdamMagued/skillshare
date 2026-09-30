//go:build !online

package integration

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"skillshare/internal/testutil"
)

func TestCompletion_Bash_OutputsScript(t *testing.T) {
	sb := testutil.NewSandbox(t)
	defer sb.Cleanup()

	result := sb.RunCLI("completion", "bash")
	result.AssertSuccess(t)
	result.AssertOutputContains(t, "complete -F _skillshare skillshare")
	for _, cmd := range []string{"sync", "install", "list", "target", "commit", "completion"} {
		result.AssertOutputContains(t, cmd)
	}
}

func TestCompletion_Zsh_OutputsScript(t *testing.T) {
	sb := testutil.NewSandbox(t)
	defer sb.Cleanup()

	result := sb.RunCLI("completion", "zsh")
	result.AssertSuccess(t)
	result.AssertOutputContains(t, "#compdef skillshare")
	result.AssertOutputContains(t, "_skillshare")
}

func TestCompletion_Fish_OutputsScript(t *testing.T) {
	sb := testutil.NewSandbox(t)
	defer sb.Cleanup()

	result := sb.RunCLI("completion", "fish")
	result.AssertSuccess(t)
	result.AssertOutputContains(t, "complete -c skillshare")
	result.AssertOutputContains(t, "__fish_skillshare_no_subcommand")
}

func TestCompletion_PowerShell_OutputsScript(t *testing.T) {
	sb := testutil.NewSandbox(t)
	defer sb.Cleanup()

	result := sb.RunCLI("completion", "powershell")
	result.AssertSuccess(t)
	result.AssertOutputContains(t, "Register-ArgumentCompleter")
}

func TestCompletion_Nushell_OutputsScript(t *testing.T) {
	sb := testutil.NewSandbox(t)
	defer sb.Cleanup()

	result := sb.RunCLI("completion", "nushell")
	result.AssertSuccess(t)
	result.AssertOutputContains(t, "export extern \"skillshare\"")
}

func TestCompletion_MCPCheck_AllShells(t *testing.T) {
	sb := testutil.NewSandbox(t)
	defer sb.Cleanup()

	for shell, subcommands := range map[string]string{
		"bash":       "add check import list remove restore",
		"zsh":        "(add check edit import list remove restore)",
		"fish":       "-a 'add check edit import list remove restore'",
		"powershell": "@{ Name = 'check'; Desc = 'MCP check' }",
		"nushell":    "[add check import list remove restore]",
	} {
		result := sb.RunCLI("completion", shell)
		result.AssertSuccess(t)
		result.AssertOutputContains(t, subcommands)
		result.AssertOutputContains(t, "no-dns")
		result.AssertOutputContains(t, "live")
		result.AssertOutputContains(t, "timeout")
	}
}

func TestCompletion_UnsupportedShell_Errors(t *testing.T) {
	sb := testutil.NewSandbox(t)
	defer sb.Cleanup()

	result := sb.RunCLI("completion", "tcsh")
	result.AssertFailure(t)
	result.AssertAnyOutputContains(t, "unsupported shell")
}

func TestCompletion_NoArgs_ShowsUsage(t *testing.T) {
	sb := testutil.NewSandbox(t)
	defer sb.Cleanup()

	result := sb.RunCLI("completion")
	result.AssertSuccess(t)
	result.AssertOutputContains(t, "USAGE")
}

func TestCompletion_Install_WritesFile(t *testing.T) {
	sb := testutil.NewSandbox(t)
	defer sb.Cleanup()

	result := sb.RunCLI("completion", "bash", "--install")
	result.AssertSuccess(t)
	result.AssertAnyOutputContains(t, "installed")

	destPath := filepath.Join(sb.Home, ".local", "share", "bash-completion", "completions", "skillshare")
	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		t.Errorf("expected completion script at %s, but file does not exist", destPath)
	}
}

// zshCapture completes a command line in an interactive zsh under zpty and prints each
// candidate compadd receives. Arguments: the fpath directory, then the command line.
const zshCapture = `zmodload zsh/zpty
zpty z zsh -f -i
zpty -w z "PS1=''; fpath=($1 \$fpath); autoload -Uz compinit; compinit -u -d $1/.zcompdump"
zpty -w z "bindkey '^I' complete-word; comppostfuncs=( exit )"
zpty -w z 'compadd () {
  if [[ ${@[1,(i)(-|--)]} == *-(O|A|D)\ * ]]; then builtin compadd "$@"; return $?; fi
  typeset -a __hits
  builtin compadd -A __hits "$@"
  local h; for h in $__hits; do print -r -- "HIT:$h"; done
}'
zpty -w z "$2"$'\t'
local line
while zpty -r z line; do
  line=${line%$'\r'}
  # The first candidate can share a line with the echoed input; skip the echoed definition.
  [[ $line == *HIT:* && $line != *'HIT:$h'* ]] && print -r -- ${line##*HIT:}
done
zpty -d z
`

// completeIn runs a real shell and returns its candidates, one per line; it skips when
// the shell is not installed, as on most CI images.
func completeIn(t *testing.T, shell string, args ...string) []string {
	t.Helper()
	if _, err := exec.LookPath(shell); err != nil {
		t.Skipf("%s is not installed", shell)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, shell, args...)
	cmd.Dir = t.TempDir()
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %v: %v\n%s", shell, args, err, out)
	}
	var candidates []string
	for _, line := range strings.Split(string(out), "\n") {
		// fish prints "candidate<TAB>description".
		if name, _, _ := strings.Cut(strings.TrimSpace(line), "\t"); name != "" {
			candidates = append(candidates, name)
		}
	}
	return candidates
}

func assertCandidates(t *testing.T, got []string, want ...string) {
	t.Helper()
	for _, w := range want {
		if !slices.Contains(got, w) {
			t.Errorf("missing completion %q in %q", w, got)
		}
	}
}

func TestCompletion_Fish_CompletesMCPCheck(t *testing.T) {
	sb := testutil.NewSandbox(t)
	defer sb.Cleanup()
	result := sb.RunCLI("completion", "fish")
	result.AssertSuccess(t)
	script := filepath.Join(sb.Home, "skillshare.fish")
	if err := os.WriteFile(script, []byte(result.Stdout), 0644); err != nil {
		t.Fatal(err)
	}

	assertCandidates(t, completeIn(t, "fish", "--no-config", "-c", "source "+script+"; complete -C 'skillshare mcp '"), "add", "check", "edit", "import", "list", "remove", "restore")
	assertCandidates(t, completeIn(t, "fish", "--no-config", "-c", "source "+script+"; complete -C 'skillshare mcp check --'"), "--no-dns", "--live", "--timeout")
	assertCandidates(t, completeIn(t, "fish", "--no-config", "-c", "source "+script+"; complete -C 'skillshare mcp add --'"), "--url", "--target", "--sync", "--replace", "--dry-run", "--json", "--no-tui")
}

func TestCompletion_Zsh_CompletesMCPCheck(t *testing.T) {
	sb := testutil.NewSandbox(t)
	defer sb.Cleanup()
	result := sb.RunCLI("completion", "zsh")
	result.AssertSuccess(t)
	dir := filepath.Join(sb.Home, "zsh-functions")
	capture := filepath.Join(sb.Home, "capture.zsh")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "_skillshare"), []byte(result.Stdout), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(capture, []byte(zshCapture), 0644); err != nil {
		t.Fatal(err)
	}

	completeIn(t, "zsh", "-n", filepath.Join(dir, "_skillshare"))
	assertCandidates(t, completeIn(t, "zsh", "-f", capture, dir, "skillshare mcp "), "add", "check", "edit", "import", "list", "remove", "restore")
	assertCandidates(t, completeIn(t, "zsh", "-f", capture, dir, "skillshare mcp check --"), "--no-dns", "--live", "--timeout")
}
