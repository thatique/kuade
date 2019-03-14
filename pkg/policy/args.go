package policy

type Args struct {
	AccountName     string              `json:"account" bson:"account"`
	Action          Action              `json:"action" bson:"action"`
	ResourceName    string              `json:"resource" bson:"resource"`
	ConditionValues map[string][]string `json:"conditions" bson:"conditions"`
	IsOwner         bool                `json:"owner" bson:"owner"`
	ObjectName      string              `json:"object" bson:"object"`
}
