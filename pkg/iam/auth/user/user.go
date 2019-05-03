package user

const (
	GroupAnonymous  = "anonymous"
	GroupIndividual = "individual"
	GroupVendor     = "vendor"
	GroupStaff      = "staff"
	GroupSuper      = "sudoer"
)

const (
	Anonymous = "anonymous"
)

type Info interface {
	// GetUserName returns the name that uniquely identifies this user among all
	// other active users. This can be an email or username.
	GetUsername() string

	// GetUID returns a unique value for a particular user that will change
	// if the user is removed from the system and another user is added with
	// the same name.
	GetUID() string

	// GetGroups returns the names of the groups the user is a member of
	GetGroups() []string

	// GetMetadata returns any additional information that the authenticator
	// thought was interesting.  One example would be scopes on a token.
	// Keys in this map should be namespaced to the authenticator or
	// authenticator/authorizer pair making use of them.
	GetMetadata() map[string][]string
}

// DefaultInfo provides a simple user information exchange object
// for components that implement the UserInfo interface.
type DefaultInfo struct {
	Name     string
	UID      string
	Groups   []string
	Metadata map[string][]string
}

func (i *DefaultInfo) GetUserName() string {
	return i.Name
}

func (i *DefaultInfo) GetUID() string {
	return i.UID
}

func (i *DefaultInfo) GetGroups() []string {
	return i.Groups
}

func (i *DefaultInfo) GetMetadata() map[string][]string {
	return i.Metadata
}
