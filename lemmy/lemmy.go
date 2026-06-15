// Package lemmy is the library behind the lemmy command line:
// the HTTP client, request shaping, and the typed data models for the
// Lemmy federated forum API at lemmy.world.
//
// The Lemmy REST API is open for public read-only data: no API key, no auth
// required. This package wraps the v3 API with a rate-limited client that the
// kit operations consume.
package lemmy

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// Host is the default Lemmy instance this client talks to.
const Host = "lemmy.world"

const defaultBaseURL = "https://lemmy.world/api/v3"

// DefaultUserAgent identifies the client to lemmy.world honestly.
const DefaultUserAgent = "lemmy-cli/0.1 (tamnd87@gmail.com)"

// Config holds constructor parameters for Client.
type Config struct {
	BaseURL   string
	UserAgent string
	Rate      time.Duration
	Timeout   time.Duration
	Retries   int
}

// DefaultConfig returns sensible defaults for lemmy.world.
func DefaultConfig() Config {
	return Config{
		BaseURL:   defaultBaseURL,
		UserAgent: DefaultUserAgent,
		Rate:      500 * time.Millisecond,
		Timeout:   15 * time.Second,
		Retries:   3,
	}
}

// Client is a rate-limited HTTP client for the Lemmy v3 API.
type Client struct {
	cfg  Config
	http *http.Client
	last time.Time
}

// NewClient returns a Client configured with cfg.
func NewClient(cfg Config) *Client {
	return &Client{
		cfg:  cfg,
		http: &http.Client{Timeout: cfg.Timeout},
	}
}

// --- output types ---

// Post is a single Lemmy post record.
type Post struct {
	ID        int    `json:"id" kit:"id"`
	Title     string `json:"title"`
	URL       string `json:"url"`
	Body      string `json:"body"`
	Community string `json:"community"`
	Author    string `json:"author"`
	Score     int    `json:"score"`
	Comments  int    `json:"comments"`
	Published string `json:"published"`
	NSFW      bool   `json:"nsfw"`
	PostURL   string `json:"post_url"`
}

// Community is a single Lemmy community record.
type Community struct {
	ID          int    `json:"id" kit:"id"`
	Name        string `json:"name"`
	Title       string `json:"title"`
	Description string `json:"description"`
	ActorID     string `json:"actor_id"`
	Subscribers int    `json:"subscribers"`
	Posts       int    `json:"posts"`
	Comments    int    `json:"comments"`
	Published   string `json:"published"`
}

// Comment is a single Lemmy comment record.
type Comment struct {
	ID        int    `json:"id" kit:"id"`
	Content   string `json:"content"`
	Author    string `json:"author"`
	Score     int    `json:"score"`
	Published string `json:"published"`
	PostID    int    `json:"post_id"`
}

// Site is instance-level statistics for a Lemmy instance.
type Site struct {
	Name        string `json:"name" kit:"id"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Users       int    `json:"users"`
	Posts       int    `json:"posts"`
	Comments    int    `json:"comments"`
	Communities int    `json:"communities"`
}

// --- wire types (full nested JSON from Lemmy API) ---

type wirePost struct {
	Post struct {
		ID        int    `json:"id"`
		Name      string `json:"name"`
		Body      string `json:"body"`
		URL       string `json:"url"`
		ApID      string `json:"ap_id"`
		Published string `json:"published"`
		NSFW      bool   `json:"nsfw"`
	} `json:"post"`
	Community struct {
		Name  string `json:"name"`
		Title string `json:"title"`
	} `json:"community"`
	Creator struct {
		Name        string `json:"name"`
		DisplayName string `json:"display_name"`
	} `json:"creator"`
	Counts struct {
		Score    int `json:"score"`
		Comments int `json:"comments"`
	} `json:"counts"`
}

type wireCommunity struct {
	Community struct {
		ID          int    `json:"id"`
		Name        string `json:"name"`
		Title       string `json:"title"`
		Description string `json:"description"`
		ActorID     string `json:"actor_id"`
		Published   string `json:"published"`
	} `json:"community"`
	Counts struct {
		Subscribers int `json:"subscribers"`
		Posts       int `json:"posts"`
		Comments    int `json:"comments"`
	} `json:"counts"`
}

type wireComment struct {
	Comment struct {
		ID        int    `json:"id"`
		Content   string `json:"content"`
		Published string `json:"published"`
		PostID    int    `json:"post_id"`
	} `json:"comment"`
	Creator struct {
		Name        string `json:"name"`
		DisplayName string `json:"display_name"`
	} `json:"creator"`
	Counts struct {
		Score int `json:"score"`
	} `json:"counts"`
}

type wirePostsResp struct {
	Posts []wirePost `json:"posts"`
}

type wireCommunityListResp struct {
	Communities []wireCommunity `json:"communities"`
}

type wireCommentsResp struct {
	Comments []wireComment `json:"comments"`
}

type wireSearchResp struct {
	Posts       []wirePost      `json:"posts"`
	Communities []wireCommunity `json:"communities"`
	Comments    []wireComment   `json:"comments"`
}

type wireSiteResp struct {
	SiteView struct {
		Site struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		} `json:"site"`
		Counts struct {
			Users       int `json:"users"`
			Posts       int `json:"posts"`
			Comments    int `json:"comments"`
			Communities int `json:"communities"`
		} `json:"counts"`
	} `json:"site_view"`
	Version string `json:"version"`
}

// --- converters ---

func toPost(w wirePost) Post {
	author := w.Creator.Name
	if w.Creator.DisplayName != "" {
		author = w.Creator.DisplayName
	}
	return Post{
		ID:        w.Post.ID,
		Title:     w.Post.Name,
		URL:       w.Post.URL,
		Body:      w.Post.Body,
		Community: w.Community.Name,
		Author:    author,
		Score:     w.Counts.Score,
		Comments:  w.Counts.Comments,
		Published: w.Post.Published,
		NSFW:      w.Post.NSFW,
		PostURL:   w.Post.ApID,
	}
}

func toCommunity(w wireCommunity) Community {
	return Community{
		ID:          w.Community.ID,
		Name:        w.Community.Name,
		Title:       w.Community.Title,
		Description: w.Community.Description,
		ActorID:     w.Community.ActorID,
		Subscribers: w.Counts.Subscribers,
		Posts:       w.Counts.Posts,
		Comments:    w.Counts.Comments,
		Published:   w.Community.Published,
	}
}

func toComment(w wireComment) Comment {
	author := w.Creator.Name
	if w.Creator.DisplayName != "" {
		author = w.Creator.DisplayName
	}
	return Comment{
		ID:        w.Comment.ID,
		Content:   w.Comment.Content,
		Author:    author,
		Score:     w.Counts.Score,
		Published: w.Comment.Published,
		PostID:    w.Comment.PostID,
	}
}

// --- API methods ---

// ListPosts fetches posts from the Lemmy API.
// sort: Active, Hot, New, TopDay, TopWeek, TopMonth (default: Active)
// listType: All, Local, Subscribed (default: All)
func (c *Client) ListPosts(ctx context.Context, sort, listType string, limit int) ([]Post, error) {
	if sort == "" {
		sort = "Active"
	}
	if listType == "" {
		listType = "All"
	}
	if limit <= 0 {
		limit = 20
	}
	q := url.Values{}
	q.Set("sort", sort)
	q.Set("type_", listType)
	q.Set("limit", strconv.Itoa(limit))

	var resp wirePostsResp
	if err := c.getJSON(ctx, c.cfg.BaseURL+"/post/list?"+q.Encode(), &resp); err != nil {
		return nil, err
	}
	out := make([]Post, len(resp.Posts))
	for i, w := range resp.Posts {
		out[i] = toPost(w)
	}
	return out, nil
}

// ListCommunities fetches communities from the Lemmy API.
// sort: Active, Hot, New, TopDay, TopMonth, TopYear, TopAll (default: Active)
// listType: All, Local (default: All)
func (c *Client) ListCommunities(ctx context.Context, sort, listType string, limit int) ([]Community, error) {
	if sort == "" {
		sort = "Active"
	}
	if listType == "" {
		listType = "All"
	}
	if limit <= 0 {
		limit = 20
	}
	q := url.Values{}
	q.Set("sort", sort)
	q.Set("type_", listType)
	q.Set("limit", strconv.Itoa(limit))

	var resp wireCommunityListResp
	if err := c.getJSON(ctx, c.cfg.BaseURL+"/community/list?"+q.Encode(), &resp); err != nil {
		return nil, err
	}
	out := make([]Community, len(resp.Communities))
	for i, w := range resp.Communities {
		out[i] = toCommunity(w)
	}
	return out, nil
}

// ListComments fetches comments on a post.
func (c *Client) ListComments(ctx context.Context, postID, limit int) ([]Comment, error) {
	if limit <= 0 {
		limit = 20
	}
	q := url.Values{}
	q.Set("post_id", strconv.Itoa(postID))
	q.Set("limit", strconv.Itoa(limit))

	var resp wireCommentsResp
	if err := c.getJSON(ctx, c.cfg.BaseURL+"/comment/list?"+q.Encode(), &resp); err != nil {
		return nil, err
	}
	out := make([]Comment, len(resp.Comments))
	for i, w := range resp.Comments {
		out[i] = toComment(w)
	}
	return out, nil
}

// Search searches for posts, communities, or comments.
// searchType: Posts, Communities, Comments, Users (default: Posts)
// sort: Active, Hot, New, TopAll (default: TopAll)
func (c *Client) Search(ctx context.Context, query, searchType, sort string, limit int) (*wireSearchResp, error) {
	if searchType == "" {
		searchType = "Posts"
	}
	if sort == "" {
		sort = "TopAll"
	}
	if limit <= 0 {
		limit = 20
	}
	q := url.Values{}
	q.Set("q", query)
	q.Set("type_", searchType)
	q.Set("sort", sort)
	q.Set("limit", strconv.Itoa(limit))

	var resp wireSearchResp
	if err := c.getJSON(ctx, c.cfg.BaseURL+"/search?"+q.Encode(), &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetSite fetches instance statistics.
func (c *Client) GetSite(ctx context.Context) (*Site, error) {
	var resp wireSiteResp
	if err := c.getJSON(ctx, c.cfg.BaseURL+"/site", &resp); err != nil {
		return nil, err
	}
	return &Site{
		Name:        resp.SiteView.Site.Name,
		Description: resp.SiteView.Site.Description,
		Version:     resp.Version,
		Users:       resp.SiteView.Counts.Users,
		Posts:       resp.SiteView.Counts.Posts,
		Comments:    resp.SiteView.Counts.Comments,
		Communities: resp.SiteView.Counts.Communities,
	}, nil
}

// --- HTTP helpers ---

func (c *Client) getJSON(ctx context.Context, rawURL string, v any) error {
	body, err := c.get(ctx, rawURL)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, v); err != nil {
		return fmt.Errorf("decode %s: %w", rawURL, err)
	}
	return nil
}

func (c *Client) get(ctx context.Context, rawURL string) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt <= c.cfg.Retries; attempt++ {
		if attempt > 0 {
			wait := time.Duration(attempt) * 500 * time.Millisecond
			if wait > 5*time.Second {
				wait = 5 * time.Second
			}
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(wait):
			}
		}
		body, retry, err := c.do(ctx, rawURL)
		if err == nil {
			return body, nil
		}
		lastErr = err
		if !retry {
			return nil, err
		}
	}
	return nil, fmt.Errorf("get %s: %w", rawURL, lastErr)
}

func (c *Client) do(ctx context.Context, rawURL string) ([]byte, bool, error) {
	c.pace()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("User-Agent", c.cfg.UserAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, true, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
		return nil, true, fmt.Errorf("http %d", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("http %d", resp.StatusCode)
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, true, err
	}
	return b, false, nil
}

// pace blocks until at least Rate has passed since the last request.
func (c *Client) pace() {
	if c.cfg.Rate <= 0 {
		return
	}
	if wait := c.cfg.Rate - time.Since(c.last); wait > 0 {
		time.Sleep(wait)
	}
	c.last = time.Now()
}
