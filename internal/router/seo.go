package router

import (
	"fmt"
	"html"
	"strings"
)

type seoPage struct {
	Title       string
	Description string
	Robots      string
	Canonical   string
	Heading     string
	Summary     string
}

var indexableSEOPages = map[string]seoPage{
	"/": {
		Title:       "Survei Online dengan Logika Bercabang · txsurvey",
		Description: "Buat survei online yang terasa seperti percakapan, lengkap dengan logika bercabang, tema hangat, respons anonim, dan analitik.",
		Robots:      "index, follow",
		Canonical:   "",
		Heading:     "Bikin survei yang terasa seperti ngobrol.",
		Summary:     "Satu pertanyaan per layar, alur bercabang yang relevan, dan hasil yang langsung bisa dipakai mengambil keputusan.",
	},
	"/legal": {
		Title:       "Ketentuan Layanan · txsurvey",
		Description: "Ketentuan layanan dan kebijakan privasi txsurvey untuk creator dan responden survei.",
		Robots:      "index, follow",
		Canonical:   "legal",
		Heading:     "Ketentuan layanan dan privasi txsurvey",
		Summary:     "Penjelasan ringkas tentang penggunaan layanan, data creator, jawaban responden, retensi, dan penghapusan.",
	},
	"/contoh-template-survei": {
		Title:       "Template Survei Gratis untuk Tim · txsurvey",
		Description: "Contoh template survei kepuasan karyawan, feedback onboarding, dan kepuasan layanan IT yang siap disesuaikan.",
		Robots:      "index, follow",
		Canonical:   "contoh-template-survei",
		Heading:     "Mulai dari pertanyaan yang sudah teruji.",
		Summary:     "Pilih template survei siap pakai, lihat seluruh pertanyaan, lalu sesuaikan bahasa, tema, dan alurnya.",
	},
	"/template-survei-kepuasan-karyawan": {
		Title:       "Template Survei Kepuasan Karyawan · txsurvey",
		Description: "Ukur kepuasan, loyalitas, dan area perbaikan tim dengan lima pertanyaan singkat.",
		Robots:      "index, follow",
		Canonical:   "template-survei-kepuasan-karyawan",
		Heading:     "Template Survei Kepuasan Karyawan",
		Summary:     "Lima pertanyaan untuk mengukur kepuasan kerja, loyalitas, rekomendasi, dan area perbaikan tim.",
	},
	"/template-feedback-onboarding": {
		Title:       "Template Survei Feedback Onboarding · txsurvey",
		Description: "Pahami pengalaman karyawan baru dan temukan bagian onboarding yang perlu diperbaiki.",
		Robots:      "index, follow",
		Canonical:   "template-feedback-onboarding",
		Heading:     "Template Survei Feedback Onboarding",
		Summary:     "Evaluasi kejelasan peran, kecukupan pelatihan, bantuan mentor, dan pengalaman karyawan baru.",
	},
	"/template-kepuasan-layanan-it": {
		Title:       "Template Survei Kepuasan Layanan IT · txsurvey",
		Description: "Nilai penyelesaian masalah, kecepatan, dan kualitas dukungan tim IT internal.",
		Robots:      "index, follow",
		Canonical:   "template-kepuasan-layanan-it",
		Heading:     "Template Survei Kepuasan Layanan IT",
		Summary:     "Ukur penyelesaian masalah, kecepatan penanganan, kepuasan, dan saran untuk tim IT internal.",
	},
	"/fitur/logika-bercabang": {
		Title:       "Survei dengan logika bercabang · txsurvey",
		Description: "Tampilkan pertanyaan yang relevan berdasarkan jawaban responden dengan alur lompat yang aman dan mudah disusun.",
		Robots:      "index, follow",
		Canonical:   "fitur/logika-bercabang",
		Heading:     "Survei dengan logika bercabang",
		Summary:     "Setiap responden hanya melihat pertanyaan yang relevan berdasarkan jawaban dan jalur surveinya.",
	},
	"/fitur/survei-anonim": {
		Title:       "Survei anonim untuk feedback yang lebih jujur · txsurvey",
		Description: "Kumpulkan respons tanpa meminta identitas responden dan batasi akses hasil hanya kepada pemilik survei.",
		Robots:      "index, follow",
		Canonical:   "fitur/survei-anonim",
		Heading:     "Survei anonim untuk feedback yang lebih jujur",
		Summary:     "Responden tidak perlu login dan hasil survei hanya tersedia di area creator yang terlindungi.",
	},
	"/panduan": {
		Title:       "Panduan Membuat Survei Online · txsurvey",
		Description: "Panduan singkat membuat, membagikan, dan membaca hasil survei online dengan pertanyaan relevan dan privasi yang jelas.",
		Robots:      "index, follow",
		Canonical:   "panduan",
		Heading:     "Buat survei singkat yang menghasilkan jawaban berguna.",
		Summary:     "Panduan menentukan tujuan, menyusun pertanyaan, mengatur alur, membagikan survei, dan membaca hasil.",
	},
}

func seoPageForPath(requestPath string) (seoPage, bool) {
	if page, ok := indexableSEOPages[requestPath]; ok {
		return page, true
	}

	// Application and respondent surfaces are valid SPA routes but deliberately
	// excluded from search. Dynamic survey content is user-generated and can be
	// sensitive, so it never receives a server canonical or share preview.
	//
	// /fitur/ is also a live React route (frontend/src/features/marketing/FeaturePage.tsx,
	// path "/fitur/:featureSlug") with its own content list; any slug not mirrored above
	// as a dedicated indexableSEOPages entry falls back to noindex 200 here rather than a
	// hard 404, so a frontend-only feature page addition degrades gracefully instead of
	// breaking the first server response for crawlers/share bots.
	if requestPath == "/login" || requestPath == "/app" || requestPath == "/templates" ||
		strings.HasPrefix(requestPath, "/forms/") || strings.HasPrefix(requestPath, "/fitur/") ||
		(strings.HasPrefix(requestPath, "/r/") && len(strings.TrimPrefix(requestPath, "/r/")) > 0) {
		return seoPage{
			Title:       "txsurvey",
			Description: "Area aplikasi txsurvey.",
			Robots:      "noindex, nofollow",
		}, true
	}

	return seoPage{}, false
}

func renderSEOHTML(indexHTML []byte, page seoPage, appBaseURL string) []byte {
	base := strings.TrimRight(appBaseURL, "/")
	title := html.EscapeString(page.Title)
	description := html.EscapeString(page.Description)
	robots := html.EscapeString(page.Robots)

	canonicalURL := ""
	twitterCard := "summary"
	if page.Robots == "index, follow" {
		canonicalURL = base + "/" + strings.TrimLeft(page.Canonical, "/")
		twitterCard = "summary_large_image"
	}

	meta := fmt.Sprintf(`<meta name="description" content="%s" />
    <meta name="robots" content="%s" />
    <meta property="og:site_name" content="txsurvey" />
    <meta property="og:type" content="website" />
    <meta property="og:title" content="%s" />
    <meta property="og:description" content="%s" />
    <meta name="twitter:card" content="%s" />
    <meta name="twitter:title" content="%s" />
    <meta name="twitter:description" content="%s" />`, description, robots, title, description, twitterCard, title, description)

	if canonicalURL != "" {
		canonical := html.EscapeString(canonicalURL)
		image := html.EscapeString(base + "/og-image.png")
		meta += fmt.Sprintf(`
    <link rel="canonical" href="%s" />
    <meta property="og:url" content="%s" />
    <meta property="og:image" content="%s" />
    <meta property="og:image:width" content="1200" />
    <meta property="og:image:height" content="630" />
    <meta property="og:image:alt" content="Alur kartu survei bercabang txsurvey" />
    <meta name="twitter:image" content="%s" />`, canonical, canonical, image, image)
	}

	fallback := ""
	if page.Heading != "" {
		fallback = fmt.Sprintf(`<main data-seo-fallback="true"><h1>%s</h1><p>%s</p><nav aria-label="Halaman publik"><a href="%s/">Beranda</a> <a href="%s/contoh-template-survei">Template survei</a> <a href="%s/fitur/logika-bercabang">Logika bercabang</a> <a href="%s/fitur/survei-anonim">Survei anonim</a> <a href="%s/panduan">Panduan</a></nav></main>`,
			html.EscapeString(page.Heading), html.EscapeString(page.Summary),
			html.EscapeString(base), html.EscapeString(base), html.EscapeString(base), html.EscapeString(base), html.EscapeString(base))
	}

	rendered := strings.Replace(string(indexHTML), "<title>txsurvey</title>", "<title>"+title+"</title>", 1)
	rendered = strings.Replace(rendered, "<!-- txsurvey:seo -->", meta, 1)
	rendered = strings.Replace(rendered, "<!-- txsurvey:prerender -->", fallback, 1)
	return []byte(rendered)
}

func sitemapXML(appBaseURL string) []byte {
	base := strings.TrimRight(appBaseURL, "/")
	var urls strings.Builder
	for _, requestPath := range []string{
		"/",
		"/contoh-template-survei",
		"/template-survei-kepuasan-karyawan",
		"/template-feedback-onboarding",
		"/template-kepuasan-layanan-it",
		"/fitur/logika-bercabang",
		"/fitur/survei-anonim",
		"/panduan",
		"/legal",
	} {
		loc := base + requestPath
		urls.WriteString("  <url><loc>")
		urls.WriteString(html.EscapeString(loc))
		urls.WriteString("</loc></url>\n")
	}
	return []byte("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n" +
		"<urlset xmlns=\"http://www.sitemaps.org/schemas/sitemap/0.9\">\n" + urls.String() + "</urlset>\n")
}

func robotsText(appBaseURL string) []byte {
	return []byte("User-agent: *\nAllow: /\n\nSitemap: " + strings.TrimRight(appBaseURL, "/") + "/sitemap.xml\n")
}
