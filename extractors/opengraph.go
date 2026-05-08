package extractor

import (
	"golang.org/x/net/html"
	"io"
	"strings"
	"time"
)

type OpenGraph struct {
	// Basic metadata
	Type  string `json:"og:type"`
	Title string `json:"og:title"`
	URL   string `json:"og:url"`

	// Optional metadata
	Description     string   `json:"og:description,omitempty"`
	Determiner      string   `json:"og:determiner,omitempty"`
	Locale          string   `json:"og:locale,omitempty"`
	LocaleAlternate []string `json:"og:locale:alternate,omitempty"`
	SiteName        string   `json:"og:site_name,omitempty"`

	// Media
	OpenGraphImage []OpenGraphImage `json:"og:image,omitempty"`
	OpenGraphVideo []OpenGraphVideo `json:"og:video,omitempty"`
	OpenGraphAudio []OpenGraphAudio `json:"og:audio,omitempty"`

	// Music specific
	Music *Music `json:"music,omitempty"`

	// Video specific
	Video *Video `json:"video,omitempty"`

	// Article specific
	Article *Article `json:"article,omitempty"`

	// Book specific
	Book *Book `json:"book,omitempty"`

	// Profile specific
	Profile *Profile `json:"profile,omitempty"`
}

// OpenGraphImage represents OpenGraph image object
type OpenGraphImage struct {
	URL       string `json:"og:image"`
	SecureURL string `json:"og:image:secure_url,omitempty"`
	Type      string `json:"og:image:type,omitempty"`
	Width     int    `json:"og:image:width,omitempty"`
	Height    int    `json:"og:image:height,omitempty"`
	Alt       string `json:"og:image:alt,omitempty"`
}

// OpenGraphVideo represents OpenGraph video object
type OpenGraphVideo struct {
	URL       string `json:"og:video"`
	SecureURL string `json:"og:video:secure_url,omitempty"`
	Type      string `json:"og:video:type,omitempty"`
	Width     int    `json:"og:video:width,omitempty"`
	Height    int    `json:"og:video:height,omitempty"`
}

// OpenGraphAudio represents OpenGraph audio object
type OpenGraphAudio struct {
	URL       string `json:"og:audio"`
	SecureURL string `json:"og:audio:secure_url,omitempty"`
	Type      string `json:"og:audio:type,omitempty"`
}

// Music represents music-specific metadata
type Music struct {
	Duration    int         `json:"music:duration,omitempty"`
	Album       string      `json:"music:album,omitempty"`
	AlbumDisc   int         `json:"music:album:disc,omitempty"`
	AlbumTrack  int         `json:"music:album:track,omitempty"`
	Musician    []string    `json:"music:musician,omitempty"`
	Song        []MusicSong `json:"music:song,omitempty"`
	Creator     []string    `json:"music:creator,omitempty"`
	ReleaseDate string      `json:"music:release_date,omitempty"`
}

type MusicSong struct {
	URL   string `json:"url,omitempty"`
	Disc  int    `json:"disc,omitempty"`
	Track int    `json:"track,omitempty"`
}

type Video struct {
	Duration    int          `json:"video:duration,omitempty"`
	Actor       []VideoActor `json:"video:actor,omitempty"`
	Director    []string     `json:"video:director,omitempty"`
	Writer      []string     `json:"video:writer,omitempty"`
	ReleaseDate time.Time    `json:"video:release_date,omitempty"`
	Tag         []string     `json:"video:tag,omitempty"`
	Series      string       `json:"video:series,omitempty"`
}

type VideoActor struct {
	URL  string `json:"url,omitempty"`
	Role string `json:"role,omitempty"`
}

// Article represents article-specific metadata
type Article struct {
	PublishedTime  time.Time `json:"article:published_time,omitempty"`
	ModifiedTime   time.Time `json:"article:modified_time,omitempty"`
	ExpirationTime time.Time `json:"article:expiration_time,omitempty"`
	Author         []string  `json:"article:author,omitempty"`
	Section        string    `json:"article:section,omitempty"`
	Tag            []string  `json:"article:tag,omitempty"`
}

// Book represents book-specific metadata
type Book struct {
	Author      []string  `json:"book:author,omitempty"`
	ISBN        string    `json:"book:isbn,omitempty"`
	ReleaseDate time.Time `json:"book:release_date,omitempty"`
	Tag         []string  `json:"book:tag,omitempty"`
}

// Profile represents profile-specific metadata
type Profile struct {
	FirstName string `json:"profile:first_name,omitempty"`
	LastName  string `json:"profile:last_name,omitempty"`
	Username  string `json:"profile:username,omitempty"`
	Gender    string `json:"profile:gender,omitempty"`
}

// NewOpenGraph creates a new OpenGraph instance with basic initialization
func NewOpenGraph() *OpenGraph {
	return &OpenGraph{}
}

func ParseOpenGraph(URL string, htmlContent string) (*OpenGraph, []error) {
	item, errors := extractOpenGraph(htmlContent)
	if item != nil && URL != "" {
		resolveOpenGraphURLs(item, URL)
	}
	return item, errors
}

func resolveOpenGraphURLs(og *OpenGraph, base string) {
	og.URL = resolveURL(base, og.URL)
	for i := range og.OpenGraphImage {
		og.OpenGraphImage[i].URL = resolveURL(base, og.OpenGraphImage[i].URL)
		og.OpenGraphImage[i].SecureURL = resolveURL(base, og.OpenGraphImage[i].SecureURL)
	}
	for i := range og.OpenGraphVideo {
		og.OpenGraphVideo[i].URL = resolveURL(base, og.OpenGraphVideo[i].URL)
		og.OpenGraphVideo[i].SecureURL = resolveURL(base, og.OpenGraphVideo[i].SecureURL)
	}
	for i := range og.OpenGraphAudio {
		og.OpenGraphAudio[i].URL = resolveURL(base, og.OpenGraphAudio[i].URL)
		og.OpenGraphAudio[i].SecureURL = resolveURL(base, og.OpenGraphAudio[i].SecureURL)
	}
}

func extractOpenGraph(htmlContent string) (*OpenGraph, []error) {
	var errors []error

	og := NewOpenGraph()
	tokenizer := html.NewTokenizer(strings.NewReader(htmlContent))

	ogHasValue := false
	for {
		if tokenizer.Err() == io.EOF {
			break
		}
		tokenType := tokenizer.Next()
		switch tokenType {
		case html.ErrorToken:
			if tokenizer.Err() == io.EOF {
				break
			}
		case html.StartTagToken, html.SelfClosingTagToken, html.EndTagToken:
			token := tokenizer.Token()
			if token.Data != "meta" || token.Attr == nil {
				continue
			}

			var property, content string
			for _, attr := range token.Attr {
				switch attr.Key {
				case "property":
					property = attr.Val
				case "content":
					content = attr.Val
				}
			}
			if property != "" && content != "" {
				parseOpenGraphMetaTag(og, property, content)
				ogHasValue = true
			}
		default:
			continue
		}
	}

	if ogHasValue {
		return og, errors
	}

	return nil, errors
}

func parseOpenGraphMetaTag(og *OpenGraph, property, content string) {
	parts := strings.Split(property, ":")

	switch {
	// Basic metadata
	case property == "og:type":
		og.Type = content
	case property == "og:title":
		og.Title = content
	case property == "og:url":
		og.URL = content

	// Optional metadata
	case property == "og:description":
		og.Description = content
	case property == "og:determiner":
		og.Determiner = content
	case property == "og:locale":
		og.Locale = content
	case property == "og:locale:alternate":
		og.LocaleAlternate = append(og.LocaleAlternate, content)
	case property == "og:site_name":
		og.SiteName = content

	// Image handling with multi-level properties
	case strings.HasPrefix(property, "og:image"):
		handleOpenGraphImageProperty(og, parts, content)

	// Video handling with multi-level properties
	case strings.HasPrefix(property, "og:video"):
		handleOpenGraphVideoProperty(og, parts, content)

	// Audio handling with multi-level properties
	case strings.HasPrefix(property, "og:audio"):
		handleOpenGraphAudioProperty(og, parts, content)

	// Music handling with multi-level properties
	case strings.HasPrefix(property, "music:"):
		parseMusicProperty(&og.Music, parts, property, content)

	// Video object handling with multi-level properties
	case strings.HasPrefix(property, "video:"):
		parseVideoObjectProperty(&og.Video, parts, property, content)

	// Article handling
	case strings.HasPrefix(property, "article:"):
		parseArticleProperty(&og.Article, property, content)

	// Book handling
	case strings.HasPrefix(property, "book:"):
		parseBookProperty(&og.Book, property, content)

	// Profile handling
	case strings.HasPrefix(property, "profile:"):
		parseProfileProperty(&og.Profile, property, content)
	}
}

func handleOpenGraphImageProperty(og *OpenGraph, parts []string, content string) {
	handleMediaSlice(&og.OpenGraphImage, "image", parts, content, func(img *OpenGraphImage, sub, val string) {
		switch sub {
		case "":
			img.URL = val
		case "secure_url":
			img.SecureURL = val
		case "type":
			img.Type = val
		case "width":
			img.Width = parseIntSafely(val)
		case "height":
			img.Height = parseIntSafely(val)
		case "alt":
			img.Alt = val
		}
	})
}

func handleOpenGraphVideoProperty(og *OpenGraph, parts []string, content string) {
	handleMediaSlice(&og.OpenGraphVideo, "video", parts, content, func(v *OpenGraphVideo, sub, val string) {
		switch sub {
		case "":
			v.URL = val
		case "secure_url":
			v.SecureURL = val
		case "type":
			v.Type = val
		case "width":
			v.Width = parseIntSafely(val)
		case "height":
			v.Height = parseIntSafely(val)
		}
	})
}

func handleOpenGraphAudioProperty(og *OpenGraph, parts []string, content string) {
	handleMediaSlice(&og.OpenGraphAudio, "audio", parts, content, func(a *OpenGraphAudio, sub, val string) {
		switch sub {
		case "":
			a.URL = val
		case "secure_url":
			a.SecureURL = val
		case "type":
			a.Type = val
		}
	})
}
