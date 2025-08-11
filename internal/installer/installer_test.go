package installer

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/google/go-github/v74/github"
)

// TODO: kill this
func TestReleaseSourceDownload(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/owner/repo/releases/tags/v1.0.0":
			_ = json.NewEncoder(w).Encode(&github.RepositoryRelease{
				TagName: github.Ptr("v1.0.0"),
			})
		case "/repos/owner/repo/tarballs/tags/v1.0.0":
			w.Write([]byte("TAR-GZ RESPONSE IN BYTES"))
		default:
			http.NotFound(w, r)
		}
	})
	srv := httptest.NewServer(handler)

	t.Cleanup(srv.Close)
	cli := github.NewClient(srv.Client())

	cli.BaseURL, _ = url.Parse(srv.URL + "/")

	inst := New(cli.Repositories)

	_ = inst // ???
}
