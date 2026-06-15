package lemmy

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Rate != 500*time.Millisecond {
		t.Errorf("Rate = %v, want 500ms", cfg.Rate)
	}
	if cfg.Retries <= 0 {
		t.Errorf("Retries = %d, want > 0", cfg.Retries)
	}
	if cfg.Timeout <= 0 {
		t.Errorf("Timeout = %v, want > 0", cfg.Timeout)
	}
	if cfg.UserAgent == "" {
		t.Error("UserAgent is empty")
	}
	if cfg.BaseURL == "" {
		t.Error("BaseURL is empty")
	}
}

func TestNewClientNotNil(t *testing.T) {
	c := NewClient(DefaultConfig())
	if c == nil {
		t.Fatal("NewClient returned nil")
	}
}

func noRateConfig() Config {
	cfg := DefaultConfig()
	cfg.Rate = 0
	cfg.Retries = 0
	return cfg
}

func TestListPosts(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") == "" {
			t.Error("request has no User-Agent")
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"posts": []map[string]any{
				{
					"post": map[string]any{
						"id": 123, "name": "Test Post", "ap_id": "https://lemmy.world/post/123",
						"published": "2026-06-15T00:00:00.000000", "nsfw": false,
					},
					"community": map[string]any{"name": "technology", "title": "Technology"},
					"creator":   map[string]any{"name": "user1", "display_name": "User 1"},
					"counts":    map[string]any{"score": 100, "comments": 25},
				},
			},
		})
	}))
	defer ts.Close()

	cfg := noRateConfig()
	cfg.BaseURL = ts.URL
	c := NewClient(cfg)

	posts, err := c.ListPosts(context.Background(), "Active", "All", 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(posts) != 1 {
		t.Fatalf("len(posts) = %d, want 1", len(posts))
	}
	p := posts[0]
	if p.ID != 123 {
		t.Errorf("ID = %d, want 123", p.ID)
	}
	if p.Title != "Test Post" {
		t.Errorf("Title = %q, want Test Post", p.Title)
	}
	if p.Community != "technology" {
		t.Errorf("Community = %q, want technology", p.Community)
	}
	if p.Author != "User 1" {
		t.Errorf("Author = %q, want User 1", p.Author)
	}
	if p.Score != 100 {
		t.Errorf("Score = %d, want 100", p.Score)
	}
	if p.Comments != 25 {
		t.Errorf("Comments = %d, want 25", p.Comments)
	}
	if p.PostURL != "https://lemmy.world/post/123" {
		t.Errorf("PostURL = %q", p.PostURL)
	}
}

func TestListCommunities(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"communities": []map[string]any{
				{
					"community": map[string]any{
						"id": 5, "name": "technology", "title": "Technology",
						"description": "Tech stuff", "actor_id": "https://lemmy.world/c/technology",
						"published": "2023-01-01T00:00:00.000000",
					},
					"counts": map[string]any{"subscribers": 5000, "posts": 1000, "comments": 3000},
				},
			},
		})
	}))
	defer ts.Close()

	cfg := noRateConfig()
	cfg.BaseURL = ts.URL
	c := NewClient(cfg)

	communities, err := c.ListCommunities(context.Background(), "Active", "All", 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(communities) != 1 {
		t.Fatalf("len(communities) = %d, want 1", len(communities))
	}
	comm := communities[0]
	if comm.ID != 5 {
		t.Errorf("ID = %d, want 5", comm.ID)
	}
	if comm.Name != "technology" {
		t.Errorf("Name = %q, want technology", comm.Name)
	}
	if comm.Subscribers != 5000 {
		t.Errorf("Subscribers = %d, want 5000", comm.Subscribers)
	}
}

func TestListComments(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("post_id") != "42" {
			t.Errorf("post_id = %q, want 42", r.URL.Query().Get("post_id"))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"comments": []map[string]any{
				{
					"comment": map[string]any{
						"id": 999, "content": "Great post!", "published": "2026-06-15T01:00:00.000000",
						"post_id": 42,
					},
					"creator": map[string]any{"name": "commenter", "display_name": "A Commenter"},
					"counts":  map[string]any{"score": 15},
				},
			},
		})
	}))
	defer ts.Close()

	cfg := noRateConfig()
	cfg.BaseURL = ts.URL
	c := NewClient(cfg)

	comments, err := c.ListComments(context.Background(), 42, 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(comments) != 1 {
		t.Fatalf("len(comments) = %d, want 1", len(comments))
	}
	cmt := comments[0]
	if cmt.ID != 999 {
		t.Errorf("ID = %d, want 999", cmt.ID)
	}
	if cmt.Content != "Great post!" {
		t.Errorf("Content = %q, want Great post!", cmt.Content)
	}
	if cmt.Author != "A Commenter" {
		t.Errorf("Author = %q, want A Commenter", cmt.Author)
	}
	if cmt.Score != 15 {
		t.Errorf("Score = %d, want 15", cmt.Score)
	}
	if cmt.PostID != 42 {
		t.Errorf("PostID = %d, want 42", cmt.PostID)
	}
}

func TestSearch(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("q") != "programming" {
			t.Errorf("q = %q, want programming", r.URL.Query().Get("q"))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"posts": []map[string]any{
				{
					"post":      map[string]any{"id": 1, "name": "Learn programming", "ap_id": "https://lemmy.world/post/1"},
					"community": map[string]any{"name": "programming", "title": "Programming"},
					"creator":   map[string]any{"name": "dev", "display_name": ""},
					"counts":    map[string]any{"score": 50, "comments": 10},
				},
			},
			"communities": []map[string]any{},
			"comments":    []map[string]any{},
		})
	}))
	defer ts.Close()

	cfg := noRateConfig()
	cfg.BaseURL = ts.URL
	c := NewClient(cfg)

	resp, err := c.Search(context.Background(), "programming", "Posts", "TopAll", 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Posts) != 1 {
		t.Fatalf("len(posts) = %d, want 1", len(resp.Posts))
	}
	if resp.Posts[0].Post.Name != "Learn programming" {
		t.Errorf("post name = %q", resp.Posts[0].Post.Name)
	}
}

func TestGetSite(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"site_view": map[string]any{
				"site": map[string]any{
					"name":        "lemmy.world",
					"description": "Lemmy World instance",
				},
				"counts": map[string]any{
					"users":       193000,
					"posts":       714000,
					"comments":    6600000,
					"communities": 15000,
				},
			},
			"version": "0.19.5",
		})
	}))
	defer ts.Close()

	cfg := noRateConfig()
	cfg.BaseURL = ts.URL
	c := NewClient(cfg)

	site, err := c.GetSite(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if site.Name != "lemmy.world" {
		t.Errorf("Name = %q, want lemmy.world", site.Name)
	}
	if site.Version != "0.19.5" {
		t.Errorf("Version = %q, want 0.19.5", site.Version)
	}
	if site.Users != 193000 {
		t.Errorf("Users = %d, want 193000", site.Users)
	}
}

func TestGetRetriesOn503(t *testing.T) {
	var hits int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if hits < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"posts": []map[string]any{},
		})
	}))
	defer ts.Close()

	cfg := DefaultConfig()
	cfg.Rate = 0
	cfg.Retries = 5
	cfg.BaseURL = ts.URL
	c := NewClient(cfg)

	_, err := c.ListPosts(context.Background(), "", "", 5)
	if err != nil {
		t.Fatal(err)
	}
	if hits != 3 {
		t.Errorf("server saw %d hits, want 3", hits)
	}
}

func TestContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	cfg := DefaultConfig()
	cfg.Rate = 0
	cfg.Retries = 0
	c := NewClient(cfg)

	_, err := c.ListPosts(ctx, "", "", 5)
	if err == nil {
		t.Error("ListPosts with cancelled context returned nil error")
	}
}

func TestPostRoundTrip(t *testing.T) {
	p := Post{
		ID:        123,
		Title:     "Test",
		Community: "tech",
		Author:    "alice",
		Score:     42,
		Comments:  7,
		NSFW:      false,
		PostURL:   "https://lemmy.world/post/123",
	}
	b, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	var got Post
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	if got.ID != p.ID || got.Title != p.Title || got.Score != p.Score {
		t.Errorf("round-trip mismatch: got %+v", got)
	}
}

func TestCommunityRoundTrip(t *testing.T) {
	c := Community{
		ID:          5,
		Name:        "technology",
		Title:       "Technology",
		Subscribers: 10000,
	}
	b, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	var got Community
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	if got.ID != c.ID || got.Name != c.Name || got.Subscribers != c.Subscribers {
		t.Errorf("round-trip mismatch: got %+v", got)
	}
}

func TestCommentRoundTrip(t *testing.T) {
	c := Comment{
		ID:      999,
		Content: "Hello world",
		Author:  "bob",
		Score:   5,
		PostID:  42,
	}
	b, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	var got Comment
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	if got.ID != c.ID || got.Content != c.Content || got.PostID != c.PostID {
		t.Errorf("round-trip mismatch: got %+v", got)
	}
}

func TestHostConstant(t *testing.T) {
	if Host != "lemmy.world" {
		t.Errorf("Host = %q, want lemmy.world", Host)
	}
}
