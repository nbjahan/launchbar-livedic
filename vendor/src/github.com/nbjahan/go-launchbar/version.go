package launchbar

import (
	"strconv"
	"strings"
)

// Version represents a version string (e.g. 1.0, 1.0, 1.0.0)
type Version string

func parseVersion(s string) (major, minor, patch int) {
	components := [3]int{0, 0, 0}
	parts := strings.Split(s, ".")
	for i, part := range parts {
		parti, _ := strconv.Atoi(part)
		components[i] = parti
	}
	major, minor, patch = components[0], components[1], components[2]
	return
}

// Cmp compares v and w and returns
//  -1 if v < w
//  0 if v == w
//  +1 if v > w
func (v Version) Cmp(w Version) int {
	v0, v1, v2 := parseVersion(string(v))
	w0, w1, w2 := parseVersion(string(w))
	switch {
	case v0 > w0:
		return 1
	case v0 < w0:
		return -1
	case v1 > w1:
		return 1
	case v1 < w1:
		return -1
	case v2 > w2:
		return 1
	case v2 < w2:
		return -1
	}
	return 0
}

// Less returns true if v < w
// Example:
//  Version("0.1.0").Less(Version("1.0")) == true
func (v Version) Less(w Version) bool {
	return v.Cmp(w) < 0
}

// Equal returns true if v == w
// Example:
//  Version("1.0").Equal(Version("1")) == true
func (v Version) Equal(w Version) bool {
	return v.Cmp(w) == 0
}
