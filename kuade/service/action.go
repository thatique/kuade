package service

import (
	"github.com/thatique/kuade/kuade/auth"
	"github.com/thatique/kuade/pkg/policy"
)

type ServiceAction interface {
	GetPolicyArgs(user *auth.User, service *Service) policy.Args
}
