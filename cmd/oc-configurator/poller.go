package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

func (s *server) waitForPoll() {
	nextPoll := time.Now()

	for !s.isQuit {
		time.Sleep(time.Second)
		if time.Now().Before(nextPoll) {
			continue
		}

		log.Printf("running poll\n")
		if err := s.poll(); err != nil {
			log.Printf("poll failed; err=%+v\n", err)
		}
		nextPoll = time.Now().Add(60 * time.Second)
	}
	log.Printf("waitForPoller done\n")
}

func (s *server) poll() error {
	ctx := context.Background()
	var commits []Commit
	var err error

	if s.lastCommitAt.IsZero() {
		commits, err = s.getFakeInitialCommit(ctx)
	} else {
		commits, err = s.getNewCommits(ctx, s.lastCommitAt)
	}
	if err != nil {
		return err
	}

	for _, commit := range commits {
		if err := s.processCommit(ctx, &commit); err != nil {
			return err
		}
		if commit.at.After(s.lastCommitAt) {
			s.lastCommitAt = commit.at.Add(time.Second)
		}
		s.commits = append(s.commits, commit)
	}

	return nil
}

func (s *server) processCommit(ctx context.Context, commit *Commit) error {
	log.Printf("processing commit; %s\n", commit.message)
	for filename := range commit.files {
		status, err := s.processFile(ctx, filename, commit.sha)
		if err != nil {
			log.Printf("failed process file; filename=%s; sha=%s; status=%s; err=%+v\n",
				filename, commit.sha, status, err)
		}
		commit.files[filename] = status

		// newContent, err := s.getFile(ctx, filename, commit.sha)
		// if err != nil {
		// 	return err
		// }
		// configName, err := parseConfigName(newContent)
		// if err != nil {
		// 	log.Printf("err=%+v\n", err)
		// }
		// url := getConfigURL(s.ocURL, configName)
		// // We could check old content and only update if it has changed.
		// //oldContent, err := s.GetConfig(url)
		// //if err != nil {
		// //	log.Printf("err=%+v\n", err)
		// //}
		// response, err := s.PatchConfig(url, newContent)
		// if err != nil {
		// 	log.Printf("failed patch config; url=%s; response=%s; err=%+v\n", url, response, err)
		// }
		// commit.files[filename] = "patched"
	}
	return nil
}

func (s *server) processFile(ctx context.Context, filename string, sha string) (string, error) {
	log.Printf("processing file=%s\n", filename)
	newContent, err := s.getFile(ctx, filename, sha)
	if err != nil {
		return "err read file from git", err
	}

	configName, err := parseConfigName(newContent)
	if err != nil {
		return "err parse config name", err
	}
	url := getConfigURL(s.ocURL, configName)

	// We could check old content and only update if it has changed.
	//oldContent, err := s.GetConfig(url)
	//if err != nil {
	//	log.Printf("err=%+v\n", err)
	//}

	response, err := s.PatchConfig(url, newContent)
	if err != nil {
		log.Printf("failed patch config; url=%s; response=%s; err=%+v\n", url, response, err)
		return "err patch oc config", err
	}
	return "patched", nil
}

type ConfigName struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string
	Metadata   struct {
		Name string
	}
}

func parseConfigName(fileContent string) (ConfigName, error) {
	name := ConfigName{}
	err := yaml.Unmarshal([]byte(fileContent), &name)
	return name, err
}

func getConfigURL(host string, configName ConfigName) string {
	url := fmt.Sprintf("%s/apis/%s/%ss/%s", host, configName.APIVersion, configName.Kind, configName.Metadata.Name)
	url = strings.ToLower(url)
	log.Printf("url=%s\n", url)
	return url
}
