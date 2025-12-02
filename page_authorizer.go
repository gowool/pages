package pages

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

type PageAuthorizer[T Resolver] interface {
	Authorize(e T, action PageAction) (Decision, error)
}

type DenyPageAuthorizer[T Resolver] struct{}

func (DenyPageAuthorizer[T]) Authorize(T, PageAction) (Decision, error) {
	return Deny, nil
}
