package cmd

import "strings"

const usageTemplate = `Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{formatExamples .Example}}{{end}}{{if .HasAvailableSubCommands}}{{$cmds := .Commands}}{{if eq (len .Groups) 0}}

Available Commands:{{range $cmds}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{else}}{{range $group := .Groups}}

{{.Title}}{{range $cmds}}{{if (and (eq .GroupID $group.ID) (or .IsAvailableCommand (eq .Name "help")))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if not .AllChildCommandsHaveGroup}}

Additional Commands:{{range $cmds}}{{if (and (eq .GroupID "") (or .IsAvailableCommand (eq .Name "help")))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`

func formatExamples(examples string) string {
	lines := strings.Split(strings.ReplaceAll(examples, "\r\n", "\n"), "\n")

	start := 0
	for start < len(lines) && strings.TrimSpace(lines[start]) == "" {
		start++
	}

	end := len(lines) - 1
	for end >= start && strings.TrimSpace(lines[end]) == "" {
		end--
	}

	if start > end {
		return ""
	}

	trimmed := lines[start : end+1]
	commonIndent := ""
	hasIndent := false
	for _, line := range trimmed {
		if strings.TrimSpace(line) == "" {
			continue
		}

		indent := leadingWhitespace(line)
		if !hasIndent {
			commonIndent = indent
			hasIndent = true
			continue
		}

		commonIndent = commonPrefix(commonIndent, indent)
	}

	for i, line := range trimmed {
		if strings.TrimSpace(line) == "" {
			trimmed[i] = ""
			continue
		}

		if commonIndent != "" && strings.HasPrefix(line, commonIndent) {
			line = line[len(commonIndent):]
		}

		trimmed[i] = "  " + line
	}

	return strings.Join(trimmed, "\n")
}

func leadingWhitespace(s string) string {
	for i, r := range s {
		if r != ' ' && r != '\t' {
			return s[:i]
		}
	}

	return s
}

func commonPrefix(a, b string) string {
	limit := len(a)
	if len(b) < limit {
		limit = len(b)
	}

	i := 0
	for i < limit && a[i] == b[i] {
		i++
	}

	return a[:i]
}
