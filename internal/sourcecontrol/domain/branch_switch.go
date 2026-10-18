package domain

// BranchSwitchPlan binds a reviewed stopped-checkout transition to exact refs.
// Restart approval is deliberately separate from checkout approval.
type BranchSwitchPlan struct {
	ProjectID     string `json:"projectId"`
	Root          string `json:"root"`
	FromBranch    string `json:"fromBranch"`
	FromHead      string `json:"fromHead"`
	ToBranch      string `json:"toBranch"`
	ToHead        string `json:"toHead"`
	Risk          string `json:"risk"`
	ReviewRestart bool   `json:"reviewRestart"`
}
