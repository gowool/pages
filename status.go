package pages

type Status int8

const (
	Draft Status = iota
	Published
)

func (s Status) String() string {
	switch s {
	case Draft:
		return "draft"
	case Published:
		return "published"
	default:
		return "unknown"
	}
}

func StatusFromString(s string) Status {
	switch s {
	case "published":
		return Published
	default:
		return Draft
	}
}
