package main

import (
	"context"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/ghodss/yaml"
	"github.com/google/go-github/v29/github"
)

type Commit struct {
	at        time.Time
	sha       string
	committer string
	message   string
	files     map[string]string
}

// When program starts we want to synch all files in git to openshift.
// This function fakes a commit with all config files from git repo.
// Then we can use this fake commit in the normal way to synch a commit to openshift.
func (s *server) getFakeInitialCommit(ctx context.Context) ([]Commit, error) {
	allFiles, err := s.getDir(ctx, s.confDir)
	if err != nil {
		return nil, err
	}
	fileMap := make(map[string]string)
	for _, filename := range allFiles {
		fileMap[filename] = ""
	}
	fakeCommit := Commit{
		at:        time.Now().Add(-60 * time.Second),
		committer: "-",
		message:   "Initial Load",
		files:     fileMap,
	}
	return []Commit{fakeCommit}, nil
}

// Get all commits after given timestamp
func (s *server) getNewCommits(ctx context.Context, since time.Time) ([]Commit, error) {
	owner, repo := getOwnerAndRepo(s.gitRepo)
	gitCommits, _, err := s.gitClient.Repositories.ListCommits(ctx, owner, repo,
		&github.CommitsListOptions{SHA: "master", Since: since})
	if err != nil {
		return nil, err
	}

	commits := []Commit{}

	for commitIdx := range gitCommits {
		gitCommit, _, err := s.gitClient.Repositories.GetCommit(ctx, owner, repo, gitCommits[commitIdx].GetSHA())
		if err != nil {
			return nil, err
		}
		commit := Commit{
			at:        gitCommit.Commit.Committer.GetDate(),
			sha:       gitCommit.GetSHA(),
			committer: gitCommit.Commit.Committer.GetName(),
			message:   gitCommit.Commit.GetMessage(),
			files:     make(map[string]string),
		}
		for _, file := range gitCommit.Files {
			filename := file.GetFilename()
			if strings.HasPrefix(filename, s.confDir) {
				// Ignore files not in /confDir/...
				commit.files[file.GetFilename()] = ""
			}
		}
		commits = append(commits, commit)
	}
	return commits, nil
}

// Returns content of a file. If file is yaml we convert it to json.
func (s *server) getFile(ctx context.Context, filename string, sha string) (string, error) {
	owner, repo := getOwnerAndRepo(s.gitRepo)
	fileContent, _, _, err := s.gitClient.Repositories.GetContents(ctx, owner, repo, filename, &github.RepositoryContentGetOptions{Ref: sha})
	if err != nil {
		return "", err
	}
	content, err := fileContent.GetContent()
	if err != nil {
		return "", err
	}

	// convert yaml to json
	if strings.HasSuffix(filename, "yaml") || strings.HasSuffix(filename, "yml") {
		jsonContent, err := yaml.YAMLToJSON([]byte(content))
		if err != nil {
			return "", err
		}
		content = string(jsonContent)
	}
	return content, nil
}

// Returns list of filenames in a dir. Includes subdirs recursivley.
func (s *server) getDir(ctx context.Context, dir string) ([]string, error) {
	log.Printf("dir=%s\n", dir)
	owner, repo := getOwnerAndRepo(s.gitRepo)
	_, fileContents, _, err := s.gitClient.Repositories.GetContents(ctx, owner, repo, dir, &github.RepositoryContentGetOptions{Ref: "master"})
	if err != nil {
		return nil, err
	}
	filenames := []string{}
	for _, f := range fileContents {
		switch f.GetType() {
		case "file":
			filenames = append(filenames, filepath.Join(dir, f.GetName()))
		case "dir":
			subfiles, err := s.getDir(ctx, filepath.Join(dir, f.GetName()))
			if err != nil {
				return nil, err
			}
			filenames = append(filenames, subfiles...)
		default:
			log.Printf("unknown fileContent type; %s\n", f.GetType())
		}
		log.Printf("DirContent: type=%s; %s\n", f.GetName(), f.GetType())
	}
	return filenames, nil
}

func getOwnerAndRepo(repoPath string) (string, string) {
	parts := strings.Split(repoPath, "/")
	return parts[0], parts[1]
}
