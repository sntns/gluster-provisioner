package device

import (
	"testing"

	"github.com/sntns/gluster-provisioner/pkg/model"
)

func TestPathFilter_Match(t *testing.T) {
	filter := PathFilter{Pattern: "/dev/sd*"}
	cases := []struct {
		name string
		path string
		want bool
	}{
		{"match sda", "/dev/sda", true},
		{"match sdb1", "/dev/sdb1", true},
		{"no match vda", "/dev/vda", false},
		{"invalid pattern", "", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			info := model.DeviceInfo{Path: c.path}
			got := filter.Match(info)
			if got != c.want {
				t.Errorf("PathFilter.Match(%q) = %v, want %v", c.path, got, c.want)
			}
		})
	}
}

func TestFilters_Match_OR(t *testing.T) {
	f1 := PathFilter{Pattern: "/dev/sda"}
	f2 := PathFilter{Pattern: "/dev/vda"}
	filters := Filters{
		Items: []Filter{f1, f2},
		Op:    OpOr,
	}
	if !filters.Match(model.DeviceInfo{Path: "/dev/sda"}) {
		t.Error("OR: expected match for /dev/sda")
	}
	if !filters.Match(model.DeviceInfo{Path: "/dev/vda"}) {
		t.Error("OR: expected match for /dev/vda")
	}
	if filters.Match(model.DeviceInfo{Path: "/dev/xvda"}) {
		t.Error("OR: did not expect match for /dev/xvda")
	}
}

func TestFilters_Match_AND(t *testing.T) {
	f1 := PathFilter{Pattern: "/dev/sd*"}
	f2 := PathFilter{Pattern: "/dev/sda"}
	filters := Filters{
		Items: []Filter{f1, f2},
		Op:    OpAnd,
	}
	if !filters.Match(model.DeviceInfo{Path: "/dev/sda"}) {
		t.Error("AND: expected match for /dev/sda")
	}
	if filters.Match(model.DeviceInfo{Path: "/dev/sdb"}) {
		t.Error("AND: did not expect match for /dev/sdb")
	}
}

func TestFilters_Match_Empty(t *testing.T) {
	filters := Filters{Items: nil, Op: OpOr}
	if !filters.Match(model.DeviceInfo{Path: "/dev/any"}) {
		t.Error("Empty: expected match for any device when no filters")
	}
}

func TestFilters_Complex_AND_OR(t *testing.T) {
	// ( ( /dev/sda OR /dev/vda ) AND /dev/sd* )
	orFilters := Filters{
		Items: []Filter{
			PathFilter{Pattern: "/dev/sda"},
			PathFilter{Pattern: "/dev/vda"},
		},
		Op: OpOr,
	}
	complexFilters := Filters{
		Items: []Filter{
			orFilters,
			PathFilter{Pattern: "/dev/sd*"},
		},
		Op: OpAnd,
	}
	testCases := []struct {
		name string
		path string
		want bool
	}{
		{"match sda", "/dev/sda", true},
		{"no match vda (not sd*)", "/dev/vda", false},
		{"no match sdb", "/dev/sdb", false},
		{"no match random", "/dev/xyz", false},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			info := model.DeviceInfo{Path: tc.path}
			got := complexFilters.Match(info)
			if got != tc.want {
				t.Errorf("Complex AND/OR: Match(%q) = %v, want %v", tc.path, got, tc.want)
			}
		})
	}
}
