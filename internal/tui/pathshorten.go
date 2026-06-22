package tui

import "strings"

// shortenProjectPath compresses a project's namespace the way vim's
// pathshorten() does, so a long path does not push the MR title off-screen in
// the interactive list.
func shortenProjectPath(path string) string {
	segments := strings.Split(path, "/")
	if len(segments) <= 1 {
		return path
	}

	for i, seg := range segments[:len(segments)-1] {
		segments[i] = shortenSegment(seg)
	}
	return strings.Join(segments, "/")
}

func shortenSegment(seg string) string {
	r := []rune(seg)
	if len(r) == 0 {
		return seg
	}
	if r[0] == '.' && len(r) > 1 {
		return string(r[:2])
	}
	return string(r[:1])
}
