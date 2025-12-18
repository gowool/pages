package pages

type Visibility int8

const (
	Private Visibility = iota
	Public
)

func (v Visibility) String() string {
	switch v {
	case Private:
		return "private"
	case Public:
		return "public"
	default:
		return "unknown"
	}
}

func VisibilityFromString(s string) Visibility {
	switch s {
	case "public":
		return Public
	default:
		return Private
	}
}
