package git

import (
	"github.com/go-git/go-git/v5"
)

func Clone(repoURL, path string) error {
	_, err := git.PlainClone(path, false, &git.CloneOptions{
		URL:      repoURL,
		Progress: nil,
	})

	if err != nil {
		return err
	}

	return nil
}
