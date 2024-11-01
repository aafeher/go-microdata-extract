package extract

import (
	"fmt"
	"golang.org/x/net/html"
	"io"
	"reflect"
	"strings"
)

type XCards struct {
	// X specific metadata
	Card    string `json:"twitter:card,omitempty"`
	Site    string `json:"twitter:site,omitempty"`
	Creator string `json:"twitter:creator,omitempty"`

	// Basic Metadata
	Type  string `json:"twitter:type,omitempty"`
	Title string `json:"twitter:title,omitempty"`
	URL   string `json:"twitter:url,omitempty"`

	// Optional metadata
	Description     string   `json:"twitter:description,omitempty"`
	Determiner      string   `json:"twitter:determiner,omitempty"`
	Locale          string   `json:"twitter:locale,omitempty"`
	LocaleAlternate []string `json:"twitter:locale:alternate,omitempty"`
	SiteName        string   `json:"twitter:site_name,omitempty"`

	// Media
	OpenGraphImage []OpenGraphImage `json:"og:image,omitempty"`
	OpenGraphAudio []OpenGraphAudio `json:"og:audio,omitempty"`
	OpenGraphVideo []OpenGraphVideo `json:"og:video,omitempty"`
	XCardsImage    []XCardsImage    `json:"twitter:image,omitempty"`
	XCardsAudio    []XCardsAudio    `json:"twitter:audio,omitempty"`
	XCardsVideo    []XCardsVideo    `json:"twitter:video,omitempty"`

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

// XCardsImage represents XCards image object
type XCardsImage struct {
	URL       string `json:"twitter:image"`
	SecureURL string `json:"twitter:image:secure_url,omitempty"`
	Type      string `json:"twitter:image:type,omitempty"`
	Width     int    `json:"twitter:image:width,omitempty"`
	Height    int    `json:"twitter:image:height,omitempty"`
	Alt       string `json:"twitter:image:alt,omitempty"`
}

// XCardsVideo represents XCards video object
type XCardsVideo struct {
	URL       string `json:"twitter:video"`
	SecureURL string `json:"twitter:video:secure_url,omitempty"`
	Type      string `json:"twitter:video:type,omitempty"`
	Width     int    `json:"twitter:video:width,omitempty"`
	Height    int    `json:"twitter:video:height,omitempty"`
}

// XCardsAudio represents XCards audio object
type XCardsAudio struct {
	URL       string `json:"twitter:audio"`
	SecureURL string `json:"twitter:audio:secure_url,omitempty"`
	Type      string `json:"twitter:audio:type,omitempty"`
}

// NewXCards creates a new XCards instance with basic initialization
func NewXCards() *XCards {
	return &XCards{}
}

func ParseXCards(URL string, htmlContent string) (interface{}, []error) {
	_ = URL
	itemXCards, errorsXCards := extractXCards(htmlContent)

	itemOpenGraph, errorsOpenGraph := extractOpenGraph(htmlContent)
	if itemOpenGraph != nil {
		if itemXCards == nil {
			itemXCards = &XCards{}
		}
		errorsFillMissing := fillMissingFieldsFromOpenGraph(itemXCards, itemOpenGraph)
		errorsXCards = append(errorsXCards, errorsFillMissing...)
	}

	var results interface{}
	if itemXCards != nil {
		results = itemXCards
	}

	return results, append(errorsXCards, errorsOpenGraph...)
}

func extractXCards(htmlContent string) (*XCards, []error) {
	var errors []error

	xc := NewXCards()
	tokenizer := html.NewTokenizer(strings.NewReader(htmlContent))

	xcHasValue := false
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
				case "name":
					property = attr.Val
				case "content":
					content = attr.Val
				}
			}
			if property != "" && content != "" {
				parseXCardsMetaTag(xc, property, content)
				xcHasValue = true
			}
		default:
			continue
		}
	}

	if xcHasValue {
		return xc, errors
	}

	return nil, errors
}

func parseXCardsMetaTag(xc *XCards, property, content string) {
	// Split property into parts to handle multi-level properties
	parts := strings.Split(property, ":")

	switch {
	// X specific metadata
	case property == "twitter:card":
		xc.Card = content
	case property == "twitter:site":
		xc.Site = content
	case property == "twitter:creator":
		xc.Creator = content

	// Basic metadata
	case property == "twitter:type":
		xc.Type = content
	case property == "twitter:title":
		xc.Title = content
	case property == "twitter:url":
		xc.URL = content

	// Optional metadata
	case property == "twitter:description":
		xc.Description = content
	case property == "twitter:determiner":
		xc.Determiner = content
	case property == "twitter:locale":
		xc.Locale = content
	case property == "twitter:locale:alternate":
		xc.LocaleAlternate = append(xc.LocaleAlternate, content)
	case property == "twitter:site_name":
		xc.SiteName = content

	// Image handling with multi-level properties
	case strings.HasPrefix(property, "twitter:image"):
		handleXCardsImageProperty(xc, parts, content)

	// Video handling with multi-level properties
	case strings.HasPrefix(property, "twitter:video"):
		handleXCardsVideoProperty(xc, parts, content)

	// Audio handling with multi-level properties
	case strings.HasPrefix(property, "twitter:audio"):
		handleXCardsAudioProperty(xc, parts, content)

	// Music handling with multi-level properties
	case strings.HasPrefix(property, "music:"):
		if xc.Music == nil {
			xc.Music = &Music{}
		}

		switch {
		case property == "music:duration":
			xc.Music.Duration = parseIntSafely(content)
		case property == "music:album":
			xc.Music.Album = content
		case property == "music:album:disc":
			xc.Music.AlbumDisc = parseIntSafely(content)
		case property == "music:album:track":
			xc.Music.AlbumTrack = parseIntSafely(content)
		case property == "music:musician":
			xc.Music.Musician = append(xc.Music.Musician, content)
		case strings.HasPrefix(property, "music:song"):
			handleMusicSongProperty(xc.Music, parts, content)
		case property == "music:creator":
			xc.Music.Creator = append(xc.Music.Creator, content)
		case property == "music:release_date":
			xc.Music.ReleaseDate = content
		}

	// Video handling with multi-level properties
	case strings.HasPrefix(property, "video:"):
		if xc.Video == nil {
			xc.Video = &Video{}
		}

		switch {
		case strings.HasPrefix(property, "video:actor"):
			handleVideoActorProperty(xc.Video, parts, content)
		case property == "video:director":
			xc.Video.Director = append(xc.Video.Director, content)
		case property == "video:writer":
			xc.Video.Writer = append(xc.Video.Writer, content)
		case property == "video:duration":
			xc.Video.Duration = parseIntSafely(content)
		case property == "video:release_date":
			xc.Video.ReleaseDate = parseTimeSafely(content)
		case property == "video:tag":
			xc.Video.Tag = append(xc.Video.Tag, content)
		case property == "video:series":
			xc.Video.Series = content
		}

	// Article handling remains the same
	case strings.HasPrefix(property, "article:"):
		if xc.Article == nil {
			xc.Article = &Article{}
		}
		switch property {
		case "article:published_time":
			xc.Article.PublishedTime = parseTimeSafely(content)
		case "article:modified_time":
			xc.Article.ModifiedTime = parseTimeSafely(content)
		case "article:expiration_time":
			xc.Article.ExpirationTime = parseTimeSafely(content)
		case "article:author":
			xc.Article.Author = append(xc.Article.Author, content)
		case "article:section":
			xc.Article.Section = content
		case "article:tag":
			xc.Article.Tag = append(xc.Article.Tag, content)
		}

	// Book handling remains the same
	case strings.HasPrefix(property, "book:"):
		if xc.Book == nil {
			xc.Book = &Book{}
		}
		switch property {
		case "book:isbn":
			xc.Book.ISBN = content
		case "book:release_date":
			xc.Book.ReleaseDate = parseTimeSafely(content)
		case "book:author":
			xc.Book.Author = append(xc.Book.Author, content)
		case "book:tag":
			xc.Book.Tag = append(xc.Book.Tag, content)
		}

	// Profile handling remains the same
	case strings.HasPrefix(property, "profile:"):
		if xc.Profile == nil {
			xc.Profile = &Profile{}
		}
		switch property {
		case "profile:first_name":
			xc.Profile.FirstName = content
		case "profile:last_name":
			xc.Profile.LastName = content
		case "profile:username":
			xc.Profile.Username = content
		case "profile:gender":
			xc.Profile.Gender = content
		}
	}
}

func handleXCardsImageProperty(xc *XCards, parts []string, content string) {
	if len(xc.XCardsImage) == 0 || parts[1] == "image" {
		if len(parts) < 3 {
			xc.XCardsImage = append(xc.XCardsImage, XCardsImage{})
		}
	}
	lastIdx := len(xc.XCardsImage) - 1

	if len(parts) == 2 {
		xc.XCardsImage[lastIdx].URL = content
		return
	}

	switch parts[2] {
	case "secure_url":
		xc.XCardsImage[lastIdx].SecureURL = content
	case "type":
		xc.XCardsImage[lastIdx].Type = content
	case "width":
		xc.XCardsImage[lastIdx].Width = parseIntSafely(content)
	case "height":
		xc.XCardsImage[lastIdx].Height = parseIntSafely(content)
	case "alt":
		xc.XCardsImage[lastIdx].Alt = content
	}
}

func handleXCardsVideoProperty(xc *XCards, parts []string, content string) {
	if len(xc.XCardsVideo) == 0 || parts[1] == "video" {
		if len(parts) < 3 {
			xc.XCardsVideo = append(xc.XCardsVideo, XCardsVideo{})
		}
	}
	lastIdx := len(xc.XCardsVideo) - 1

	if len(parts) == 2 {
		xc.XCardsVideo[lastIdx].URL = content
		return
	}

	switch parts[2] {
	case "secure_url":
		xc.XCardsVideo[lastIdx].SecureURL = content
	case "type":
		xc.XCardsVideo[lastIdx].Type = content
	case "width":
		xc.XCardsVideo[lastIdx].Width = parseIntSafely(content)
	case "height":
		xc.XCardsVideo[lastIdx].Height = parseIntSafely(content)
	}
}

func handleXCardsAudioProperty(xc *XCards, parts []string, content string) {
	if len(xc.XCardsAudio) == 0 || parts[1] == "audio" {
		if len(parts) < 3 {
			xc.XCardsAudio = append(xc.XCardsAudio, XCardsAudio{})
		}
	}
	lastIdx := len(xc.XCardsAudio) - 1

	if len(parts) == 2 {
		xc.XCardsAudio[lastIdx].URL = content
		return
	}

	switch parts[2] {
	case "secure_url":
		xc.XCardsAudio[lastIdx].SecureURL = content
	case "type":
		xc.XCardsAudio[lastIdx].Type = content
	}
}

// fillMissingFieldsFromOpenGraph fills missing fields in the target struct with values from the source struct.
func fillMissingFieldsFromOpenGraph(target, source interface{}) []error {
	var errors []error

	// Check that both target and source are non-nil pointers to structs
	tVal := reflect.ValueOf(target)
	if tVal.Kind() != reflect.Ptr || tVal.IsNil() {
		errors = append(errors, fmt.Errorf("target must be a non-nil pointer to a struct"))
	}
	tVal = tVal.Elem()

	sVal := reflect.ValueOf(source)
	if sVal.Kind() != reflect.Ptr || sVal.IsNil() {
		errors = append(errors, fmt.Errorf("source must be a non-nil pointer to a struct"))
	}
	sVal = sVal.Elem()

	// Iterate over fields in source, matching by field name
	for i := 0; i < sVal.NumField(); i++ {
		sField := sVal.Field(i)
		sFieldName := sVal.Type().Field(i).Name

		// Check if target has the same field
		tField := tVal.FieldByName(sFieldName)
		if !tField.IsValid() {
			continue // Skip if target does not have this field
		}

		switch tField.Kind() {
		case reflect.String:
			if tField.String() == "" {
				tField.Set(sField)
			}
		case reflect.Ptr:
			if tField.IsNil() && !sField.IsNil() {
				tField.Set(sField)
			} else if !tField.IsNil() && !sField.IsNil() {
				errs := fillMissingFieldsFromOpenGraph(tField.Interface(), sField.Interface())
				errors = append(errors, errs...)
			}
		case reflect.Slice:
			if tField.IsNil() && sField.Len() > 0 {
				tField.Set(sField)
			}
		case reflect.Struct:
			errs := fillMissingFieldsFromOpenGraph(tField.Addr().Interface(), sField.Addr().Interface())
			errors = append(errors, errs...)
		default:
			continue
		}
	}

	return errors
}
