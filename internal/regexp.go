package internal

import (
	"fmt"

	"github.com/dlclark/regexp2"
)

const (
	reOptions    = regexp2.IgnoreCase & regexp2.RE2
	rePathExpr   = "^(%s)(/.*|$)"
	reNoPathExpr = "^()(/.*|$)"
)

var ReNoPath = regexp2.MustCompile(reNoPathExpr, reOptions)

func Regexp(expr string) (*regexp2.Regexp, bool) {
	re, err := regexp2.Compile(expr, reOptions)
	return re, err == nil
}

func RegexpPath(path string) (*regexp2.Regexp, error) {
	return regexp2.Compile(fmt.Sprintf(rePathExpr, path), reOptions)
}
