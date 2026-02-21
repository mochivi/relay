package router

import "testing"

func TestTree_Insert(t *testing.T) {
	testCases := []struct {
		name     string
		inserts  [][]string
		searches []struct {
			pattern  string
			expected string
			found    bool
		}
	}{
		{
			name:    "single_insert",
			inserts: [][]string{{"/api", "api"}},
			searches: []struct {
				pattern  string
				expected string
				found    bool
			}{
				{"/api", "api", true},
			},
		},
		{
			name:    "root_catchall",
			inserts: [][]string{{"/", "root"}},
			searches: []struct {
				pattern  string
				expected string
				found    bool
			}{
				{"/", "root", true},
				{"/anything", "root", true},
				{"/anything/nested", "root", true},
			},
		},
		{
			name:    "nested_insert",
			inserts: [][]string{{"/api/web/test", "test"}},
			searches: []struct {
				pattern  string
				expected string
				found    bool
			}{
				{"/api/web/test", "test", true},
				{"/api/web/test/extra", "test", true}, // longest prefix match
				{"/api/web", "", false},
				{"/api", "", false},
			},
		},
		{
			name: "multiple_routes_longest_prefix",
			inserts: [][]string{
				{"/api", "api"},
				{"/api/web", "web"},
				{"/api/web/test", "test"},
			},
			searches: []struct {
				pattern  string
				expected string
				found    bool
			}{
				{"/api", "api", true},
				{"/api/web", "web", true},
				{"/api/web/test", "test", true},
				{"/api/web/test/extra", "test", true}, // falls back to longest match
				{"/api/web/other", "web", true},       // falls back to /api/web
				{"/api/other", "api", true},           // falls back to /api
				{"/other", "", false},
			},
		},
		{
			name: "root_and_nested",
			inserts: [][]string{
				{"/", "root"},
				{"/api", "api"},
			},
			searches: []struct {
				pattern  string
				expected string
				found    bool
			}{
				{"/", "root", true},
				{"/api", "api", true},
				{"/other", "root", true},       // falls back to root
				{"/api/extra", "api", true},    // falls back to /api
				{"/other/extra", "root", true}, // falls back to root
			},
		},
		{
			name: "overwrite_existing",
			inserts: [][]string{
				{"/api", "api_v1"},
				{"/api", "api_v2"},
			},
			searches: []struct {
				pattern  string
				expected string
				found    bool
			}{
				{"/api", "api_v2", true},
			},
		},
		{
			name: "sibling_routes",
			inserts: [][]string{
				{"/api", "api"},
				{"/web", "web"},
				{"/health", "health"},
			},
			searches: []struct {
				pattern  string
				expected string
				found    bool
			}{
				{"/api", "api", true},
				{"/web", "web", true},
				{"/health", "health", true},
				{"/other", "", false},
			},
		},
		{
			name:    "no_match",
			inserts: [][]string{{"/api/web", "web"}},
			searches: []struct {
				pattern  string
				expected string
				found    bool
			}{
				{"/other", "", false},
				{"/api/other", "", false},
				{"/", "", false},
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			tree := tree{}
			for i := range tt.inserts {
				if err := tree.insert(tt.inserts[i][0], tt.inserts[i][1]); err != nil {
					t.Fatalf("insert(%q) unexpected error: %v", tt.inserts[i][0], err)
				}
			}
			for _, s := range tt.searches {
				res, ok := tree.search(s.pattern)
				if ok != s.found || res != s.expected {
					t.Errorf("search(%q): expected (%q, %v), got (%q, %v)", s.pattern, s.expected, s.found, res, ok)
				}
			}
		})
	}
}

// TestTree_OnlyRegisteredPathsHaveVal documents that only nodes where a route was
// registered have val set (are tied to a service). So /api and /web/dashboard
// both registered: /api -> api, /web -> no route, /web/dashboard -> dashboard.
func TestTree_OnlyRegisteredPathsHaveVal(t *testing.T) {
	tree := tree{}
	tree.insert("/api", "api")
	tree.insert("/web/dashboard", "dashboard")

	tests := []struct {
		path    string
		wantOK  bool
		wantVal string
	}{
		{"/api", true, "api"},
		{"/web", false, ""}, // path exists but no route registered at /web
		{"/web/dashboard", true, "dashboard"},
		{"/web/other", false, ""}, // no route
		{"/unknown", false, ""},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			val, ok := tree.search(tt.path)
			if ok != tt.wantOK {
				t.Errorf("search(%q) ok = %v, want %v", tt.path, ok, tt.wantOK)
			}
			if tt.wantOK && val != tt.wantVal {
				t.Errorf("search(%q) val = %q, want %q", tt.path, val, tt.wantVal)
			}
		})
	}
}

// TestTree_RootCatchallWithAdmin mirrors config: path / -> api, path /admin -> admin.
// /api must match the root route (api), not 404.
func TestTree_RootCatchallWithAdmin(t *testing.T) {
	tree, err := newTreeFromPatterns(
		[]string{"/", "/admin"},
		[]string{"api", "admin"},
	)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		path    string
		wantVal string
		wantOK  bool
	}{
		{"/", "api", true},
		{"/api", "api", true},
		{"/anything", "api", true},
		{"/admin", "admin", true},
		{"/admin/", "admin", true},
		{"/admin/foo", "admin", true},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			val, ok := tree.search(tt.path)
			if ok != tt.wantOK || val != tt.wantVal {
				t.Errorf("search(%q) = (%q, %v), want (%q, %v)", tt.path, val, ok, tt.wantVal, tt.wantOK)
			}
		})
	}
}

func TestTree_MultiWildcard(t *testing.T) {
	tree, _ := newTreeFromPatterns(
		[]string{
			"/",
			"/api",
			"/api/*/v1",
			"/api/*/v1/*/settings",
			"/api/*/v1/*/settings/advanced",
			"/api/*/v2/*/dashboard",
		},
		[]string{
			"root",
			"api",
			"v1",
			"settings",
			"advanced",
			"dashboard",
		},
	)

	tests := []struct {
		path    string
		wantOK  bool
		wantVal string
		desc    string
	}{
		// Exact multi-wildcard matches
		{"/api/tenant1/v1/user1/settings", true, "settings", "both wildcards match arbitrary segments"},
		{"/api/org-abc/v1/profile/settings", true, "settings", "wildcards match slugs with dashes"},
		{"/api/123/v1/456/settings", true, "settings", "wildcards match numeric segments"},
		{"/api/tenant1/v1/user1/settings/advanced", true, "advanced", "three level wildcard chain matches"},

		// Second wildcard pattern
		{"/api/tenant1/v2/user1/dashboard", true, "dashboard", "v2 wildcard pattern matches"},
		{"/api/org/v2/profile/dashboard", true, "dashboard", "v2 wildcard matches different org"},

		// Fallback when tail doesn't match — should fall back to deepest registered prefix
		{"/api/tenant1/v1/user1/other", true, "v1", "unmatched tail falls back to /api/*/v1"},
		{"/api/tenant1/v1/user1/settings/unknown", true, "settings", "unmatched tail falls back to settings"},
		{"/api/tenant1/v2/user1/other", true, "api", "v2 unmatched tail falls back to /api (no /api/*/v2 registered)"},

		// Partial paths that don't reach the terminal wildcard node
		{"/api/tenant1/v1", true, "v1", "exact match on intermediate wildcard node /api/*/v1"},
		{"/api/tenant1/v3", true, "api", "unknown version falls back to /api"},

		// Cross-version should not bleed
		{"/api/tenant1/v2/user1/settings", true, "api", "settings not registered under v2, falls back to /api"},

		// Root fallback
		{"/completely/unknown/path", true, "root", "unknown path falls back to root"},
		{"/api/only", true, "api", "partial api path falls back to /api"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			val, ok := tree.search(tt.path)
			if ok != tt.wantOK {
				t.Errorf("search(%q) ok = %v, want %v", tt.path, ok, tt.wantOK)
			}
			if tt.wantOK && val != tt.wantVal {
				t.Errorf("search(%q) val = %q, want %q", tt.path, val, tt.wantVal)
			}
		})
	}
}
