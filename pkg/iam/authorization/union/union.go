package union

import (
	"strings"

	"github.com/thatique/kuade/pkg/iam/authorization/authorizer"
	"github.com/thatique/kuade/pkg/kerr"
)

// unionAuthzHandler authorizer against a chain of authorizer.Authorizer
type unionAuthzHandler []authorizer.Authorizer

func New(authorizationHandlers ...authorizer.Authorizer) authorizer.Authorizer {
	return unionAuthzHandler(authorizationHandlers)
}

func (authzHandler unionAuthzHandler) Authorize(a authorizer.Args) (authorizer.Decision, string, error) {
	var (
		errlist    []error
		reasonlist []string
	)

	for _, authz := range authzHandler {
		decision, reason, err := authz.Authorize(a)

		if err != nil {
			errlist = append(errlist, err)
		}
		if reason != "" {
			reasonlist = append(reasonlist, reason)
		}
		switch decision {
		case authorizer.DecisionAllow, authorizer.DecisionDeny:
			return decision, reason, err
		case authorizer.DecisionNoOpinion:
			// continue to the next authorizer
		}
	}

	return authorizer.DecisionNoOpinion, strings.Join(reasonlist, "\n"), kerr.NewAggregate(errlist)
}
