package github

import "context"

// EnsurePR creates a PR if the branch doesn't have one yet, or returns the existing PR number.
func EnsurePR(ctx context.Context, client Client, owner, repo string, input CreatePRInput, existingPR *int) (*int, string, error) {
	if existingPR != nil {
		pr, err := client.GetPR(ctx, owner, repo, *existingPR)
		if err != nil {
			return existingPR, "", err
		}
		return existingPR, pr.HTMLURL, nil
	}

	pr, err := client.CreatePR(ctx, input)
	if err != nil {
		return nil, "", err
	}
	return &pr.Number, pr.HTMLURL, nil
}

// UpdatePRBodyIfChanged updates a PR body only if it has actually changed.
func UpdatePRBodyIfChanged(ctx context.Context, client Client, owner, repo string, number int, newBody string) error {
	pr, err := client.GetPR(ctx, owner, repo, number)
	if err != nil {
		return err
	}
	if pr.Body == newBody {
		return nil
	}
	return client.UpdatePR(ctx, owner, repo, number, UpdatePRInput{Body: newBody})
}
