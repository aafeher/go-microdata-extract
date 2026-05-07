package extractor

import (
	"testing"
	"time"
)

func TestNewOpenGraph(t *testing.T) {
	og := NewOpenGraph()
	if og == nil {
		t.Fatal("NewOpenGraph returned nil")
	}
}

func TestParseOpenGraph_EmptyURL(t *testing.T) {
	// With empty URL, resolveOpenGraphURLs should NOT be called.
	html := `<meta property="og:title" content="Hello">`
	og, errs := ParseOpenGraph("", html)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if og == nil {
		t.Fatal("expected non-nil OpenGraph")
	}
	if og.Title != "Hello" {
		t.Errorf("expected Title=Hello, got %q", og.Title)
	}
}

func TestParseOpenGraph_WithURLNilResult(t *testing.T) {
	// No OG tags → nil result, URL provided but no resolution should happen.
	og, errs := ParseOpenGraph("http://example.com", `<html></html>`)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if og != nil {
		t.Errorf("expected nil OpenGraph, got %+v", og)
	}
}

func TestParseOpenGraph_WithURLNonNilResult(t *testing.T) {
	html := `<meta property="og:title" content="Test"><meta property="og:url" content="/page">`
	og, errs := ParseOpenGraph("http://example.com", html)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if og == nil {
		t.Fatal("expected non-nil OpenGraph")
	}
	if og.URL != "http://example.com/page" {
		t.Errorf("expected resolved URL, got %q", og.URL)
	}
}

func TestExtractOpenGraph_NoTags(t *testing.T) {
	// Using HTML with text nodes and comments to trigger the default:continue branch
	og, errs := extractOpenGraph(`<html><head><!-- comment --></head><body>text node</body></html>`)
	if len(errs) != 0 {
		t.Errorf("unexpected errors: %v", errs)
	}
	if og != nil {
		t.Errorf("expected nil, got %+v", og)
	}
}

func TestExtractOpenGraph_WithTags(t *testing.T) {
	html := `<meta property="og:title" content="My Title"><meta property="og:type" content="website">`
	og, errs := extractOpenGraph(html)
	if len(errs) != 0 {
		t.Errorf("unexpected errors: %v", errs)
	}
	if og == nil {
		t.Fatal("expected non-nil OpenGraph")
	}
	if og.Title != "My Title" {
		t.Errorf("expected Title=My Title, got %q", og.Title)
	}
	if og.Type != "website" {
		t.Errorf("expected Type=website, got %q", og.Type)
	}
}

func TestResolveOpenGraphURLs(t *testing.T) {
	og := &OpenGraph{
		URL: "/page",
		OpenGraphImage: []OpenGraphImage{
			{URL: "/img.jpg", SecureURL: "/secure.jpg"},
		},
		OpenGraphVideo: []OpenGraphVideo{
			{URL: "/video.mp4", SecureURL: "/secure.mp4"},
		},
		OpenGraphAudio: []OpenGraphAudio{
			{URL: "/audio.mp3", SecureURL: "/secure.mp3"},
		},
	}
	resolveOpenGraphURLs(og, "http://example.com")
	if og.URL != "http://example.com/page" {
		t.Errorf("URL not resolved: %q", og.URL)
	}
	if og.OpenGraphImage[0].URL != "http://example.com/img.jpg" {
		t.Errorf("Image URL not resolved: %q", og.OpenGraphImage[0].URL)
	}
	if og.OpenGraphImage[0].SecureURL != "http://example.com/secure.jpg" {
		t.Errorf("Image SecureURL not resolved: %q", og.OpenGraphImage[0].SecureURL)
	}
	if og.OpenGraphVideo[0].URL != "http://example.com/video.mp4" {
		t.Errorf("Video URL not resolved: %q", og.OpenGraphVideo[0].URL)
	}
	if og.OpenGraphVideo[0].SecureURL != "http://example.com/secure.mp4" {
		t.Errorf("Video SecureURL not resolved: %q", og.OpenGraphVideo[0].SecureURL)
	}
	if og.OpenGraphAudio[0].URL != "http://example.com/audio.mp3" {
		t.Errorf("Audio URL not resolved: %q", og.OpenGraphAudio[0].URL)
	}
	if og.OpenGraphAudio[0].SecureURL != "http://example.com/secure.mp3" {
		t.Errorf("Audio SecureURL not resolved: %q", og.OpenGraphAudio[0].SecureURL)
	}
}

func TestParseOpenGraphMetaTag_BasicMetadata(t *testing.T) {
	og := NewOpenGraph()
	parseOpenGraphMetaTag(og, "og:type", "article")
	parseOpenGraphMetaTag(og, "og:title", "My Title")
	parseOpenGraphMetaTag(og, "og:url", "/page")
	if og.Type != "article" {
		t.Errorf("Type: want article, got %q", og.Type)
	}
	if og.Title != "My Title" {
		t.Errorf("Title: want My Title, got %q", og.Title)
	}
	if og.URL != "/page" {
		t.Errorf("URL: want /page, got %q", og.URL)
	}
}

func TestParseOpenGraphMetaTag_OptionalMetadata(t *testing.T) {
	og := NewOpenGraph()
	parseOpenGraphMetaTag(og, "og:description", "A description")
	parseOpenGraphMetaTag(og, "og:determiner", "the")
	parseOpenGraphMetaTag(og, "og:locale", "en_US")
	parseOpenGraphMetaTag(og, "og:locale:alternate", "fr_FR")
	parseOpenGraphMetaTag(og, "og:locale:alternate", "de_DE")
	parseOpenGraphMetaTag(og, "og:site_name", "My Site")

	if og.Description != "A description" {
		t.Errorf("Description: %q", og.Description)
	}
	if og.Determiner != "the" {
		t.Errorf("Determiner: %q", og.Determiner)
	}
	if og.Locale != "en_US" {
		t.Errorf("Locale: %q", og.Locale)
	}
	if len(og.LocaleAlternate) != 2 {
		t.Errorf("LocaleAlternate: %v", og.LocaleAlternate)
	}
	if og.SiteName != "My Site" {
		t.Errorf("SiteName: %q", og.SiteName)
	}
}

func TestParseOpenGraphMetaTag_Image(t *testing.T) {
	og := NewOpenGraph()
	parseOpenGraphMetaTag(og, "og:image", "/img.jpg")
	parseOpenGraphMetaTag(og, "og:image:secure_url", "/secure.jpg")
	parseOpenGraphMetaTag(og, "og:image:type", "image/jpeg")
	parseOpenGraphMetaTag(og, "og:image:width", "800")
	parseOpenGraphMetaTag(og, "og:image:height", "600")
	parseOpenGraphMetaTag(og, "og:image:alt", "Alt text")

	if len(og.OpenGraphImage) != 1 {
		t.Fatalf("expected 1 image, got %d", len(og.OpenGraphImage))
	}
	img := og.OpenGraphImage[0]
	if img.URL != "/img.jpg" {
		t.Errorf("Image URL: %q", img.URL)
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

	// Second og:image tag creates a new entry
	parseOpenGraphMetaTag(og, "og:image", "/img2.jpg")
	if len(og.OpenGraphImage) != 2 {
		t.Fatalf("expected 2 images after second og:image, got %d", len(og.OpenGraphImage))
	}
}

func TestParseOpenGraphMetaTag_Video(t *testing.T) {
	og := NewOpenGraph()
	parseOpenGraphMetaTag(og, "og:video", "/video.mp4")
	parseOpenGraphMetaTag(og, "og:video:secure_url", "/secure.mp4")
	parseOpenGraphMetaTag(og, "og:video:type", "video/mp4")
	parseOpenGraphMetaTag(og, "og:video:width", "1920")
	parseOpenGraphMetaTag(og, "og:video:height", "1080")

	if len(og.OpenGraphVideo) != 1 {
		t.Fatalf("expected 1 video, got %d", len(og.OpenGraphVideo))
	}
	v := og.OpenGraphVideo[0]
	if v.URL != "/video.mp4" {
		t.Errorf("Video URL: %q", v.URL)
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

	// Second og:video
	parseOpenGraphMetaTag(og, "og:video", "/video2.mp4")
	if len(og.OpenGraphVideo) != 2 {
		t.Fatalf("expected 2 videos, got %d", len(og.OpenGraphVideo))
	}
}

func TestParseOpenGraphMetaTag_Audio(t *testing.T) {
	og := NewOpenGraph()
	parseOpenGraphMetaTag(og, "og:audio", "/audio.mp3")
	parseOpenGraphMetaTag(og, "og:audio:secure_url", "/secure.mp3")
	parseOpenGraphMetaTag(og, "og:audio:type", "audio/mpeg")

	if len(og.OpenGraphAudio) != 1 {
		t.Fatalf("expected 1 audio, got %d", len(og.OpenGraphAudio))
	}
	a := og.OpenGraphAudio[0]
	if a.URL != "/audio.mp3" {
		t.Errorf("Audio URL: %q", a.URL)
	}
	if a.SecureURL != "/secure.mp3" {
		t.Errorf("SecureURL: %q", a.SecureURL)
	}
	if a.Type != "audio/mpeg" {
		t.Errorf("Type: %q", a.Type)
	}

	// Second og:audio
	parseOpenGraphMetaTag(og, "og:audio", "/audio2.mp3")
	if len(og.OpenGraphAudio) != 2 {
		t.Fatalf("expected 2 audios, got %d", len(og.OpenGraphAudio))
	}
}

func TestParseOpenGraphMetaTag_Music(t *testing.T) {
	og := NewOpenGraph()
	parseOpenGraphMetaTag(og, "music:duration", "240")
	parseOpenGraphMetaTag(og, "music:album", "Greatest Hits")
	parseOpenGraphMetaTag(og, "music:album:disc", "1")
	parseOpenGraphMetaTag(og, "music:album:track", "3")
	parseOpenGraphMetaTag(og, "music:musician", "http://example.com/artist")
	parseOpenGraphMetaTag(og, "music:song", "http://example.com/song1")
	parseOpenGraphMetaTag(og, "music:song:disc", "1")
	parseOpenGraphMetaTag(og, "music:song:track", "2")
	parseOpenGraphMetaTag(og, "music:release_date", "2023-01-15")
	parseOpenGraphMetaTag(og, "music:creator", "http://example.com/creator")

	if og.Music == nil {
		t.Fatal("Music is nil")
	}
	if og.Music.Duration != 240 {
		t.Errorf("Duration: %d", og.Music.Duration)
	}
	if og.Music.Album != "Greatest Hits" {
		t.Errorf("Album: %q", og.Music.Album)
	}
	if og.Music.AlbumDisc != 1 {
		t.Errorf("AlbumDisc: %d", og.Music.AlbumDisc)
	}
	if og.Music.AlbumTrack != 3 {
		t.Errorf("AlbumTrack: %d", og.Music.AlbumTrack)
	}
	if len(og.Music.Musician) != 1 {
		t.Errorf("Musician: %v", og.Music.Musician)
	}
	if len(og.Music.Song) != 1 {
		t.Errorf("Song: %v", og.Music.Song)
	}
	if og.Music.Song[0].URL != "http://example.com/song1" {
		t.Errorf("Song URL: %q", og.Music.Song[0].URL)
	}
	if og.Music.Song[0].Disc != 1 {
		t.Errorf("Song Disc: %d", og.Music.Song[0].Disc)
	}
	if og.Music.Song[0].Track != 2 {
		t.Errorf("Song Track: %d", og.Music.Song[0].Track)
	}
	if og.Music.ReleaseDate != "2023-01-15" {
		t.Errorf("ReleaseDate: %q", og.Music.ReleaseDate)
	}
	if len(og.Music.Creator) != 1 {
		t.Errorf("Creator: %v", og.Music.Creator)
	}
}

func TestParseOpenGraphMetaTag_VideoSpecific(t *testing.T) {
	og := NewOpenGraph()
	parseOpenGraphMetaTag(og, "video:actor", "http://example.com/actor1")
	parseOpenGraphMetaTag(og, "video:actor:role", "Hero")
	parseOpenGraphMetaTag(og, "video:director", "http://example.com/director1")
	parseOpenGraphMetaTag(og, "video:writer", "http://example.com/writer1")
	parseOpenGraphMetaTag(og, "video:duration", "120")
	parseOpenGraphMetaTag(og, "video:release_date", "2023-06-15")
	parseOpenGraphMetaTag(og, "video:tag", "action")
	parseOpenGraphMetaTag(og, "video:tag", "thriller")
	parseOpenGraphMetaTag(og, "video:series", "http://example.com/series")

	if og.Video == nil {
		t.Fatal("Video is nil")
	}
	if len(og.Video.Actor) != 1 {
		t.Fatalf("expected 1 actor, got %d", len(og.Video.Actor))
	}
	if og.Video.Actor[0].URL != "http://example.com/actor1" {
		t.Errorf("Actor URL: %q", og.Video.Actor[0].URL)
	}
	if og.Video.Actor[0].Role != "Hero" {
		t.Errorf("Actor Role: %q", og.Video.Actor[0].Role)
	}
	if len(og.Video.Director) != 1 {
		t.Errorf("Director: %v", og.Video.Director)
	}
	if len(og.Video.Writer) != 1 {
		t.Errorf("Writer: %v", og.Video.Writer)
	}
	if og.Video.Duration != 120 {
		t.Errorf("Duration: %d", og.Video.Duration)
	}
	if og.Video.ReleaseDate.IsZero() {
		t.Error("ReleaseDate is zero")
	}
	if len(og.Video.Tag) != 2 {
		t.Errorf("Tag: %v", og.Video.Tag)
	}
	if og.Video.Series != "http://example.com/series" {
		t.Errorf("Series: %q", og.Video.Series)
	}
}

func TestParseOpenGraphMetaTag_Article(t *testing.T) {
	og := NewOpenGraph()
	parseOpenGraphMetaTag(og, "article:published_time", "2023-01-01T00:00:00Z")
	parseOpenGraphMetaTag(og, "article:modified_time", "2023-02-01T00:00:00Z")
	parseOpenGraphMetaTag(og, "article:expiration_time", "2024-01-01T00:00:00Z")
	parseOpenGraphMetaTag(og, "article:author", "http://example.com/author")
	parseOpenGraphMetaTag(og, "article:section", "Technology")
	parseOpenGraphMetaTag(og, "article:tag", "go")
	parseOpenGraphMetaTag(og, "article:tag", "programming")

	if og.Article == nil {
		t.Fatal("Article is nil")
	}
	if og.Article.PublishedTime.IsZero() {
		t.Error("PublishedTime is zero")
	}
	if og.Article.ModifiedTime.IsZero() {
		t.Error("ModifiedTime is zero")
	}
	if og.Article.ExpirationTime.IsZero() {
		t.Error("ExpirationTime is zero")
	}
	if len(og.Article.Author) != 1 {
		t.Errorf("Author: %v", og.Article.Author)
	}
	if og.Article.Section != "Technology" {
		t.Errorf("Section: %q", og.Article.Section)
	}
	if len(og.Article.Tag) != 2 {
		t.Errorf("Tag: %v", og.Article.Tag)
	}
}

func TestParseOpenGraphMetaTag_Book(t *testing.T) {
	og := NewOpenGraph()
	parseOpenGraphMetaTag(og, "book:author", "http://example.com/author")
	parseOpenGraphMetaTag(og, "book:isbn", "978-3-16-148410-0")
	parseOpenGraphMetaTag(og, "book:release_date", "2023-03-15")
	parseOpenGraphMetaTag(og, "book:tag", "fiction")

	if og.Book == nil {
		t.Fatal("Book is nil")
	}
	if len(og.Book.Author) != 1 {
		t.Errorf("Author: %v", og.Book.Author)
	}
	if og.Book.ISBN != "978-3-16-148410-0" {
		t.Errorf("ISBN: %q", og.Book.ISBN)
	}
	if og.Book.ReleaseDate.IsZero() {
		t.Error("ReleaseDate is zero")
	}
	if len(og.Book.Tag) != 1 {
		t.Errorf("Tag: %v", og.Book.Tag)
	}
}

func TestParseOpenGraphMetaTag_Profile(t *testing.T) {
	og := NewOpenGraph()
	parseOpenGraphMetaTag(og, "profile:first_name", "John")
	parseOpenGraphMetaTag(og, "profile:last_name", "Doe")
	parseOpenGraphMetaTag(og, "profile:username", "johndoe")
	parseOpenGraphMetaTag(og, "profile:gender", "male")

	if og.Profile == nil {
		t.Fatal("Profile is nil")
	}
	if og.Profile.FirstName != "John" {
		t.Errorf("FirstName: %q", og.Profile.FirstName)
	}
	if og.Profile.LastName != "Doe" {
		t.Errorf("LastName: %q", og.Profile.LastName)
	}
	if og.Profile.Username != "johndoe" {
		t.Errorf("Username: %q", og.Profile.Username)
	}
	if og.Profile.Gender != "male" {
		t.Errorf("Gender: %q", og.Profile.Gender)
	}
}

func TestHandleOpenGraphImageProperty_SecondImageCreatesNewEntry(t *testing.T) {
	og := NewOpenGraph()
	// First image
	og.OpenGraphImage = append(og.OpenGraphImage, OpenGraphImage{URL: "/img1.jpg"})
	// Sub-property on existing entry (parts[1]=="image" but len(parts)>=3 and len>0 so no new entry)
	parts := []string{"og", "image", "alt"}
	handleOpenGraphImageProperty(og, parts, "Alt 1")
	if len(og.OpenGraphImage) != 1 {
		t.Errorf("expected 1 image, got %d", len(og.OpenGraphImage))
	}
	if og.OpenGraphImage[0].Alt != "Alt 1" {
		t.Errorf("Alt: %q", og.OpenGraphImage[0].Alt)
	}

	// Second og:image (parts[1]=="image", len(parts)<3 → new entry)
	parts2 := []string{"og", "image"}
	handleOpenGraphImageProperty(og, parts2, "/img2.jpg")
	if len(og.OpenGraphImage) != 2 {
		t.Fatalf("expected 2 images, got %d", len(og.OpenGraphImage))
	}
	if og.OpenGraphImage[1].URL != "/img2.jpg" {
		t.Errorf("Image[1].URL: %q", og.OpenGraphImage[1].URL)
	}
}

func TestHandleOpenGraphVideoProperty_AllSubProperties(t *testing.T) {
	og := NewOpenGraph()
	parts := []string{"og", "video"}
	handleOpenGraphVideoProperty(og, parts, "/video.mp4")
	if len(og.OpenGraphVideo) != 1 {
		t.Fatalf("expected 1 video")
	}

	// Sub-properties
	handleOpenGraphVideoProperty(og, []string{"og", "video", "secure_url"}, "/secure.mp4")
	handleOpenGraphVideoProperty(og, []string{"og", "video", "type"}, "video/mp4")
	handleOpenGraphVideoProperty(og, []string{"og", "video", "width"}, "1280")
	handleOpenGraphVideoProperty(og, []string{"og", "video", "height"}, "720")
	v := og.OpenGraphVideo[0]
	if v.SecureURL != "/secure.mp4" {
		t.Errorf("SecureURL: %q", v.SecureURL)
	}
	if v.Type != "video/mp4" {
		t.Errorf("Type: %q", v.Type)
	}
	if v.Width != 1280 {
		t.Errorf("Width: %d", v.Width)
	}
	if v.Height != 720 {
		t.Errorf("Height: %d", v.Height)
	}
}

func TestHandleOpenGraphAudioProperty_AllSubProperties(t *testing.T) {
	og := NewOpenGraph()
	handleOpenGraphAudioProperty(og, []string{"og", "audio"}, "/audio.mp3")
	if len(og.OpenGraphAudio) != 1 {
		t.Fatalf("expected 1 audio")
	}
	handleOpenGraphAudioProperty(og, []string{"og", "audio", "secure_url"}, "/secure.mp3")
	handleOpenGraphAudioProperty(og, []string{"og", "audio", "type"}, "audio/mpeg")
	a := og.OpenGraphAudio[0]
	if a.SecureURL != "/secure.mp3" {
		t.Errorf("SecureURL: %q", a.SecureURL)
	}
	if a.Type != "audio/mpeg" {
		t.Errorf("Type: %q", a.Type)
	}
}

func TestHandleMusicSongProperty(t *testing.T) {
	music := &Music{}
	// first song (len(parts)<3 creates new entry)
	handleMusicSongProperty(music, []string{"music", "song"}, "http://example.com/s1")
	if len(music.Song) != 1 {
		t.Fatalf("expected 1 song, got %d", len(music.Song))
	}
	if music.Song[0].URL != "http://example.com/s1" {
		t.Errorf("Song URL: %q", music.Song[0].URL)
	}

	// disc and track for that song
	handleMusicSongProperty(music, []string{"music", "song", "disc"}, "1")
	handleMusicSongProperty(music, []string{"music", "song", "track"}, "5")
	if music.Song[0].Disc != 1 {
		t.Errorf("Disc: %d", music.Song[0].Disc)
	}
	if music.Song[0].Track != 5 {
		t.Errorf("Track: %d", music.Song[0].Track)
	}
}

func TestHandleVideoActorProperty(t *testing.T) {
	video := &Video{}
	// first actor
	handleVideoActorProperty(video, []string{"video", "actor"}, "http://example.com/actor1")
	if len(video.Actor) != 1 {
		t.Fatalf("expected 1 actor, got %d", len(video.Actor))
	}
	if video.Actor[0].URL != "http://example.com/actor1" {
		t.Errorf("Actor URL: %q", video.Actor[0].URL)
	}

	// role
	handleVideoActorProperty(video, []string{"video", "actor", "role"}, "Villain")
	if video.Actor[0].Role != "Villain" {
		t.Errorf("Actor Role: %q", video.Actor[0].Role)
	}
}

func TestParseIntSafely(t *testing.T) {
	if v := parseIntSafely("42"); v != 42 {
		t.Errorf("expected 42, got %d", v)
	}
	if v := parseIntSafely("abc"); v != 0 {
		t.Errorf("expected 0, got %d", v)
	}
}

func TestParseTimeSafely(t *testing.T) {
	// RFC3339 format
	t1 := parseTimeSafely("2023-01-15T10:30:00Z")
	if t1.IsZero() {
		t.Error("RFC3339 time should not be zero")
	}
	if t1.Year() != 2023 {
		t.Errorf("expected year 2023, got %d", t1.Year())
	}

	// Date-only format
	t2 := parseTimeSafely("2023-01-15")
	if t2.IsZero() {
		t.Error("date-only time should not be zero")
	}
	if t2.Year() != 2023 {
		t.Errorf("expected year 2023, got %d", t2.Year())
	}

	// Invalid format
	t3 := parseTimeSafely("not-a-date")
	if !t3.Equal(time.Time{}) {
		t.Errorf("expected zero time, got %v", t3)
	}
}
