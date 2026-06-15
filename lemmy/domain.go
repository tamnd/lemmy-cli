package lemmy

import (
	"context"
	"fmt"
	"strconv"

	"github.com/tamnd/any-cli/kit"
	"github.com/tamnd/any-cli/kit/errs"
)

// domain.go registers the lemmy kit Domain so a blank import in a multi-domain
// host (ant) enables the driver:
//
//	import _ "github.com/tamnd/lemmy-cli/lemmy"
//
// The Domain also builds the standalone lemmy binary via cli.NewApp.
func init() { kit.Register(Domain{}) }

// Domain is the Lemmy driver. It carries no state; the per-run client is built
// by the factory Register hands kit.
type Domain struct{}

// Info describes the scheme and the identity the single-site binary inherits.
func (Domain) Info() kit.DomainInfo {
	return kit.DomainInfo{
		Scheme: "lemmy",
		Hosts:  []string{Host},
		Identity: kit.Identity{
			Binary: "lemmy",
			Short:  "Read public Lemmy federated forum data",
			Long: `A command line for Lemmy, the federated link aggregator.

lemmy reads public data from lemmy.world over plain HTTPS, shapes it into
clean records, and prints output that pipes into the rest of your tools. No API
key, nothing to run alongside it.`,
			Site: Host,
			Repo: "https://github.com/tamnd/lemmy-cli",
		},
	}
}

// Register installs the client factory and every operation onto app.
func (Domain) Register(app *kit.App) {
	app.SetClient(newClient)

	kit.Handle(app, kit.OpMeta{Name: "posts", Group: "read", List: true,
		Summary: "List posts (--sort Active|Hot|New|TopDay --type All|Local)"}, listPosts)

	kit.Handle(app, kit.OpMeta{Name: "communities", Group: "read", List: true,
		Summary: "List communities by activity"}, listCommunities)

	kit.Handle(app, kit.OpMeta{Name: "comments", Group: "read", List: true,
		Summary: "List comments on a post",
		Args:    []kit.Arg{{Name: "post-id", Help: "post ID"}}}, listComments)

	kit.Handle(app, kit.OpMeta{Name: "search", Group: "read", List: true,
		Summary: "Search posts and communities",
		Args:    []kit.Arg{{Name: "query", Help: "search query"}}}, search)

	kit.Handle(app, kit.OpMeta{Name: "site", Group: "read", Single: true,
		Summary: "Show Lemmy instance statistics"}, getSite)
}

// newClient builds a Client from the resolved kit Config.
func newClient(_ context.Context, cfg kit.Config) (any, error) {
	c := DefaultConfig()
	if cfg.UserAgent != "" {
		c.UserAgent = cfg.UserAgent
	}
	if cfg.Rate > 0 {
		c.Rate = cfg.Rate
	}
	if cfg.Retries > 0 {
		c.Retries = cfg.Retries
	}
	if cfg.Timeout > 0 {
		c.Timeout = cfg.Timeout
	}
	return NewClient(c), nil
}

// --- input structs ---

type postsInput struct {
	Sort   string  `kit:"flag" help:"sort order (Active, Hot, New, TopDay, TopWeek)"`
	Type   string  `kit:"flag" help:"type (All, Local)"`
	Limit  int     `kit:"flag,inherit" help:"max results"`
	Client *Client `kit:"inject"`
}

type communitiesInput struct {
	Sort   string  `kit:"flag" help:"sort order (Active, Hot, New)"`
	Type   string  `kit:"flag" help:"type (All, Local)"`
	Limit  int     `kit:"flag,inherit" help:"max results"`
	Client *Client `kit:"inject"`
}

type commentsInput struct {
	PostID string  `kit:"arg" help:"post ID"`
	Limit  int     `kit:"flag,inherit" help:"max results"`
	Client *Client `kit:"inject"`
}

type searchInput struct {
	Query  string  `kit:"arg" help:"search query"`
	Type   string  `kit:"flag" help:"type (Posts, Communities, Comments)"`
	Sort   string  `kit:"flag" help:"sort order (Active, Hot, New, TopAll)"`
	Limit  int     `kit:"flag,inherit" help:"max results"`
	Client *Client `kit:"inject"`
}

type siteInput struct {
	Client *Client `kit:"inject"`
}

// --- handlers ---

func listPosts(ctx context.Context, in postsInput, emit func(Post) error) error {
	posts, err := in.Client.ListPosts(ctx, in.Sort, in.Type, in.Limit)
	if err != nil {
		return err
	}
	for _, p := range posts {
		if err := emit(p); err != nil {
			return err
		}
	}
	return nil
}

func listCommunities(ctx context.Context, in communitiesInput, emit func(Community) error) error {
	communities, err := in.Client.ListCommunities(ctx, in.Sort, in.Type, in.Limit)
	if err != nil {
		return err
	}
	for _, c := range communities {
		if err := emit(c); err != nil {
			return err
		}
	}
	return nil
}

func listComments(ctx context.Context, in commentsInput, emit func(Comment) error) error {
	postID, err := strconv.Atoi(in.PostID)
	if err != nil {
		return errs.Usage("post-id must be an integer, got %q", in.PostID)
	}
	comments, err := in.Client.ListComments(ctx, postID, in.Limit)
	if err != nil {
		return err
	}
	for _, c := range comments {
		if err := emit(c); err != nil {
			return err
		}
	}
	return nil
}

func search(ctx context.Context, in searchInput, emit func(Post) error) error {
	resp, err := in.Client.Search(ctx, in.Query, in.Type, in.Sort, in.Limit)
	if err != nil {
		return err
	}
	for _, w := range resp.Posts {
		if err := emit(toPost(w)); err != nil {
			return err
		}
	}
	return nil
}

func getSite(ctx context.Context, in siteInput, emit func(*Site) error) error {
	s, err := in.Client.GetSite(ctx)
	if err != nil {
		return err
	}
	return emit(s)
}

// --- Resolver ---

// Classify turns any accepted input into the canonical (uriType, id).
func (Domain) Classify(input string) (uriType, id string, err error) {
	if input == "" {
		return "", "", errs.Usage("lemmy: empty input")
	}
	// numeric → post id
	if _, err := strconv.Atoi(input); err == nil {
		return "post", input, nil
	}
	return "", "", errs.Usage("lemmy: unrecognized reference: %q", input)
}

// Locate returns the canonical URL for a (uriType, id).
func (Domain) Locate(uriType, id string) (string, error) {
	switch uriType {
	case "post":
		n, err := strconv.Atoi(id)
		if err != nil {
			return "", errs.Usage("lemmy: post id must be numeric, got %q", id)
		}
		return fmt.Sprintf("https://%s/post/%d", Host, n), nil
	default:
		return "", errs.Usage("lemmy has no resource type %q", uriType)
	}
}
