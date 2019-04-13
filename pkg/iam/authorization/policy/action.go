package iampolicy

import (
	"github.com/minio/minio/pkg/wildcard"
)

// Action is action taken by user
type Action string

// Match match an action with action pattern
func (action Action) Match(act2 Action) bool {
	return wildcard.Match(string(action), string(act2))
}
