package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

// --------------------------------------------------------------------------------

type WatchItem struct {
	Repo   string `json:"repo"`
	Branch string `json:"branch"`
	Script string `json:"script"`
}

type Config struct {
	BindHost string      `json:"bind"`
	Items    []WatchItem `json:"items"`
}

// --------------------------------------------------------------------------------

type Repository struct {
	Url         string `json:"url"` // "https://github.com/qiniu/api"
	AbsoluteUrl string `json:"absolute_url"`
}

type Commit struct {
	Branch string `json:"branch"`
}

type Payload struct {
	Ref      string     `json:"ref"` // "refs/heads/develop"
	Repo     Repository `json:"repository"`
	CanonUrl string     `json:"canon_url"`
	Commits  []Commit   `json:"commits"`
}

// --------------------------------------------------------------------------------

var cfg Config

// --------------------------------------------------------------------------------

func runScript(item *WatchItem) (outStr string, err error) {
	script := "./" + item.Script
	out, err := exec.Command("bash", "-c", script).Output()
	if err != nil {
		log.Printf("Exec command failed: %s\n", err)
		return err.Error(), err
	}

	log.Printf("Run %s output: %s\n", script, string(out))
	return string(out), nil
}

func handleGithub(event Payload, cfg *Config) (result string, err error) {
	result = "miss"
	for _, item := range cfg.Items {
		if event.Repo.Url == item.Repo && strings.Contains(event.Ref, item.Branch) {
			result, err = runScript(&item)
			if err != nil {
				log.Printf("run script error: %s\n", err)
			}
			break
		}
	}

	return
}

// func handleBitbucket(event Payload, cfg *Config) {
// 	changingBranches := make(map[string]bool)
//
// 	for _, commit := range event.Commits {
// 		changingBranches[commit.Branch] = true
// 	}
//
// 	repo := strings.TrimRight(event.CanonUrl+event.Repo.AbsoluteUrl, "/")
//
// 	for _, item := range cfg.Items {
// 		if strings.TrimRight(item.Repo, "/") == repo && changingBranches[item.Branch] {
// 			runScript(&item)
// 		}
// 	}
// 	return
// }

func handle(w http.ResponseWriter, req *http.Request) {
	defer func() {
		_ = req.Body.Close()
	}()
	decoder := json.NewDecoder(req.Body)
	var event Payload
	err := decoder.Decode(&event)
	if err != nil {
		log.Printf("payload json decode failed: %s\n", err)
		return
	}
	log.Println("payload json decode success: ", event)

	var out string
	// if event.CanonUrl == "https://bitbucket.org" {
	// 	handleBitbucket(event, &cfg)
	// 	return
	// }

	out, _ = handleGithub(event, &cfg)
	_, _ = fmt.Fprintf(w, out)
}

// --------------------------------------------------------------------------------

func main() {

	if len(os.Args) < 2 {
		println("Usage: webhook <ConfigFile>\n")
		return
	}

	cfgBuf, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		log.Println("Read config file failed:", err)
		return
	}

	err = json.Unmarshal(cfgBuf, &cfg)
	if err != nil {
		log.Println("Unmarshal config failed:", err)
		return
	}

	http.HandleFunc("/", handle)
	log.Fatal(http.ListenAndServe(cfg.BindHost, nil))
}

// --------------------------------------------------------------------------------
