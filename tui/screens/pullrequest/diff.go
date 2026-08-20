package pullrequest

import "strings"

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
		case strings.HasPrefix(line, "+++ "):
			if name := strings.TrimPrefix(line, "+++ "); name != "/dev/null" {
				current.Path = strings.TrimPrefix(name, "b/")
			}
		case strings.HasPrefix(line, "--- "):
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
	fields := strings.Fields(line)
	if len(fields) < 4 {
		return strings.TrimSpace(strings.TrimPrefix(line, "diff --git "))
	}
	return strings.TrimPrefix(fields[3], "b/")
}
