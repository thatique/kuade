package policy

type Args struct {
	AccountName     string              `json:"account"`
	Action          Action              `json:"action"`
	ResourceName    string              `json:"resource"`
	ConditionValues map[string][]string `json:"conditions"`
	IsOwner         bool                `json:"owner"`
	ObjectName      string              `json:"object"`
}
