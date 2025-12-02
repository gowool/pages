package pages

type Status int8

const (
	DraftStatus Status = iota
	PublishStatus
	PrivateStatus
)

func (s Status) String() string {
	switch s {
	case DraftStatus:
		return "draft"
	case PublishStatus:
		return "publish"
	case PrivateStatus:
		return "private"
	}

	return "unknown"
}
