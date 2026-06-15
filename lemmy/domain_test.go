package lemmy

import (
	"testing"
)

// These tests cover the URI driver's pure string functions and the domain info.

func TestDomainInfo(t *testing.T) {
	info := Domain{}.Info()
	if info.Scheme != "lemmy" {
		t.Errorf("Scheme = %q, want lemmy", info.Scheme)
	}
	if len(info.Hosts) == 0 || info.Hosts[0] != Host {
		t.Errorf("Hosts = %v, want [%s]", info.Hosts, Host)
	}
	if info.Identity.Binary != "lemmy" {
		t.Errorf("Identity.Binary = %q, want lemmy", info.Identity.Binary)
	}
	if info.Identity.Short == "" {
		t.Error("Identity.Short is empty")
	}
}

func TestClassify(t *testing.T) {
	cases := []struct {
		in      string
		wantTyp string
		wantID  string
		wantErr bool
	}{
		{"12345", "post", "12345", false},
		{"", "", "", true},
		{"not-a-number", "", "", true},
	}
	for _, tc := range cases {
		typ, id, err := Domain{}.Classify(tc.in)
		if tc.wantErr {
			if err == nil {
				t.Errorf("Classify(%q): want error, got nil", tc.in)
			}
			continue
		}
		if err != nil {
			t.Errorf("Classify(%q): unexpected error: %v", tc.in, err)
			continue
		}
		if typ != tc.wantTyp || id != tc.wantID {
			t.Errorf("Classify(%q) = (%q, %q), want (%q, %q)",
				tc.in, typ, id, tc.wantTyp, tc.wantID)
		}
	}
}

func TestLocate(t *testing.T) {
	got, err := Domain{}.Locate("post", "123")
	want := "https://" + Host + "/post/123"
	if err != nil || got != want {
		t.Errorf("Locate = (%q, %v), want (%q, nil)", got, err, want)
	}

	_, err = Domain{}.Locate("unknown", "123")
	if err == nil {
		t.Error("Locate with unknown type: want error, got nil")
	}
}
