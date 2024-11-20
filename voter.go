package pages

import (
	"context"
	neturl "net/url"

	"github.com/gowool/pages/model"
)

// Voter represents an interface for determining whether a node is current.
type Voter interface {
	// MatchNode checks whether a node is current.
	//
	// If the voter is not able to determine a result,
	// it should return nil to let other voters do the job.
	MatchNode(ctx context.Context, node *model.Node) *bool
}

type URLVoter struct{}

func NewURLVoter() URLVoter {
	return URLVoter{}
}

func (URLVoter) MatchNode(ctx context.Context, node *model.Node) *bool {
	if u, ok := ctx.Value("url").(neturl.URL); ok && u.Path == node.URI {
		return &ok
	}
	return nil
}

// Matcher represents an interface for matching nodes.
// It provides methods for checking whether a node is current or an ancestor.
type Matcher interface {
	// IsCurrent checks whether a node is current
	IsCurrent(ctx context.Context, node *model.Node) bool

	// IsAncestor checks whether a node is the ancestor of a current node
	IsAncestor(ctx context.Context, node *model.Node) bool
}

type DefaultMatcher struct {
	voters []Voter
}

func NewDefaultMatcher(voters ...Voter) DefaultMatcher {
	return DefaultMatcher{voters: voters}
}

func (m DefaultMatcher) IsCurrent(ctx context.Context, node *model.Node) bool {
	if node == nil {
		return false
	}

	if node.Current {
		return true
	}

	for _, voter := range m.voters {
		if current := voter.MatchNode(ctx, node); current != nil {
			node.Current = *current
			return *current
		}
	}
	return false
}

func (m DefaultMatcher) IsAncestor(ctx context.Context, node *model.Node) bool {
	if node == nil {
		return false
	}

	if node.Ancestor {
		return true
	}

	for _, child := range node.Children {
		if m.IsCurrent(ctx, child) || m.IsAncestor(ctx, child) {
			return true
		}
	}
	return false
}
