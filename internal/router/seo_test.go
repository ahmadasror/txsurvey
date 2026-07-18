package router

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/gin-gonic/gin"
)

func seoTestRouter(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	dist := fstest.MapFS{
		"index.html":         &fstest.MapFile{Data: []byte(`<!doctype html><html lang="id"><head><title>txsurvey</title><!-- txsurvey:seo --></head><body><div id="root"><!-- txsurvey:prerender --></div></body></html>`)},
		"assets/app-hash.js": &fstest.MapFile{Data: []byte("export {}")},
	}
	var filesystem fs.FS = dist
	r := gin.New()
	serveSPA(r, filesystem, "https://brainzap.net/txsurvey")
	return r
}

func TestSEOIndexablePageFirstResponse(t *testing.T) {
	r := seoTestRouter(t)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/fitur/logika-bercabang", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	body := w.Body.String()
	for _, want := range []string{
		`<meta name="robots" content="index, follow"`,
		`<link rel="canonical" href="https://brainzap.net/txsurvey/fitur/logika-bercabang"`,
		`<meta property="og:image" content="https://brainzap.net/txsurvey/og-image.png"`,
		`<h1>Survei dengan logika bercabang</h1>`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("response missing %q", want)
		}
	}
}

func TestSEOIndexHTMLMatchesRootResponse(t *testing.T) {
	r := seoTestRouter(t)

	root := httptest.NewRecorder()
	r.ServeHTTP(root, httptest.NewRequest(http.MethodGet, "/", nil))
	if root.Code != http.StatusOK || root.Header().Get("Cache-Control") != "no-cache" {
		t.Fatalf("/ response = %d %q", root.Code, root.Header().Get("Cache-Control"))
	}

	indexHTMLResp := httptest.NewRecorder()
	r.ServeHTTP(indexHTMLResp, httptest.NewRequest(http.MethodGet, "/index.html", nil))
	if indexHTMLResp.Code != http.StatusOK || indexHTMLResp.Header().Get("Cache-Control") != "no-cache" {
		t.Fatalf("/index.html response = %d %q, want the same SEO-enriched no-cache response as /", indexHTMLResp.Code, indexHTMLResp.Header().Get("Cache-Control"))
	}
	if indexHTMLResp.Body.String() != root.Body.String() {
		t.Fatal("/index.html must render the same SEO-enriched body as /, not the raw static file")
	}
}

func TestSEOUnknownFeatureSlugIsNoIndexNotFound(t *testing.T) {
	r := seoTestRouter(t)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/fitur/does-not-exist", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (soft noindex, not a hard 404)", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, `content="noindex, nofollow"`) {
		t.Fatal("unmirrored /fitur/ slug must be noindex")
	}
	if strings.Contains(body, `rel="canonical"`) {
		t.Fatal("unmirrored /fitur/ slug must not publish a canonical")
	}
}

func TestSEOPrivateRouteIsNoIndex(t *testing.T) {
	r := seoTestRouter(t)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/forms/abc/results", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, `content="noindex, nofollow"`) {
		t.Fatal("private route must be noindex")
	}
	if strings.Contains(body, `rel="canonical"`) {
		t.Fatal("private route must not publish a canonical")
	}
}

func TestSEOUnknownRouteReturns404(t *testing.T) {
	r := seoTestRouter(t)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/definitely-not-real", nil))

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", w.Code)
	}
	if !strings.Contains(w.Body.String(), `content="noindex, nofollow"`) {
		t.Fatal("404 response must be noindex")
	}
}

func TestSEOSitemapAndRobots(t *testing.T) {
	r := seoTestRouter(t)

	sitemap := httptest.NewRecorder()
	r.ServeHTTP(sitemap, httptest.NewRequest(http.MethodGet, "/sitemap.xml", nil))
	if sitemap.Code != http.StatusOK || !strings.Contains(sitemap.Header().Get("Content-Type"), "application/xml") {
		t.Fatalf("sitemap response = %d %q", sitemap.Code, sitemap.Header().Get("Content-Type"))
	}
	if got := strings.Count(sitemap.Body.String(), "<url>"); got != 9 {
		t.Fatalf("sitemap URL count = %d, want 9", got)
	}
	if strings.Contains(sitemap.Body.String(), "/login") || strings.Contains(sitemap.Body.String(), "/forms/") {
		t.Fatal("sitemap must exclude noindex application routes")
	}

	robots := httptest.NewRecorder()
	r.ServeHTTP(robots, httptest.NewRequest(http.MethodGet, "/robots.txt", nil))
	if robots.Code != http.StatusOK || !strings.Contains(robots.Body.String(), "Sitemap: https://brainzap.net/txsurvey/sitemap.xml") {
		t.Fatalf("unexpected robots response: %d %q", robots.Code, robots.Body.String())
	}
}

func TestSEOAssetCachingAndTrailingSlash(t *testing.T) {
	r := seoTestRouter(t)

	asset := httptest.NewRecorder()
	r.ServeHTTP(asset, httptest.NewRequest(http.MethodGet, "/assets/app-hash.js", nil))
	if asset.Code != http.StatusOK || asset.Header().Get("Cache-Control") != "public, max-age=31536000, immutable" {
		t.Fatalf("asset cache response = %d %q", asset.Code, asset.Header().Get("Cache-Control"))
	}

	redirect := httptest.NewRecorder()
	r.ServeHTTP(redirect, httptest.NewRequest(http.MethodGet, "/panduan/?from=test", nil))
	if redirect.Code != http.StatusMovedPermanently || redirect.Header().Get("Location") != "/panduan?from=test" {
		t.Fatalf("redirect = %d %q", redirect.Code, redirect.Header().Get("Location"))
	}
}
