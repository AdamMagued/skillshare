package main

import (
	"fmt"
	"strings"

	"skillshare/internal/mcp"
	"skillshare/internal/theme"
	"skillshare/internal/ui"
)

// runMCPCheck is read-only: it writes no file and no operation log entry.
func runMCPCheck(service *mcp.Service, args []string) error {
	var opts mcp.CheckOptions
	asJSON := false
	for _, arg := range args {
		switch arg {
		case "--json":
			asJSON = true
		case "--no-dns":
			opts.SkipDNS = true
		default:
			if strings.HasPrefix(arg, "-") {
				return fmt.Errorf("unknown mcp check argument %q; usage: skillshare mcp check [name...] [--json] [--no-dns]", arg)
			}
			opts.Names = append(opts.Names, arg)
		}
	}
	report, err := service.Check(opts)
	if err != nil {
		if asJSON {
			return writeJSONError(err)
		}
		return err
	}
	var failed error
	if report.Summary.Errors > 0 {
		failed = fmt.Errorf("MCP check found %d error(s)", report.Summary.Errors)
	}
	if asJSON {
		return writeJSONResult(report, failed)
	}
	printMCPCheck(report)
	if failed != nil {
		// The summary line already says so.
		return &jsonSilentError{cause: failed}
	}
	return nil
}

func printMCPCheck(report *mcp.CheckReport) {
	ui.Info("MCP source: %s", report.SourcePath)
	a := theme.ANSI()
	marks := map[string]string{"error": ui.Colorize(a.Danger, "✗"), "warning": ui.Colorize(a.Warning, "!"), "info": ui.Colorize(a.Muted, "·")}
	for _, server := range report.Servers {
		heading := server.Name
		if server.Project != "" {
			heading += "  (project " + shortenPath(server.Project) + ")"
		}
		switch {
		case !server.OK:
			ui.Error("%s", heading)
		case mcpCheckHasWarning(server):
			ui.Warning("%s", heading)
		default:
			ui.Success("%s", heading)
		}
		for _, f := range server.Findings {
			message := f.Message
			if f.Target != "" {
				message = f.Target + ": " + message
			}
			fmt.Printf("  %s %s\n", marks[f.Level], message)
		}
	}
	fmt.Printf("%d server(s) checked: %d error(s), %d warning(s)\n", len(report.Servers), report.Summary.Errors, report.Summary.Warnings)
}

func mcpCheckHasWarning(server mcp.CheckServer) bool {
	for _, f := range server.Findings {
		if f.Level == "warning" {
			return true
		}
	}
	return false
}
