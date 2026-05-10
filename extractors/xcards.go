package extractor

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

func ParseXCards(URL string, htmlContent string) (*XCards, []error) {
	itemXCards, errorsXCards := extractXCards(htmlContent)

	itemOpenGraph, errorsOpenGraph := extractOpenGraph(htmlContent)
	if itemOpenGraph != nil {
		if itemXCards == nil {
			itemXCards = &XCards{}
		}
		errorsFillMissing := fillMissingFieldsFromOpenGraph(itemXCards, itemOpenGraph)
		errorsXCards = append(errorsXCards, errorsFillMissing...)
	}

	if itemXCards != nil && URL != "" {
		resolveXCardsURLs(itemXCards, URL)
	}

	return itemXCards, append(errorsXCards, errorsOpenGraph...)
}

func resolveXCardsURLs(xc *XCards, base string) {
	xc.URL = resolveURL(base, xc.URL)
	for i := range xc.OpenGraphImage {
		xc.OpenGraphImage[i].URL = resolveURL(base, xc.OpenGraphImage[i].URL)
		xc.OpenGraphImage[i].SecureURL = resolveURL(base, xc.OpenGraphImage[i].SecureURL)
	}
	for i := range xc.OpenGraphVideo {
		xc.OpenGraphVideo[i].URL = resolveURL(base, xc.OpenGraphVideo[i].URL)
		xc.OpenGraphVideo[i].SecureURL = resolveURL(base, xc.OpenGraphVideo[i].SecureURL)
	}
	for i := range xc.OpenGraphAudio {
		xc.OpenGraphAudio[i].URL = resolveURL(base, xc.OpenGraphAudio[i].URL)
		xc.OpenGraphAudio[i].SecureURL = resolveURL(base, xc.OpenGraphAudio[i].SecureURL)
	}
	for i := range xc.XCardsImage {
		xc.XCardsImage[i].URL = resolveURL(base, xc.XCardsImage[i].URL)
		xc.XCardsImage[i].SecureURL = resolveURL(base, xc.XCardsImage[i].SecureURL)
	}
	for i := range xc.XCardsVideo {
		xc.XCardsVideo[i].URL = resolveURL(base, xc.XCardsVideo[i].URL)
		xc.XCardsVideo[i].SecureURL = resolveURL(base, xc.XCardsVideo[i].SecureURL)
	}
	for i := range xc.XCardsAudio {
		xc.XCardsAudio[i].URL = resolveURL(base, xc.XCardsAudio[i].URL)
		xc.XCardsAudio[i].SecureURL = resolveURL(base, xc.XCardsAudio[i].SecureURL)
	}
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
			if property != "" && content != "" && isXCardsProperty(property) {
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
		parseMusicProperty(&xc.Music, parts, property, content)

	// Video object handling with multi-level properties
	case strings.HasPrefix(property, "video:"):
		parseVideoObjectProperty(&xc.Video, parts, property, content)

	// Article handling
	case strings.HasPrefix(property, "article:"):
		parseArticleProperty(&xc.Article, property, content)

	// Book handling
	case strings.HasPrefix(property, "book:"):
		parseBookProperty(&xc.Book, property, content)

	// Profile handling
	case strings.HasPrefix(property, "profile:"):
		parseProfileProperty(&xc.Profile, property, content)
	}
}

func handleXCardsImageProperty(xc *XCards, parts []string, content string) {
	handleMediaSlice(&xc.XCardsImage, "image", parts, content, func(img *XCardsImage, sub, val string) {
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

func handleXCardsVideoProperty(xc *XCards, parts []string, content string) {
	handleMediaSlice(&xc.XCardsVideo, "video", parts, content, func(v *XCardsVideo, sub, val string) {
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

func handleXCardsAudioProperty(xc *XCards, parts []string, content string) {
	handleMediaSlice(&xc.XCardsAudio, "audio", parts, content, func(a *XCardsAudio, sub, val string) {
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

// fillMissingFieldsFromOpenGraph fills missing fields in the target struct with values from the source struct.
func fillMissingFieldsFromOpenGraph(target, source any) []error {
	var errors []error

	// Check that both target and source are non-nil pointers to structs
	tVal := reflect.ValueOf(target)
	if tVal.Kind() != reflect.Ptr || tVal.IsNil() {
		errors = append(errors, fmt.Errorf("target must be a non-nil pointer to a struct"))
	}

	sVal := reflect.ValueOf(source)
	if sVal.Kind() != reflect.Ptr || sVal.IsNil() {
		errors = append(errors, fmt.Errorf("source must be a non-nil pointer to a struct"))
	}

	if len(errors) > 0 {
		return errors
	}

	tVal = tVal.Elem()
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

func isXCardsProperty(property string) bool {
	return strings.HasPrefix(property, "twitter:") ||
		strings.HasPrefix(property, "og:") ||
		strings.HasPrefix(property, "music:") ||
		strings.HasPrefix(property, "video:") ||
		strings.HasPrefix(property, "article:") ||
		strings.HasPrefix(property, "book:") ||
		strings.HasPrefix(property, "profile:")
}
