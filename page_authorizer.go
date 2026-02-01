package pages

import "context"

type PageAction int8

const (
	ViewDraftPage PageAction = iota + 1
	ViewPrivatePage
	CreatePage
)

func (a PageAction) String() string {
	switch a {
	case ViewDraftPage:
		return "VIEW_DRAFT_PAGE"
	case ViewPrivatePage:
		return "VIEW_PRIVATE_PAGE"
	case CreatePage:
		return "CREATE_PAGE"
	default:
		return "UNKNOWN"
	}
}

type Decision int8

const (
	Deny Decision = iota
	Allow
)

func (d Decision) String() string {
	switch d {
	case Deny:
		return "deny"
	case Allow:
		return "allow"
	default:
		return "unknown"
	}
}

type PageAuthorizer interface {
	Authorize(ctx context.Context, action PageAction) (Decision, error)
}

type DenyPageAuthorizer struct{}

func (DenyPageAuthorizer) Authorize(context.Context, PageAction) (Decision, error) {
	return Deny, nil
}
