package extractor

import (
	"fmt"
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

func ParseOpenGraph(URL string, htmlContent string) (any, []error) {
	_ = URL
	item, errors := extractOpenGraph(htmlContent)

	var results any
	if item != nil {
		results = item
	}

	return results, errors
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
			errors = append(errors, tokenizer.Err())
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
	// Split property into parts to handle multi-level properties
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
		if og.Music == nil {
			og.Music = &Music{}
		}

		switch {
		case property == "music:duration":
			og.Music.Duration = parseIntSafely(content)
		case property == "music:album":
			og.Music.Album = content
		case property == "music:album:disc":
			og.Music.AlbumDisc = parseIntSafely(content)
		case property == "music:album:track":
			og.Music.AlbumTrack = parseIntSafely(content)
		case property == "music:musician":
			og.Music.Musician = append(og.Music.Musician, content)
		case strings.HasPrefix(property, "music:song"):
			handleMusicSongProperty(og.Music, parts, content)
		case property == "music:release_date":
			og.Music.ReleaseDate = content
		case property == "music:creator":
			og.Music.Creator = append(og.Music.Creator, content)
		}

	// Video handling with multi-level properties
	case strings.HasPrefix(property, "video:"):
		if og.Video == nil {
			og.Video = &Video{}
		}

		switch {
		case strings.HasPrefix(property, "video:actor"):
			handleVideoActorProperty(og.Video, parts, content)
		case property == "video:director":
			og.Video.Director = append(og.Video.Director, content)
		case property == "video:writer":
			og.Video.Writer = append(og.Video.Writer, content)
		case property == "video:duration":
			og.Video.Duration = parseIntSafely(content)
		case property == "video:release_date":
			og.Video.ReleaseDate = parseTimeSafely(content)
		case property == "video:tag":
			og.Video.Tag = append(og.Video.Tag, content)
		case property == "video:series":
			og.Video.Series = content
		}

	// Article handling remains the same
	case strings.HasPrefix(property, "article:"):
		if og.Article == nil {
			og.Article = &Article{}
		}
		switch property {
		case "article:published_time":
			og.Article.PublishedTime = parseTimeSafely(content)
		case "article:modified_time":
			og.Article.ModifiedTime = parseTimeSafely(content)
		case "article:expiration_time":
			og.Article.ExpirationTime = parseTimeSafely(content)
		case "article:author":
			og.Article.Author = append(og.Article.Author, content)
		case "article:section":
			og.Article.Section = content
		case "article:tag":
			og.Article.Tag = append(og.Article.Tag, content)
		}

	// Book handling remains the same
	case strings.HasPrefix(property, "book:"):
		if og.Book == nil {
			og.Book = &Book{}
		}
		switch property {
		case "book:isbn":
			og.Book.ISBN = content
		case "book:release_date":
			og.Book.ReleaseDate = parseTimeSafely(content)
		case "book:author":
			og.Book.Author = append(og.Book.Author, content)
		case "book:tag":
			og.Book.Tag = append(og.Book.Tag, content)
		}

	// Profile handling remains the same
	case strings.HasPrefix(property, "profile:"):
		if og.Profile == nil {
			og.Profile = &Profile{}
		}
		switch property {
		case "profile:first_name":
			og.Profile.FirstName = content
		case "profile:last_name":
			og.Profile.LastName = content
		case "profile:username":
			og.Profile.Username = content
		case "profile:gender":
			og.Profile.Gender = content
		}
	}
}

func handleOpenGraphImageProperty(og *OpenGraph, parts []string, content string) {
	if len(og.OpenGraphImage) == 0 || parts[1] == "image" {
		if len(parts) < 3 {
			og.OpenGraphImage = append(og.OpenGraphImage, OpenGraphImage{})
		}
	}
	lastIdx := len(og.OpenGraphImage) - 1

	if len(parts) == 2 {
		og.OpenGraphImage[lastIdx].URL = content
		return
	}

	switch parts[2] {
	case "secure_url":
		og.OpenGraphImage[lastIdx].SecureURL = content
	case "type":
		og.OpenGraphImage[lastIdx].Type = content
	case "width":
		og.OpenGraphImage[lastIdx].Width = parseIntSafely(content)
	case "height":
		og.OpenGraphImage[lastIdx].Height = parseIntSafely(content)
	case "alt":
		og.OpenGraphImage[lastIdx].Alt = content
	}
}

func handleOpenGraphVideoProperty(og *OpenGraph, parts []string, content string) {
	if len(og.OpenGraphVideo) == 0 || parts[1] == "video" {
		if len(parts) < 3 {
			og.OpenGraphVideo = append(og.OpenGraphVideo, OpenGraphVideo{})
		}
	}
	lastIdx := len(og.OpenGraphVideo) - 1

	if len(parts) == 2 {
		og.OpenGraphVideo[lastIdx].URL = content
		return
	}

	switch parts[2] {
	case "secure_url":
		og.OpenGraphVideo[lastIdx].SecureURL = content
	case "type":
		og.OpenGraphVideo[lastIdx].Type = content
	case "width":
		og.OpenGraphVideo[lastIdx].Width = parseIntSafely(content)
	case "height":
		og.OpenGraphVideo[lastIdx].Height = parseIntSafely(content)
	}
}

func handleOpenGraphAudioProperty(og *OpenGraph, parts []string, content string) {
	if len(og.OpenGraphAudio) == 0 || parts[1] == "audio" {
		if len(parts) < 3 {
			og.OpenGraphAudio = append(og.OpenGraphAudio, OpenGraphAudio{})
		}
	}
	lastIdx := len(og.OpenGraphAudio) - 1

	if len(parts) == 2 {
		og.OpenGraphAudio[lastIdx].URL = content
		return
	}

	switch parts[2] {
	case "secure_url":
		og.OpenGraphAudio[lastIdx].SecureURL = content
	case "type":
		og.OpenGraphAudio[lastIdx].Type = content
	}
}

func handleMusicSongProperty(music *Music, parts []string, content string) {
	if len(music.Song) == 0 || parts[1] == "song" {
		if len(parts) < 3 {
			music.Song = append(music.Song, MusicSong{})
		}
	}
	lastIdx := len(music.Song) - 1

	if len(parts) == 2 {
		music.Song[lastIdx].URL = content
		return
	}

	switch parts[2] {
	case "disc":
		music.Song[lastIdx].Disc = parseIntSafely(content)
	case "track":
		music.Song[lastIdx].Track = parseIntSafely(content)
	}
}

func handleVideoActorProperty(video *Video, parts []string, content string) {
	if len(video.Actor) == 0 || parts[1] == "actor" {
		if len(parts) < 3 {
			video.Actor = append(video.Actor, VideoActor{})
		}
	}
	lastIdx := len(video.Actor) - 1

	if len(parts) == 2 {
		video.Actor[lastIdx].URL = content
		return
	}

	switch parts[2] {
	case "role":
		video.Actor[lastIdx].Role = content
	}
}

func parseIntSafely(s string) int {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	if err != nil {
		return 0
	}
	return result
}

func parseTimeSafely(s string) time.Time {
	// Try common date formats
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z0700",
		"2006-01-02T15:04:05",
		"2006-01-02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t
		}
	}
	return time.Time{}
}
