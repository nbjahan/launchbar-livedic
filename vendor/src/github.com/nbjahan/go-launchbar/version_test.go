package launchbar

import "testing"

func TestVersionCmp(t *testing.T) {
	tests := []struct {
		a, b string
		out  int
	}{
		{"1.0.0", "1.0.0", 0},
		{"0", "0.1", -1},
		{"0", "1", -1},
		{"0.0", "0.0", 0},
		{"0.0.0", "0.0.0", 0},
		{"0.1", "0", 1},
		{"0.1", "0.1", 0},
		{"0.1", "1.0", -1},
		{"0.1.0", "1.0", -1},
		{"0.1.0", "1.1.0", -1},
		{"1", "0", 1},
		{"1", "1", 0},
		{"1.0", "0", 1},
		{"1.0", "0.9", 1},
		{"1.0", "1", 0},
		{"1.0", "1.0", 0},
		{"1.0.0", "0", 1},
		{"1.0.0", "0.0", 1},
		{"1.0.0", "0.0.0", 1},
		{"1.0.0", "1", 0},
		{"1.0.0", "1.0", 0},
		{"1.0.0", "1.0.1", -1},
		{"1.0.0", "1.1.0", -1},
		{"1.0.0", "1.1.1", -1},
		{"1.1", "1.0.1", 1},
		{"1.1.0", "1.0.0", 1},
		{"1.1.1", "1.2", -1},
		{"1.2.1", "1.2", 1},
		{"11.0.1", "10.9.1", 1},
		{"11.1.2", "11.1.3", -1},
		{"1.12", "1.11", 1},
		{"1.12", "1.12", 0},
		{"1.12", "1.13", -1},
	}

	for _, test := range tests {
		out := Version(test.a).Cmp(Version(test.b))
		if out != test.out {
			t.Errorf("%q Cmp %q? expected %v, got %v", test.a, test.b, test.out, out)
		}
	}
}

func TestVersionEqual(t *testing.T) {
	tests := []struct {
		a, b string
		out  bool
	}{
		{"1.0.0", "1.0.0", true},
		{"1.0.0", "1.0", true},
		{"1.0.0", "1", true},
		{"1.0", "1.0", true},
		{"1.0", "1", true},
		{"1", "1", true},
		{"1.0.0", "0.0.0", false},
		{"1.0.0", "0.0", false},
		{"1.0.0", "0", false},
		{"1.0", "0", false},
		{"1", "0", false},
		{"1.12", "1.12", true},
		{"1.12", "1.012", true},
		{"1.12", "1.13", false},
	}

	for _, test := range tests {
		out := Version(test.a).Equal(Version(test.b))
		if out != test.out {
			t.Errorf("%q Equal %q? expected %v, got %v", test.a, test.b, test.out, out)
		}
	}
}

func TestVersionLess(t *testing.T) {
	tests := []struct {
		a, b string
		out  bool
	}{
		{"0.0.0", "0.0.0", false},
		{"0.0", "0.0", false},
		{"0", "1", true},
		{"1.0.0", "1.0.1", true},
		{"1.0.0", "1.1.1", true},
		{"1.0.0", "1.1.0", true},
		{"0.1.0", "1.1.0", true},
		{"1.1.0", "1.0.0", false},
		{"1.0", "0.9", false},
		{"11.0.1", "10.9.1", false},
		{"11.1.2", "11.1.3", true},
		{"1.1", "1.0.1", false},
		{"1.1.1", "1.2", true},
		{"1.2.1", "1.2", false},
		{"0.1", "1.0", true},
		{"0.1", "0.1", false},
		{"0.1", "0", false},
		{"0", "0.1", true},
		{"0.1.0", "1.0", true},
		{"0.12", "0.13", true},
		{"0.12", "1.12", true},
		{"0.12", "0.12", false},
	}

	for _, test := range tests {
		out := Version(test.a).Less(Version(test.b))
		if out != test.out {
			t.Errorf("%q Less than %q? expected %v, got %v", test.a, test.b, test.out, out)
		}
	}
}
