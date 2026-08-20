package pullrequest

import (
	"strconv"
	"strings"
)

type DiffHunk struct {
	Kind byte
	Text string
}

type DiffFile struct {
	Path  string
	Hunks []DiffHunk
}

type twoColumnRow struct {
	header      bool
	text        string
	left, right *DiffHunk
}

func twoColumnRows(hunks []DiffHunk) []twoColumnRow {
	rows := make([]twoColumnRow, 0, len(hunks))
	for i := 0; i < len(hunks); {
		if hunks[i].Kind == '@' {
			rows = append(rows, twoColumnRow{header: true, text: hunks[i].Text})
			i++
			continue
		}
		if hunks[i].Kind == ' ' {
			line := hunks[i]
			rows = append(rows, twoColumnRow{left: &line, right: &line})
			i++
			continue
		}
		var removed, added []DiffHunk
		for i < len(hunks) && hunks[i].Kind != ' ' && hunks[i].Kind != '@' {
			switch hunks[i].Kind {
			case '-':
				removed = append(removed, hunks[i])
			case '+':
				added = append(added, hunks[i])
			}
			i++
		}
		for j := 0; j < max(len(removed), len(added)); j++ {
			row := twoColumnRow{}
			if j < len(removed) {
				row.left = &removed[j]
			}
			if j < len(added) {
				row.right = &added[j]
			}
			rows = append(rows, row)
		}
	}
	return rows
}

func ParseDiff(diff string) []DiffFile {
	var files []DiffFile
	var current *DiffFile
	inHunk := false
	lines := strings.Split(diff, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, "diff --git "):
			files = append(files, DiffFile{Path: pathFromHeader(line)})
			current = &files[len(files)-1]
			inHunk = false
		case current == nil:
			continue
		case !inHunk && strings.HasPrefix(line, "+++ "):
			if name := gitPath(strings.TrimPrefix(line, "+++ ")); name != "/dev/null" {
				current.Path = strings.TrimPrefix(name, "b/")
			}
		case !inHunk && strings.HasPrefix(line, "--- "):
			continue
		case strings.HasPrefix(line, "@@"):
			inHunk = true
			current.Hunks = append(current.Hunks, DiffHunk{Kind: '@', Text: line})
		case !inHunk:
			continue
		case strings.HasPrefix(line, `\`):
			continue
		case line == "":
			current.Hunks = append(current.Hunks, DiffHunk{Kind: ' ', Text: ""})
		default:
			current.Hunks = append(current.Hunks, DiffHunk{Kind: line[0], Text: line[1:]})
		}
	}
	return files
}

func pathFromHeader(line string) string {
	rest := strings.TrimSpace(strings.TrimPrefix(line, "diff --git "))
	if !strings.HasPrefix(rest, `"`) {
		if split := strings.LastIndex(rest, " b/"); split >= 0 {
			return strings.TrimPrefix(strings.TrimSpace(rest[split+1:]), "b/")
		}
	}
	fields := gitTokens(rest)
	if len(fields) > 1 {
		for _, field := range fields[1:] {
			if strings.HasPrefix(field, "b/") {
				return strings.TrimPrefix(field, "b/")
			}
		}
	}
	if len(fields) > 0 {
		return strings.TrimPrefix(fields[len(fields)-1], "b/")
	}
	return rest
}

func gitPath(value string) string {
	value = strings.TrimSpace(value)
	if strings.HasPrefix(value, `"`) {
		fields := gitTokens(value)
		if len(fields) > 0 {
			return fields[0]
		}
	}
	return value
}

// gitTokens handles the C-style quoting used by git for paths containing
// whitespace or special characters.
func gitTokens(value string) []string {
	var fields []string
	for value != "" {
		value = strings.TrimLeft(value, " \t")
		if value == "" {
			break
		}
		if value[0] != '"' {
			end := strings.IndexAny(value, " \t")
			if end < 0 {
				fields = append(fields, value)
				break
			}
			fields = append(fields, value[:end])
			value = value[end:]
			continue
		}
		end := 1
		for end < len(value) {
			if value[end] == '\\' {
				end += 2
				continue
			}
			if value[end] == '"' {
				end++
				break
			}
			end++
		}
		raw := value[:min(end, len(value))]
		field, err := strconv.Unquote(raw)
		if err != nil {
			field = strings.Trim(raw, `"`)
		}
		fields = append(fields, field)
		value = value[min(end, len(value)):]
	}
	return fields
}
