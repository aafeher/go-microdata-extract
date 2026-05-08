package extractor

import (
	"fmt"
	"strings"
	"time"
)

// handleMediaSlice appends a new zero item when starting a fresh base property or when
// the slice is empty, then delegates field assignment to apply.
func handleMediaSlice[T any](items *[]T, mediaName string, parts []string, content string, apply func(*T, string, string)) {
	isBase := parts[1] == mediaName
	hasSub := len(parts) >= 3
	if (isBase && !hasSub) || len(*items) == 0 {
		var zero T
		*items = append(*items, zero)
	}
	sub := ""
	if hasSub {
		sub = parts[2]
	}
	apply(&(*items)[len(*items)-1], sub, content)
}

func parseMusicProperty(music **Music, parts []string, property, content string) {
	if *music == nil {
		*music = &Music{}
	}
	switch {
	case property == "music:duration":
		(*music).Duration = parseIntSafely(content)
	case property == "music:album":
		(*music).Album = content
	case property == "music:album:disc":
		(*music).AlbumDisc = parseIntSafely(content)
	case property == "music:album:track":
		(*music).AlbumTrack = parseIntSafely(content)
	case property == "music:musician":
		(*music).Musician = append((*music).Musician, content)
	case strings.HasPrefix(property, "music:song"):
		handleMusicSongProperty(*music, parts, content)
	case property == "music:release_date":
		(*music).ReleaseDate = content
	case property == "music:creator":
		(*music).Creator = append((*music).Creator, content)
	}
}

func parseVideoObjectProperty(video **Video, parts []string, property, content string) {
	if *video == nil {
		*video = &Video{}
	}
	switch {
	case strings.HasPrefix(property, "video:actor"):
		handleVideoActorProperty(*video, parts, content)
	case property == "video:director":
		(*video).Director = append((*video).Director, content)
	case property == "video:writer":
		(*video).Writer = append((*video).Writer, content)
	case property == "video:duration":
		(*video).Duration = parseIntSafely(content)
	case property == "video:release_date":
		(*video).ReleaseDate = parseTimeSafely(content)
	case property == "video:tag":
		(*video).Tag = append((*video).Tag, content)
	case property == "video:series":
		(*video).Series = content
	}
}

func parseArticleProperty(article **Article, property, content string) {
	if *article == nil {
		*article = &Article{}
	}
	switch property {
	case "article:published_time":
		(*article).PublishedTime = parseTimeSafely(content)
	case "article:modified_time":
		(*article).ModifiedTime = parseTimeSafely(content)
	case "article:expiration_time":
		(*article).ExpirationTime = parseTimeSafely(content)
	case "article:author":
		(*article).Author = append((*article).Author, content)
	case "article:section":
		(*article).Section = content
	case "article:tag":
		(*article).Tag = append((*article).Tag, content)
	}
}

func parseBookProperty(book **Book, property, content string) {
	if *book == nil {
		*book = &Book{}
	}
	switch property {
	case "book:isbn":
		(*book).ISBN = content
	case "book:release_date":
		(*book).ReleaseDate = parseTimeSafely(content)
	case "book:author":
		(*book).Author = append((*book).Author, content)
	case "book:tag":
		(*book).Tag = append((*book).Tag, content)
	}
}

func parseProfileProperty(profile **Profile, property, content string) {
	if *profile == nil {
		*profile = &Profile{}
	}
	switch property {
	case "profile:first_name":
		(*profile).FirstName = content
	case "profile:last_name":
		(*profile).LastName = content
	case "profile:username":
		(*profile).Username = content
	case "profile:gender":
		(*profile).Gender = content
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
