package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	port := flag.Int("port", 8080, "listen port")
	ocToken := flag.String("token", "8F8MDyuc1m9OrUpvYjbPXLFPmYzyrrSy3f1CG-CKNoU", "openshift token")
	ocURL := flag.String("ocurl", "https://192.168.42.93:8443", "openshift server url")
	gitRepo := flag.String("repo", "matobi/kafka-auto-conf", "repo")
	confDir := flag.String("dir", "conf", "dir in repo containing kafka config")
	flag.Parse()

	stopChan := make(chan os.Signal, 3)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

	server, err := NewServer(*ocToken, *ocURL, *gitRepo, *confDir)
	if err != nil {
		log.Printf("error; main; NewServer; err=%+v\n", err)
		return
	}

	log.Printf("Commit history: http://localhost:%d/history\n", *port)

	var wg sync.WaitGroup

	go func() {
		log.Printf("gofunc start\n")
		wg.Add(1)
		defer wg.Done()
		server.waitForPoll()
		log.Printf("gofunc done\n")
	}()

	//func (s *RepositoriesService) ListCommits(ctx context.Context, owner, repo string, opts *CommitsListOptions) ([]*RepositoryCommit, *Response, error)

	// filenames, err := server.getDir(context.Background(), server.confDir)
	// if err != nil {
	// 	log.Printf("err=%+v\n", err)
	// }
	// for _, filename := range filenames {
	// 	log.Printf("FILE: %s\n", filename)
	// }
	// ##########################################3
	// server.commits, err = server.listCommits(context.Background())
	// if err != nil {
	// 	log.Printf("err=%+v\n", err)
	// }
	// for _, commit := range server.commits {
	// 	log.Printf("##### commit=%s\n", commit.message)
	// 	for filename := range commit.files {
	// 		_, err := server.getFile(context.Background(), filename, commit.sha)
	// 		if err != nil {
	// 			log.Printf("err=%s\n", err)
	// 		}
	// 	}
	// }

	//repo, _, err := server.gitClient.Repositories.GetBranch(context.Background(), "matobi", "kafka-auto-conf", "master")
	// commits, _, err := server.gitClient.Repositories.ListCommits(context.Background(), "matobi", "kafka-auto-conf",
	// 	&github.CommitsListOptions{SHA: "master"}) // 02ff703310a2d93ec4b3540e8631a35d84095209
	// if err != nil {
	// 	fmt.Printf("err=%+v\n", err)
	// 	return
	// }
	// log.Printf("commitsCount: %d\n", len(commits))
	// for _, commit := range commits {
	// 	//log.Printf("[%d] commit: sha=%s; comitter=%s; comment=%s; commit=%+v\n",
	// 	//	id, commit.GetSHA(), commit.GetCommit().GetCommitter().GetName(), commit.GetCommit().GetMessage(), commit)
	// 	cmtA := commit.GetCommit()
	// 	log.Printf("COMMIT###A=%+v\n", cmtA)

	// 	cmt, _, err := server.gitClient.Repositories.GetCommit(context.Background(), "matobi", "kafka-auto-conf", commit.GetSHA())
	// 	if err != nil {
	// 		log.Printf("err=%+v\n", err)
	// 	}
	// 	log.Printf("COMMIT=%+v\n", cmt.GetCommit().GetCommitter().GetDate())
	// }

	//fmt.Printf("repo=%+v\n", repo)

	srv := &http.Server{Addr: fmt.Sprintf(":%d", *port), Handler: server.router}
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Printf("error; failed to serve http; err=%+v\n", err)
		}
	}()

	<-stopChan // wait for SIGINT
	log.Printf("got kill signal")
	server.isQuit = true
	ctx, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelFunc()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("error; failed shutdown http server; err=%+v\n", err)
	}
	wg.Wait() // todo: timeout
	log.Printf("bye\n")
}
