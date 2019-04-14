package authorizer

import (
	"github.com/thatique/kuade/pkg/arn"
	"github.com/thatique/kuade/pkg/iam/auth/user"
)

// Action being performed
type Action string

// Args is argument to be passed to Authorizer
type Args struct {
	User   user.Info
	Action Action
	ARN    arn.ARN
	IsOwner bool
	// the specific metadata
	Metadata map[string][]string
}

// Authorizer Authorize Args
type Authorizer interface {
	Authorize(args Args) (authorized Decision, reason string, err error)
}

// AuthorizerFunc function to implement Auhorizer interface
type AuthorizerFunc func(args Args) (Decision, string, error)

// Authorize implments Authorizer interface
func (f AuthorizerFunc) Authorize(args Args) (authorized Decision, reason string, err error) {
	return f(args)
}

// Decision is result running Authorizer, it's either DecisionDeny, DecisionAllow
// or DecisionNoOpinion
type Decision int32

const (
	// DecisionDeny Deny the request
	DecisionDeny Decision = iota
	// DecisionAllow Allow the request
	DecisionAllow
	// DecisionNoOpinion means that an authorizer has no opinion on whether
	// to allow or deny an action.
	DecisionNoOpinion
)
