package github

// PullRequest represents a GitHub pull request.
type PullRequest struct {
	Owner  string
	Repo   string
	Number int
}

// Review reviews a pull request.
func (pr *PullRequest) Review() error {
	// Review the pull request.
	return nil
}
