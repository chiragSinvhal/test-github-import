package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github/stub"

	"github.com/google/go-github/v40/github"
	"github.com/gorilla/mux"
	"golang.org/x/oauth2"
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/",fetchData).Methods("GET")
	log.Fatal(http.ListenAndServe("localhost:10000",r))
}

type PrimaryData struct {
	LhInstanceType string      `json:"lhInstanceType"`
	SolutionCid    interface{} `json:"solutionCID"`
}

func fetchData(rw http.ResponseWriter, r *http.Request){
	jsonURL := "https://github.hpe.com/hpe/lhat-scid-encryption-integration-test/tenants/"
	gitPersonalAccessToken := FromEnv("LIGHTHOUSE_OPS_CONSOLE_GIT_TOKEN","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")
	content, err := DefaultFactory.New(gitPersonalAccessToken).GetFile(jsonURL + "test" + "/sites/" + "test" + "/infra.json", "")
	if err != nil {

		stub.JSON(rw, 500, err)
		return
	}

	var payload PrimaryData
	data := payload.SolutionCid
	json.Unmarshal([]byte(content), &data)

	payload.LhInstanceType = "GreenLake_Lighthouse_Integrated_Systems"
	payload.SolutionCid = data

	stub.JSON(rw, 200, payload)
	return
}

func FromEnv(key, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = fallback
	}

	return value
}

// New returns a new Repo for GitOPs.
func (defaultFactory) New(personalAccessToken string) OPs {
	return &Repo{PersonalAccessToken: personalAccessToken}
}

type defaultFactory struct{}

var DefaultFactory Factory = defaultFactory{}


type Factory interface {
	New(personalAccessToken string) OPs
}

type OPs interface {
	// GetFile gets a file at the corresponding ref.
	GetFile(urlPath, ref string) (string, error)
}

type Repo struct {
	PersonalAccessToken string
}

func (r *Repo) GetFile(urlPath, ref string) (contents string, err error) {
	organization, repositoryName, subDirectory, u, err := parseRepository(urlPath)
	if err != nil {
		return
	}

	ctx, done := context.WithTimeout(context.Background(), 120*time.Second)
	defer done()

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{
			AccessToken: r.PersonalAccessToken,
		},
	)

	client, err := getClient(u, oauth2.NewClient(ctx, ts))
	if err != nil {
		return
	}

	opts := &github.RepositoryContentGetOptions{
		Ref: ref,
	}
	fileContent, _, _, err := client.Repositories.GetContents(ctx, organization, repositoryName, subDirectory, opts)
	if err != nil {
		return
	}

	return fileContent.GetContent()
}


func parseRepository(urlPath string) (organization, repositoryName, subPath string, u *url.URL, err error) {
	// Ensure we have a sane URL
	u, err = url.Parse(urlPath)
	if err != nil {
		return
	}

	// Verify we have a https://<hostname>/<repo> as a minimum
	components := strings.Split(u.Path, "/")
	if len(components) < 3 {
		err = fmt.Errorf("repository must be of the form https://github.com/<org>/<repo> as the minimum")

		return
	}
	// leading /
	if components[0] == "" {
		components = components[1:]
	}

	// extract parts of the URL now that we know the form for sure.
	organization = components[0]
	repositoryName = components[1]
	subPath = strings.Join(components[2:], "/")

	return
}

func getClient(u *url.URL, httpClient *http.Client) (client *github.Client, err error) {
	if u.Scheme != "https" {
		return nil, fmt.Errorf("only https locations are supported")
	}

	if strings.HasPrefix(u.Host, "github.com") {
		client = github.NewClient(httpClient)
	} else {
		base := u.Scheme + "://" + u.Hostname()
		client, err = github.NewEnterpriseClient(base, base, httpClient)
	}

	return
}


