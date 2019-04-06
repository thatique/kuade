package service

import (
	"github.com/thatique/kuade/pkg/auth/user"
	"github.com/thatique/kuade/pkg/policy"
)

type ServiceAction interface {
	GetPolicyArgs(user user.Info, service *Service) policy.Args
}
