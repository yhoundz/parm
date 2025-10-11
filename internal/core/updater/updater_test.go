package updater

import "testing"

func TestNeedsUpdate(t *testing.T) {
	type needsUpdateData struct {
		curr     string
		latest   string
		expected bool
	}
	greaterThan := []needsUpdateData{
		{curr: "1.0.0", latest: "1.0.1", expected: true},
		{curr: "1.0.0", latest: "1.0.0", expected: false},
		{curr: "1.0.0-beta", latest: "1.0.0", expected: true},
		{curr: "v1.0.0", latest: "1.0.0", expected: false},
		{curr: "1.0", latest: "1.0.0", expected: false},
		{curr: "1.0.0", latest: "2.0.0", expected: true},
		{curr: "1.2.2", latest: "0.5.2", expected: false},
		{curr: "5.2.2-alpha", latest: "5.2.2", expected: true},
	}
	for _, test := range greaterThan {
		res, _ := NeedsUpdate(test.curr, test.latest)
		if res != test.expected {
			t.Errorf("test failed: had %t, wanted %t with curr %s and latest %s", res, test.expected, test.curr, test.latest)
		}
	}
	lessThan := []needsUpdateData{
		{curr: "0.9.9", latest: "1.0.0", expected: true},
		{curr: "1.0.0", latest: "1.0.1", expected: true},
		{curr: "1.0.0", latest: "1.1.0", expected: true},
		{curr: "1.0.0", latest: "2.0.0", expected: true},
		{curr: "1.2.3", latest: "1.2.3-rc.1", expected: false},
		{curr: "1.2.3", latest: "1.2.4-rc.1", expected: true},
	}
	for _, test := range lessThan {
		res, _ := NeedsUpdate(test.curr, test.latest)
		if res != test.expected {
			t.Errorf("test failed: had %t, wanted %t with curr %s and latest %s", res, test.expected, test.curr, test.latest)
		}
	}

	equalTo := []needsUpdateData{
		{curr: "1.0.0", latest: "1.0.0", expected: false},
		{curr: "1.2.3+build.1", latest: "1.2.3+build.2", expected: false},
		{curr: "v1.2.3", latest: "1.2.3", expected: false},
		{curr: "1.2.3", latest: "1.2.3+meta", expected: false},
	}
	for _, test := range equalTo {
		res, _ := NeedsUpdate(test.curr, test.latest)
		if res != test.expected {
			t.Errorf("test failed: had %t, wanted %t with curr %s and latest %s", res, test.expected, test.curr, test.latest)
		}
	}

	preReleaseCases := []needsUpdateData{
		{curr: "1.0.0-alpha", latest: "1.0.0-beta", expected: true},
		{curr: "1.0.0-beta", latest: "1.0.0-rc.1", expected: true},
		{curr: "1.0.0-rc.1", latest: "1.0.0-rc.2", expected: true},
		{curr: "1.0.0-alpha.1", latest: "1.0.0-alpha.2", expected: true},
		{curr: "1.0.0-rc.1", latest: "1.0.0-beta", expected: false},
		{curr: "1.0.0", latest: "1.0.0-rc.1", expected: false},
	}
	for _, test := range preReleaseCases {
		res, _ := NeedsUpdate(test.curr, test.latest)
		if res != test.expected {
			t.Errorf("test failed: had %t, wanted %t with curr %s and latest %s", res, test.expected, test.curr, test.latest)
		}
	}

	invalidCases := []needsUpdateData{
		{curr: "", latest: "1.0.0", expected: false},
		{curr: "1.0.0", latest: "", expected: false},
	}
	for _, test := range invalidCases {
		res, _ := NeedsUpdate(test.curr, test.latest)
		if res != test.expected {
			t.Errorf("test failed: had %t, wanted %t with curr %s and latest %s", res, test.expected, test.curr, test.latest)
		}
	}
}

func TestNeedsUpdateErrors(t *testing.T) {
	// invalid semver should error
	if _, err := NeedsUpdate("not-semver", "1.0.0"); err == nil {
		t.Fatalf("expected error for invalid current version")
	}
	if _, err := NeedsUpdate("1.0.0", "also-bad"); err == nil {
		t.Fatalf("expected error for invalid latest version")
	}
}
