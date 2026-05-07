package extractor

import (
	"testing"
)

func TestNewXCards(t *testing.T) {
	xc := NewXCards()
	if xc == nil {
		t.Fatal("NewXCards returned nil")
	}
}

func TestParseXCards_TwitterOnly(t *testing.T) {
	html := `<meta name="twitter:card" content="summary"><meta name="twitter:title" content="Test Title">`
	xc, errs := ParseXCards("", html)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if xc == nil {
		t.Fatal("expected non-nil XCards")
	}
	if xc.Card != "summary" {
		t.Errorf("Card: %q", xc.Card)
	}
	if xc.Title != "Test Title" {
		t.Errorf("Title: %q", xc.Title)
	}
}

func TestParseXCards_WithOGFallback(t *testing.T) {
	// No twitter tags, but OG tags present → XCards filled from OG
	html := `<meta property="og:title" content="OG Title"><meta property="og:type" content="website">`
	xc, errs := ParseXCards("", html)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if xc == nil {
		t.Fatal("expected non-nil XCards from OG fallback")
	}
	if xc.Title != "OG Title" {
		t.Errorf("Title: %q", xc.Title)
	}
}

func TestParseXCards_WithURLResolution(t *testing.T) {
	html := `<meta name="twitter:card" content="summary"><meta name="twitter:url" content="/page">`
	xc, errs := ParseXCards("http://example.com", html)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if xc == nil {
		t.Fatal("expected non-nil XCards")
	}
	if xc.URL != "http://example.com/page" {
		t.Errorf("URL: %q", xc.URL)
	}
}

func TestParseXCards_NeitherTwitterNorOG(t *testing.T) {
	// No twitter or OG tags → nil
	xc, errs := ParseXCards("", `<html><head></head></html>`)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if xc != nil {
		t.Errorf("expected nil, got %+v", xc)
	}
}

func TestExtractXCards_NoProperties(t *testing.T) {
	// Using HTML with text nodes and comments to trigger the default:continue branch
	xc, errs := extractXCards(`<html><head><!-- comment --></head><body>text node</body></html>`)
	if len(errs) != 0 {
		t.Errorf("unexpected errors: %v", errs)
	}
	if xc != nil {
		t.Errorf("expected nil, got %+v", xc)
	}
}

func TestExtractXCards_WithProperties(t *testing.T) {
	html := `<meta name="twitter:card" content="summary_large_image"><meta name="twitter:site" content="@site">`
	xc, errs := extractXCards(html)
	if len(errs) != 0 {
		t.Errorf("unexpected errors: %v", errs)
	}
	if xc == nil {
		t.Fatal("expected non-nil XCards")
	}
	if xc.Card != "summary_large_image" {
		t.Errorf("Card: %q", xc.Card)
	}
	if xc.Site != "@site" {
		t.Errorf("Site: %q", xc.Site)
	}
}

func TestIsXCardsProperty(t *testing.T) {
	tests := []struct {
		prop string
		want bool
	}{
		{"twitter:card", true},
		{"og:title", true},
		{"music:duration", true},
		{"video:actor", true},
		{"article:published_time", true},
		{"book:isbn", true},
		{"profile:first_name", true},
		{"unknown:prop", false},
		{"noprefix", false},
	}
	for _, tt := range tests {
		t.Run(tt.prop, func(t *testing.T) {
			got := isXCardsProperty(tt.prop)
			if got != tt.want {
				t.Errorf("isXCardsProperty(%q) = %v; want %v", tt.prop, got, tt.want)
			}
		})
	}
}

func TestParseXCardsMetaTag_AllBranches(t *testing.T) {
	xc := NewXCards()

	// X-specific
	parseXCardsMetaTag(xc, "twitter:card", "summary")
	parseXCardsMetaTag(xc, "twitter:site", "@mysite")
	parseXCardsMetaTag(xc, "twitter:creator", "@creator")

	// Basic
	parseXCardsMetaTag(xc, "twitter:type", "article")
	parseXCardsMetaTag(xc, "twitter:title", "Card Title")
	parseXCardsMetaTag(xc, "twitter:url", "/card-url")

	// Optional
	parseXCardsMetaTag(xc, "twitter:description", "Card desc")
	parseXCardsMetaTag(xc, "twitter:determiner", "a")
	parseXCardsMetaTag(xc, "twitter:locale", "en_US")
	parseXCardsMetaTag(xc, "twitter:locale:alternate", "fr_FR")
	parseXCardsMetaTag(xc, "twitter:site_name", "My Site")

	if xc.Card != "summary" {
		t.Errorf("Card: %q", xc.Card)
	}
	if xc.Site != "@mysite" {
		t.Errorf("Site: %q", xc.Site)
	}
	if xc.Creator != "@creator" {
		t.Errorf("Creator: %q", xc.Creator)
	}
	if xc.Type != "article" {
		t.Errorf("Type: %q", xc.Type)
	}
	if xc.Title != "Card Title" {
		t.Errorf("Title: %q", xc.Title)
	}
	if xc.URL != "/card-url" {
		t.Errorf("URL: %q", xc.URL)
	}
	if xc.Description != "Card desc" {
		t.Errorf("Description: %q", xc.Description)
	}
	if xc.Determiner != "a" {
		t.Errorf("Determiner: %q", xc.Determiner)
	}
	if xc.Locale != "en_US" {
		t.Errorf("Locale: %q", xc.Locale)
	}
	if len(xc.LocaleAlternate) != 1 {
		t.Errorf("LocaleAlternate: %v", xc.LocaleAlternate)
	}
	if xc.SiteName != "My Site" {
		t.Errorf("SiteName: %q", xc.SiteName)
	}
}

func TestParseXCardsMetaTag_Image(t *testing.T) {
	xc := NewXCards()
	parseXCardsMetaTag(xc, "twitter:image", "/img.jpg")
	parseXCardsMetaTag(xc, "twitter:image:secure_url", "/secure.jpg")
	parseXCardsMetaTag(xc, "twitter:image:type", "image/jpeg")
	parseXCardsMetaTag(xc, "twitter:image:width", "800")
	parseXCardsMetaTag(xc, "twitter:image:height", "600")
	parseXCardsMetaTag(xc, "twitter:image:alt", "Alt text")

	if len(xc.XCardsImage) != 1 {
		t.Fatalf("expected 1 image, got %d", len(xc.XCardsImage))
	}
	img := xc.XCardsImage[0]
	if img.URL != "/img.jpg" {
		t.Errorf("URL: %q", img.URL)
	}
	if img.SecureURL != "/secure.jpg" {
		t.Errorf("SecureURL: %q", img.SecureURL)
	}
	if img.Type != "image/jpeg" {
		t.Errorf("Type: %q", img.Type)
	}
	if img.Width != 800 {
		t.Errorf("Width: %d", img.Width)
	}
	if img.Height != 600 {
		t.Errorf("Height: %d", img.Height)
	}
	if img.Alt != "Alt text" {
		t.Errorf("Alt: %q", img.Alt)
	}
}

func TestParseXCardsMetaTag_Video(t *testing.T) {
	xc := NewXCards()
	parseXCardsMetaTag(xc, "twitter:video", "/vid.mp4")
	parseXCardsMetaTag(xc, "twitter:video:secure_url", "/secure.mp4")
	parseXCardsMetaTag(xc, "twitter:video:type", "video/mp4")
	parseXCardsMetaTag(xc, "twitter:video:width", "1920")
	parseXCardsMetaTag(xc, "twitter:video:height", "1080")

	if len(xc.XCardsVideo) != 1 {
		t.Fatalf("expected 1 video")
	}
	v := xc.XCardsVideo[0]
	if v.URL != "/vid.mp4" {
		t.Errorf("URL: %q", v.URL)
	}
	if v.SecureURL != "/secure.mp4" {
		t.Errorf("SecureURL: %q", v.SecureURL)
	}
	if v.Type != "video/mp4" {
		t.Errorf("Type: %q", v.Type)
	}
	if v.Width != 1920 {
		t.Errorf("Width: %d", v.Width)
	}
	if v.Height != 1080 {
		t.Errorf("Height: %d", v.Height)
	}
}

func TestParseXCardsMetaTag_Audio(t *testing.T) {
	xc := NewXCards()
	parseXCardsMetaTag(xc, "twitter:audio", "/audio.mp3")
	parseXCardsMetaTag(xc, "twitter:audio:secure_url", "/secure.mp3")
	parseXCardsMetaTag(xc, "twitter:audio:type", "audio/mpeg")

	if len(xc.XCardsAudio) != 1 {
		t.Fatalf("expected 1 audio")
	}
	a := xc.XCardsAudio[0]
	if a.URL != "/audio.mp3" {
		t.Errorf("URL: %q", a.URL)
	}
	if a.SecureURL != "/secure.mp3" {
		t.Errorf("SecureURL: %q", a.SecureURL)
	}
	if a.Type != "audio/mpeg" {
		t.Errorf("Type: %q", a.Type)
	}
}

func TestParseXCardsMetaTag_Music(t *testing.T) {
	xc := NewXCards()
	parseXCardsMetaTag(xc, "music:duration", "300")
	parseXCardsMetaTag(xc, "music:album", "Album")
	parseXCardsMetaTag(xc, "music:album:disc", "2")
	parseXCardsMetaTag(xc, "music:album:track", "7")
	parseXCardsMetaTag(xc, "music:musician", "http://example.com/artist")
	parseXCardsMetaTag(xc, "music:song", "http://example.com/song")
	parseXCardsMetaTag(xc, "music:creator", "http://example.com/creator")
	parseXCardsMetaTag(xc, "music:release_date", "2023-05-01")

	if xc.Music == nil {
		t.Fatal("Music is nil")
	}
	if xc.Music.Duration != 300 {
		t.Errorf("Duration: %d", xc.Music.Duration)
	}
	if xc.Music.Album != "Album" {
		t.Errorf("Album: %q", xc.Music.Album)
	}
	if xc.Music.AlbumDisc != 2 {
		t.Errorf("AlbumDisc: %d", xc.Music.AlbumDisc)
	}
	if xc.Music.AlbumTrack != 7 {
		t.Errorf("AlbumTrack: %d", xc.Music.AlbumTrack)
	}
	if len(xc.Music.Musician) != 1 {
		t.Errorf("Musician: %v", xc.Music.Musician)
	}
	if len(xc.Music.Song) != 1 {
		t.Errorf("Song: %v", xc.Music.Song)
	}
	if len(xc.Music.Creator) != 1 {
		t.Errorf("Creator: %v", xc.Music.Creator)
	}
	if xc.Music.ReleaseDate != "2023-05-01" {
		t.Errorf("ReleaseDate: %q", xc.Music.ReleaseDate)
	}
}

func TestParseXCardsMetaTag_VideoSpecific(t *testing.T) {
	xc := NewXCards()
	parseXCardsMetaTag(xc, "video:actor", "http://example.com/actor")
	parseXCardsMetaTag(xc, "video:actor:role", "Hero")
	parseXCardsMetaTag(xc, "video:director", "http://example.com/director")
	parseXCardsMetaTag(xc, "video:writer", "http://example.com/writer")
	parseXCardsMetaTag(xc, "video:duration", "90")
	parseXCardsMetaTag(xc, "video:release_date", "2023-07-01")
	parseXCardsMetaTag(xc, "video:tag", "drama")
	parseXCardsMetaTag(xc, "video:series", "http://example.com/series")

	if xc.Video == nil {
		t.Fatal("Video is nil")
	}
	if len(xc.Video.Actor) != 1 {
		t.Errorf("Actor: %v", xc.Video.Actor)
	}
	if xc.Video.Actor[0].Role != "Hero" {
		t.Errorf("Actor Role: %q", xc.Video.Actor[0].Role)
	}
	if len(xc.Video.Director) != 1 {
		t.Errorf("Director: %v", xc.Video.Director)
	}
	if len(xc.Video.Writer) != 1 {
		t.Errorf("Writer: %v", xc.Video.Writer)
	}
	if xc.Video.Duration != 90 {
		t.Errorf("Duration: %d", xc.Video.Duration)
	}
	if xc.Video.ReleaseDate.IsZero() {
		t.Error("ReleaseDate is zero")
	}
	if len(xc.Video.Tag) != 1 {
		t.Errorf("Tag: %v", xc.Video.Tag)
	}
	if xc.Video.Series != "http://example.com/series" {
		t.Errorf("Series: %q", xc.Video.Series)
	}
}

func TestParseXCardsMetaTag_Article(t *testing.T) {
	xc := NewXCards()
	parseXCardsMetaTag(xc, "article:published_time", "2023-01-01T00:00:00Z")
	parseXCardsMetaTag(xc, "article:modified_time", "2023-02-01T00:00:00Z")
	parseXCardsMetaTag(xc, "article:expiration_time", "2024-01-01T00:00:00Z")
	parseXCardsMetaTag(xc, "article:author", "http://example.com/author")
	parseXCardsMetaTag(xc, "article:section", "Tech")
	parseXCardsMetaTag(xc, "article:tag", "go")

	if xc.Article == nil {
		t.Fatal("Article is nil")
	}
	if xc.Article.PublishedTime.IsZero() {
		t.Error("PublishedTime is zero")
	}
	if xc.Article.ModifiedTime.IsZero() {
		t.Error("ModifiedTime is zero")
	}
	if xc.Article.ExpirationTime.IsZero() {
		t.Error("ExpirationTime is zero")
	}
	if len(xc.Article.Author) != 1 {
		t.Errorf("Author: %v", xc.Article.Author)
	}
	if xc.Article.Section != "Tech" {
		t.Errorf("Section: %q", xc.Article.Section)
	}
	if len(xc.Article.Tag) != 1 {
		t.Errorf("Tag: %v", xc.Article.Tag)
	}
}

func TestParseXCardsMetaTag_Book(t *testing.T) {
	xc := NewXCards()
	parseXCardsMetaTag(xc, "book:author", "http://example.com/author")
	parseXCardsMetaTag(xc, "book:isbn", "978-0-06-112008-4")
	parseXCardsMetaTag(xc, "book:release_date", "2023-09-01")
	parseXCardsMetaTag(xc, "book:tag", "sci-fi")

	if xc.Book == nil {
		t.Fatal("Book is nil")
	}
	if len(xc.Book.Author) != 1 {
		t.Errorf("Author: %v", xc.Book.Author)
	}
	if xc.Book.ISBN != "978-0-06-112008-4" {
		t.Errorf("ISBN: %q", xc.Book.ISBN)
	}
	if xc.Book.ReleaseDate.IsZero() {
		t.Error("ReleaseDate is zero")
	}
	if len(xc.Book.Tag) != 1 {
		t.Errorf("Tag: %v", xc.Book.Tag)
	}
}

func TestParseXCardsMetaTag_Profile(t *testing.T) {
	xc := NewXCards()
	parseXCardsMetaTag(xc, "profile:first_name", "Jane")
	parseXCardsMetaTag(xc, "profile:last_name", "Smith")
	parseXCardsMetaTag(xc, "profile:username", "janesmith")
	parseXCardsMetaTag(xc, "profile:gender", "female")

	if xc.Profile == nil {
		t.Fatal("Profile is nil")
	}
	if xc.Profile.FirstName != "Jane" {
		t.Errorf("FirstName: %q", xc.Profile.FirstName)
	}
	if xc.Profile.LastName != "Smith" {
		t.Errorf("LastName: %q", xc.Profile.LastName)
	}
	if xc.Profile.Username != "janesmith" {
		t.Errorf("Username: %q", xc.Profile.Username)
	}
	if xc.Profile.Gender != "female" {
		t.Errorf("Gender: %q", xc.Profile.Gender)
	}
}

func TestHandleXCardsImageProperty_LastIdxNegativeFallback(t *testing.T) {
	// Trigger the lastIdx < 0 fallback: send twitter:image:alt before any twitter:image
	xc := NewXCards()
	// parts[1] is "image" but len(parts)==3 → no new entry created in first check.
	// len(xc.XCardsImage)==0 so lastIdx = -1 → fallback creates entry.
	handleXCardsImageProperty(xc, []string{"twitter", "image", "alt"}, "Alternative text")
	if len(xc.XCardsImage) != 1 {
		t.Fatalf("expected 1 image after lastIdx<0 fallback, got %d", len(xc.XCardsImage))
	}
	if xc.XCardsImage[0].Alt != "Alternative text" {
		t.Errorf("Alt: %q", xc.XCardsImage[0].Alt)
	}
}

func TestHandleXCardsVideoProperty(t *testing.T) {
	xc := NewXCards()
	handleXCardsVideoProperty(xc, []string{"twitter", "video"}, "/vid.mp4")
	if len(xc.XCardsVideo) != 1 {
		t.Fatalf("expected 1 video")
	}
	handleXCardsVideoProperty(xc, []string{"twitter", "video", "secure_url"}, "/s.mp4")
	handleXCardsVideoProperty(xc, []string{"twitter", "video", "type"}, "video/mp4")
	handleXCardsVideoProperty(xc, []string{"twitter", "video", "width"}, "640")
	handleXCardsVideoProperty(xc, []string{"twitter", "video", "height"}, "480")
	v := xc.XCardsVideo[0]
	if v.SecureURL != "/s.mp4" {
		t.Errorf("SecureURL: %q", v.SecureURL)
	}
	if v.Type != "video/mp4" {
		t.Errorf("Type: %q", v.Type)
	}
	if v.Width != 640 {
		t.Errorf("Width: %d", v.Width)
	}
	if v.Height != 480 {
		t.Errorf("Height: %d", v.Height)
	}
}

func TestHandleXCardsAudioProperty(t *testing.T) {
	xc := NewXCards()
	handleXCardsAudioProperty(xc, []string{"twitter", "audio"}, "/audio.mp3")
	if len(xc.XCardsAudio) != 1 {
		t.Fatalf("expected 1 audio")
	}
	handleXCardsAudioProperty(xc, []string{"twitter", "audio", "secure_url"}, "/s.mp3")
	handleXCardsAudioProperty(xc, []string{"twitter", "audio", "type"}, "audio/mpeg")
	a := xc.XCardsAudio[0]
	if a.SecureURL != "/s.mp3" {
		t.Errorf("SecureURL: %q", a.SecureURL)
	}
	if a.Type != "audio/mpeg" {
		t.Errorf("Type: %q", a.Type)
	}
}

func TestResolveXCardsURLs(t *testing.T) {
	xc := &XCards{
		URL: "/page",
		OpenGraphImage: []OpenGraphImage{
			{URL: "/img.jpg", SecureURL: "/simg.jpg"},
		},
		OpenGraphVideo: []OpenGraphVideo{
			{URL: "/vid.mp4", SecureURL: "/svid.mp4"},
		},
		OpenGraphAudio: []OpenGraphAudio{
			{URL: "/audio.mp3", SecureURL: "/saudio.mp3"},
		},
		XCardsImage: []XCardsImage{
			{URL: "/ximg.jpg", SecureURL: "/sximg.jpg"},
		},
		XCardsVideo: []XCardsVideo{
			{URL: "/xvid.mp4", SecureURL: "/sxvid.mp4"},
		},
		XCardsAudio: []XCardsAudio{
			{URL: "/xaudio.mp3"},
		},
	}
	resolveXCardsURLs(xc, "http://example.com")

	if xc.URL != "http://example.com/page" {
		t.Errorf("URL: %q", xc.URL)
	}
	if xc.OpenGraphImage[0].URL != "http://example.com/img.jpg" {
		t.Errorf("OGImage URL: %q", xc.OpenGraphImage[0].URL)
	}
	if xc.OpenGraphImage[0].SecureURL != "http://example.com/simg.jpg" {
		t.Errorf("OGImage SecureURL: %q", xc.OpenGraphImage[0].SecureURL)
	}
	if xc.OpenGraphVideo[0].URL != "http://example.com/vid.mp4" {
		t.Errorf("OGVideo URL: %q", xc.OpenGraphVideo[0].URL)
	}
	if xc.OpenGraphVideo[0].SecureURL != "http://example.com/svid.mp4" {
		t.Errorf("OGVideo SecureURL: %q", xc.OpenGraphVideo[0].SecureURL)
	}
	if xc.OpenGraphAudio[0].URL != "http://example.com/audio.mp3" {
		t.Errorf("OGAudio URL: %q", xc.OpenGraphAudio[0].URL)
	}
	if xc.OpenGraphAudio[0].SecureURL != "http://example.com/saudio.mp3" {
		t.Errorf("OGAudio SecureURL: %q", xc.OpenGraphAudio[0].SecureURL)
	}
	if xc.XCardsImage[0].URL != "http://example.com/ximg.jpg" {
		t.Errorf("XCardsImage URL: %q", xc.XCardsImage[0].URL)
	}
	if xc.XCardsImage[0].SecureURL != "http://example.com/sximg.jpg" {
		t.Errorf("XCardsImage SecureURL: %q", xc.XCardsImage[0].SecureURL)
	}
	if xc.XCardsVideo[0].URL != "http://example.com/xvid.mp4" {
		t.Errorf("XCardsVideo URL: %q", xc.XCardsVideo[0].URL)
	}
	if xc.XCardsVideo[0].SecureURL != "http://example.com/sxvid.mp4" {
		t.Errorf("XCardsVideo SecureURL: %q", xc.XCardsVideo[0].SecureURL)
	}
	if xc.XCardsAudio[0].URL != "http://example.com/xaudio.mp3" {
		t.Errorf("XCardsAudio URL: %q", xc.XCardsAudio[0].URL)
	}
}

func TestFillMissingFieldsFromOpenGraph_NilTarget(t *testing.T) {
	errs := FillMissingFieldsFromOpenGraph(nil, &OpenGraph{})
	if len(errs) == 0 {
		t.Error("expected error for nil target")
	}
}

func TestFillMissingFieldsFromOpenGraph_NilSource(t *testing.T) {
	errs := FillMissingFieldsFromOpenGraph(&XCards{}, nil)
	if len(errs) == 0 {
		t.Error("expected error for nil source")
	}
}

func TestFillMissingFieldsFromOpenGraph_BothNil(t *testing.T) {
	errs := FillMissingFieldsFromOpenGraph(nil, nil)
	if len(errs) == 0 {
		t.Error("expected errors for both nil")
	}
}

func TestFillMissingFieldsFromOpenGraph_StringFieldCopy(t *testing.T) {
	og := &OpenGraph{Title: "OG Title", Type: "website"}
	xc := NewXCards()
	errs := FillMissingFieldsFromOpenGraph(xc, og)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if xc.Title != "OG Title" {
		t.Errorf("Title: %q", xc.Title)
	}
}

func TestFillMissingFieldsFromOpenGraph_StringFieldNoOverwrite(t *testing.T) {
	og := &OpenGraph{Title: "OG Title"}
	xc := &XCards{Title: "XC Title"}
	FillMissingFieldsFromOpenGraph(xc, og)
	if xc.Title != "XC Title" {
		t.Errorf("Title should not be overwritten: %q", xc.Title)
	}
}

func TestFillMissingFieldsFromOpenGraph_PtrFieldCopy(t *testing.T) {
	// target Article is nil, source has Article → should copy pointer
	og := &OpenGraph{
		Article: &Article{Section: "Technology"},
	}
	xc := NewXCards()
	FillMissingFieldsFromOpenGraph(xc, og)
	if xc.Article == nil {
		t.Fatal("expected Article to be copied")
	}
	if xc.Article.Section != "Technology" {
		t.Errorf("Section: %q", xc.Article.Section)
	}
}

func TestFillMissingFieldsFromOpenGraph_PtrFieldRecurse(t *testing.T) {
	// Both target and source have non-nil Article → recurse
	og := &OpenGraph{
		Article: &Article{Section: "Tech", Author: []string{"Alice"}},
	}
	xc := &XCards{
		Article: &Article{Section: "Existing"},
	}
	FillMissingFieldsFromOpenGraph(xc, og)
	// Section should NOT be overwritten (non-empty)
	if xc.Article.Section != "Existing" {
		t.Errorf("Section should not be overwritten: %q", xc.Article.Section)
	}
	// Author should be filled from OG (nil slice)
	if len(xc.Article.Author) != 1 {
		t.Errorf("Author: %v", xc.Article.Author)
	}
}

func TestFillMissingFieldsFromOpenGraph_SliceFieldCopy(t *testing.T) {
	// Target slice is nil, source has items → copy
	og := &OpenGraph{
		OpenGraphImage: []OpenGraphImage{{URL: "/img.jpg"}},
	}
	xc := NewXCards()
	FillMissingFieldsFromOpenGraph(xc, og)
	if len(xc.OpenGraphImage) != 1 {
		t.Errorf("OpenGraphImage: %v", xc.OpenGraphImage)
	}
}

func TestFillMissingFieldsFromOpenGraph_StructFieldRecurse(t *testing.T) {
	// Struct fields (non-pointer) trigger recursive FillMissingFields.
	// Article is a pointer, but time.Time (embedded in Article) is a struct.
	// We can trigger the struct branch via an Article with time fields.
	og := &OpenGraph{
		Article: &Article{Section: "Tech"},
	}
	xc := &XCards{
		Article: &Article{},
	}
	errs := FillMissingFieldsFromOpenGraph(xc, og)
	// The struct branch causes recursion into Article's time.Time fields
	// which themselves have int/uint64 fields that hit the default:continue branch
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if xc.Article.Section != "Tech" {
		t.Errorf("Section: %q", xc.Article.Section)
	}
}

// sourceWithExtra has a field (Extra) not present in targetWithoutExtra.
// This is used to trigger the !tField.IsValid() → continue path in FillMissingFieldsFromOpenGraph.
type sourceWithExtra struct {
	Name  string
	Extra string // this field is not in targetWithoutExtra
}

type targetWithoutExtra struct {
	Name string
}

func TestFillMissingFieldsFromOpenGraph_FieldNotInTarget(t *testing.T) {
	// source has field "Extra" that target doesn't have → !tField.IsValid() → continue
	src := &sourceWithExtra{Name: "source-name", Extra: "extra-value"}
	tgt := &targetWithoutExtra{}
	errs := FillMissingFieldsFromOpenGraph(tgt, src)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if tgt.Name != "source-name" {
		t.Errorf("Name: %q", tgt.Name)
	}
}
