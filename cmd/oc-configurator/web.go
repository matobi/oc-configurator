package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/oauth2"

	"github.com/google/go-github/v29/github"
)

type server struct {
	ocToken      string
	ocURL        string
	gitRepo      string
	confDir      string
	lastCommitAt time.Time

	commits []Commit
	isQuit  bool

	gitClient *github.Client
	router    *http.ServeMux
	ocClient  *http.Client
}

// Create a Server object with all config.
func NewServer(ocToken, ocURL, gitRepo, confDir string) (*server, error) {

	// minishift certificate is self-signed. Ignore certificate check.
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	ocClient := &http.Client{
		Timeout:   time.Second * 10,
		Transport: tr,
	}

	ctx := context.Background()
	tsGithub := oauth2.StaticTokenSource(
		// githup token
		&oauth2.Token{AccessToken: "65ca65d560599012441ab5b4a0d95a0aeefafbc8"},
	)
	tcGithub := oauth2.NewClient(ctx, tsGithub)

	s := &server{
		ocToken:      ocToken,
		ocURL:        ocURL,
		gitRepo:      gitRepo,
		confDir:      confDir,
		lastCommitAt: time.Time{},
		commits:      []Commit{},
		router:       http.NewServeMux(),
		ocClient:     ocClient,
		gitClient:    github.NewClient(tcGithub),
	}
	s.routes()
	return s, nil
}

func (s *server) routes() {
	s.router.HandleFunc("/history", s.webHistory())
}

// Shows simple web page with commit history.
func (s *server) webHistory() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`
		<html><head>
		<style>
		* {
			font-size: 100%;
			font-family: Roboto, Tahoma, Geneva, sans-serif;
		}
		table {
			border-collapse: collapse;
			width: 100%;
		}
		th, td {
			text-align: left;
			padding: 8px;
		}
		tr:nth-child(even){background-color: #f2f2f2}
		th {
			background-color: #808080;
			color: white;
		}
		</style>
		</head><body><h1>commits</h1><table>`))

		w.Write([]byte(fmt.Sprintf(`<tr><th>time</th><th>message</th><th>committer</th><th>file</th><th>status</th></tr>`)))

		// Show newset commit first.
		for i := len(s.commits) - 1; i >= 0; i-- {
			commit := s.commits[i]
			for file, result := range commit.files {
				w.Write([]byte(fmt.Sprintf(`<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>`,
					commit.at.Format("06-01-02 15:04:05"), commit.message, commit.committer, file, result)))
			}
		}
		w.Write([]byte(`</table></body></html>`))
	}
}
