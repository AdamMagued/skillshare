package mcp

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"os"
	"os/exec"
	"slices"
	"strings"
	"time"
)

// CheckFinding is one result of a check. It never carries an environment value.
type CheckFinding struct {
	Level   string `json:"level"`
	Check   string `json:"check"`
	Target  string `json:"target"`
	Message string `json:"message"`
}

// CheckServer holds the findings for one server; OK means none is an error.
type CheckServer struct {
	Name     string         `json:"name"`
	OK       bool           `json:"ok"`
	Findings []CheckFinding `json:"findings"`
}

type CheckSummary struct {
	Errors   int `json:"errors"`
	Warnings int `json:"warnings"`
}

// CheckReport answers, per server, whether it will work as synced.
type CheckReport struct {
	SourcePath string        `json:"-"`
	Servers    []CheckServer `json:"servers"`
	Summary    CheckSummary  `json:"summary"`
}

// CheckOptions selects servers and supplies the lookups, which tests replace.
type CheckOptions struct {
	Names      []string
	SkipDNS    bool
	LookupEnv  func(string) (string, bool)
	LookPath   func(string) (string, error)
	LookupHost func(context.Context, string) ([]string, error)
}

// dnsTimeout bounds each host lookup, the only network access a check makes.
const dnsTimeout = 3 * time.Second

// Check verifies the source's servers without starting them or sending requests: referenced
// variables are set, commands resolve, hosts resolve, and the plan has each entry in sync.
func (s *Service) Check(opts CheckOptions) (*CheckReport, error) {
	if opts.LookupEnv == nil {
		opts.LookupEnv = os.LookupEnv
	}
	if opts.LookPath == nil {
		opts.LookPath = exec.LookPath
	}
	if opts.LookupHost == nil {
		opts.LookupHost = net.DefaultResolver.LookupHost
	}
	source, err := LoadSource(s.ConfigPath)
	if err != nil {
		return nil, err
	}
	names := opts.Names
	if len(names) == 0 {
		names = sortedKeys(source.Servers)
	}
	for _, name := range names {
		if _, ok := source.Servers[name]; !ok {
			known := strings.Join(sortedKeys(source.Servers), ", ")
			if known == "" {
				known = "none"
			}
			return nil, fmt.Errorf("unknown MCP server %q; known servers: %s", name, known)
		}
	}
	report := &CheckReport{SourcePath: source.Path, Servers: []CheckServer{}}
	// Every server's client rules count, selected or not: the plan below covers them all.
	preflight := map[string][]CheckFinding{}
	// refused holds the targets whose rules refuse a server; an empty name means all of them.
	refused := map[string][]string{}
	for name, server := range source.Servers {
		preflight[name] = s.checkClientRules(source, name, server)
		for _, f := range preflight[name] {
			if f.Level == "error" {
				refused[name] = append(refused[name], f.Target)
			}
		}
	}
	seen := map[string]bool{}
	for _, name := range names {
		if seen[name] {
			continue
		}
		seen[name] = true
		server := source.Servers[name]
		result := CheckServer{Name: name, Findings: []CheckFinding{}}
		result.Findings = append(result.Findings, checkEnv(server, opts.LookupEnv)...)
		if !server.Disabled {
			result.Findings = append(result.Findings, checkLaunch(server, opts)...)
		}
		result.Findings = append(result.Findings, preflight[name]...)
		report.Servers = append(report.Servers, result)
	}
	byName := map[string]*CheckServer{}
	for i := range report.Servers {
		byName[report.Servers[i].Name] = &report.Servers[i]
	}
	// Plan only what the Agents accept, so one refused entry does not hide the rest.
	accepted := *source
	accepted.Servers = map[string]Server{}
	for name, server := range source.Servers {
		if len(refused[name]) > 0 {
			if slices.Contains(refused[name], "") {
				continue
			}
			server.Targets = slices.DeleteFunc(slices.Clone(server.Targets.orDefault(source.Targets)), func(target string) bool {
				return slices.Contains(refused[name], target)
			})
		}
		accepted.Servers[name] = server
	}
	p, err := s.previewSource(&accepted)
	if err != nil {
		return nil, err
	}
	var synced []Change
	for _, c := range p.Changes {
		if c.Root == "" && byName[c.Name] != nil && !slices.Contains(refused[c.Name], c.Target) {
			synced = append(synced, c)
		}
	}
	slices.SortStableFunc(synced, func(a, b Change) int { return strings.Compare(a.Target, b.Target) })
	for _, c := range synced {
		result := byName[c.Name]
		result.Findings = append(result.Findings, syncFinding(c))
	}
	for i := range report.Servers {
		result := &report.Servers[i]
		result.OK = true
		for _, f := range result.Findings {
			switch f.Level {
			case "error":
				result.OK = false
				report.Summary.Errors++
			case "warning":
				report.Summary.Warnings++
			}
		}
	}
	return report, nil
}

// checkEnv reports each fromEnv reference whose variable is unset or empty.
func checkEnv(server Server, lookup func(string) (string, bool)) []CheckFinding {
	var out []CheckFinding
	missing := func(field string, v Value) {
		if v.FromEnv == "" {
			return
		}
		if value, ok := lookup(v.FromEnv); !ok || value == "" {
			out = append(out, CheckFinding{Level: "error", Check: "env", Message: fmt.Sprintf("%s reads %s, which is not set", field, v.FromEnv)})
		}
	}
	for _, key := range sortedKeys(server.Env) {
		missing("env "+key, server.Env[key])
	}
	for _, key := range sortedKeys(server.Headers) {
		missing("header "+key, server.Headers[key])
	}
	if server.BearerToken != nil {
		missing("bearerToken", *server.BearerToken)
	}
	return out
}

// checkLaunch resolves a stdio command on PATH, or a remote host through DNS.
func checkLaunch(server Server, opts CheckOptions) []CheckFinding {
	if server.Command != "" {
		command, err := expandHome(server.Command)
		if err == nil {
			_, err = opts.LookPath(command)
		}
		if err != nil {
			return []CheckFinding{{Level: "error", Check: "command", Message: fmt.Sprintf("command %s was not found on PATH", server.Command)}}
		}
		return nil
	}
	u, err := url.Parse(server.URL)
	if err != nil || u.Hostname() == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return []CheckFinding{{Level: "error", Check: "url", Message: "url is not a valid HTTP(S) URL"}}
	}
	host := u.Hostname()
	if opts.SkipDNS || net.ParseIP(host) != nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), dnsTimeout)
	defer cancel()
	if _, err := opts.LookupHost(ctx, host); err != nil {
		return []CheckFinding{{Level: "warning", Check: "dns", Message: fmt.Sprintf("host %s did not resolve", host)}}
	}
	return nil
}

// checkClientRules renders the server alone, once per target, so a rule an Agent enforces
// names that Agent. A server that names no targets is kept in Skillshare only.
func (s *Service) checkClientRules(source *Source, name string, server Server) []CheckFinding {
	if server.Targets != nil && len(server.Targets) == 0 {
		return []CheckFinding{{Level: "info", Check: "targets", Message: "kept in Skillshare only; no Agent receives it"}}
	}
	render := func(server Server) error {
		alone := *source
		alone.Servers = map[string]Server{name: server}
		alone.Projects = nil
		_, err := s.render(&alone)
		return err
	}
	selected := server.Targets.orDefault(source.Targets)
	// A switch-only entry works out its own targets, so it is rendered as written.
	if server.Disabled || len(selected) == 0 {
		if err := render(server); err != nil {
			return []CheckFinding{{Level: "error", Check: "client-rule", Message: err.Error()}}
		}
		return nil
	}
	var out []CheckFinding
	for _, target := range selected {
		one := server
		one.Targets = TargetList{target}
		if err := render(one); err != nil {
			out = append(out, CheckFinding{Level: "error", Check: "client-rule", Target: target, Message: err.Error()})
		}
	}
	return out
}

// syncFinding says whether an Agent's entry matches what the source asks for.
func syncFinding(c Change) CheckFinding {
	f := CheckFinding{Level: "info", Check: "sync", Target: c.Target, Message: "in sync"}
	switch c.Action {
	case "unchanged":
		// Claude's local scope server of the same name wins over the synced one.
		if c.Message != "" {
			f.Level, f.Message = "warning", c.Message
		}
	case "adopt":
		f.Message = "already present; the next sync takes it over"
	case "add", "update":
		f.Level, f.Message = "warning", "not synced yet; run skillshare sync mcp"
	case "remove":
		f.Level, f.Message = "warning", "not synced yet; the next sync removes it from this Agent"
	case "conflict":
		f.Level, f.Message = "error", c.Message
	default:
		f.Level, f.Message = "warning", c.Action
	}
	return f
}

// orDefault is the list a server is written to: its own, or the config's default.
func (t TargetList) orDefault(defaults []string) TargetList {
	if t == nil {
		return defaults
	}
	return t
}
