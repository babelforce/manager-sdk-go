package manager

// Guards the spec inventory. The //go:generate directives in generate.go are necessarily static
// (go:generate cannot read JSON), so this test pins them to specs/sources.json — the canonical
// spec list — and fails when either side gains or loses a spec.

import (
	"encoding/json"
	"os"
	"regexp"
	"testing"
)

func TestGenerateDirectivesMatchSpecInventory(t *testing.T) {
	raw, err := os.ReadFile("../specs/sources.json")
	if os.IsNotExist(err) {
		t.Skip("specs/sources.json is internal-only; the inventory guard runs in the monorepo")
	}
	if err != nil {
		t.Fatal(err)
	}
	var sources struct {
		Specs []struct {
			Name string `json:"name"`
		} `json:"specs"`
	}
	if err := json.Unmarshal(raw, &sources); err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{}
	for _, s := range sources.Specs {
		want[s.Name] = true
	}

	gen, err := os.ReadFile("generate.go")
	if err != nil {
		t.Fatal(err)
	}
	re := regexp.MustCompile(`(?m)^//go:generate .* \.\./specs/([a-z0-9-]+)\.openapi\.yaml$`)
	got := map[string]bool{}
	for _, m := range re.FindAllStringSubmatch(string(gen), -1) {
		got[m[1]] = true
	}
	if len(got) == 0 {
		t.Fatal("no //go:generate directives found in generate.go")
	}

	for name := range want {
		if !got[name] {
			t.Errorf("specs/sources.json lists %q but generate.go has no //go:generate directive for it", name)
		}
	}
	for name := range got {
		if !want[name] {
			t.Errorf("generate.go generates %q which is not in specs/sources.json", name)
		}
	}
}
