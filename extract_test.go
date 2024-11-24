package extract

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	extract "github.com/aafeher/go-microdata-extract/extractors"
	"reflect"
	"testing"
	"time"
)

func TestExtractor_setConfigDefaults(t *testing.T) {
	tests := []struct {
		name string
		e    *Extractor
		want config
	}{
		{
			name: "default config",
			e:    &Extractor{},
			want: config{
				syntaxes:     SYNTAXES,
				userAgent:    "go-microdata-extract (+https://github.com/aafeher/go-microdata-extract/blob/main/README.md)",
				fetchTimeout: 3,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.e.setConfigDefaults()

			if !areSyntaxSlicesEqual(test.e.cfg.syntaxes, test.want.syntaxes) || test.e.cfg.userAgent != test.want.userAgent || test.e.cfg.fetchTimeout != test.want.fetchTimeout {
				t.Errorf("expected %v, got %v", test.want, test.e.cfg)
			}
		})
	}
}

func TestExtractor_SetSyntaxes(t *testing.T) {
	tests := []struct {
		name     string
		syntaxes []Syntax
		want     []Syntax
	}{
		{
			name:     "Empty syntax list",
			syntaxes: []Syntax{},
			want:     SYNTAXES,
		},
		{
			name:     "invalid syntax list",
			syntaxes: []Syntax{"a"},
			want:     SYNTAXES,
		},
		{
			name:     "mixed syntax list",
			syntaxes: []Syntax{"a", SyntaxOpenGraph},
			want:     []Syntax{SyntaxOpenGraph},
		},
		{
			name:     "valid syntax list",
			syntaxes: []Syntax{SyntaxOpenGraph},
			want:     []Syntax{SyntaxOpenGraph},
		},
		{
			name:     "full syntax list",
			syntaxes: SYNTAXES,
			want:     SYNTAXES,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			e := New()
			e.SetSyntaxes(test.syntaxes)
			if !areSyntaxSlicesEqual(e.cfg.syntaxes, test.want) {
				t.Errorf("expected %q, got %q", test.want, e.cfg.userAgent)
			}
		})
	}
}

func TestExtractor_SetUserAgent(t *testing.T) {
	tests := []struct {
		name      string
		userAgent string
		want      string
	}{
		{
			name:      "Empty User Agent",
			userAgent: "",
			want:      "",
		},
		{
			name:      "Normal User Agent",
			userAgent: "Mozilla/5.0 Firefox/61.0",
			want:      "Mozilla/5.0 Firefox/61.0",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			e := New()
			e.SetUserAgent(test.userAgent)
			if e.cfg.userAgent != test.want {
				t.Errorf("expected %q, got %q", test.want, e.cfg.userAgent)
			}
		})
	}
}

func TestExtractor_SetFetchTimeout(t *testing.T) {
	tests := []struct {
		name    string
		timeout uint8
	}{
		{
			name:    "PositiveTimeout",
			timeout: 5,
		},
		{
			name:    "ZeroTimeout",
			timeout: 0,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			e := New()
			e.SetFetchTimeout(test.timeout)
			if e.cfg.fetchTimeout != test.timeout {
				t.Errorf("expected %v, got %v", test.timeout, e.cfg.fetchTimeout)
			}
		})
	}
}

func TestExtractor_Extract(t *testing.T) {
	server := testServer()
	defer server.Close()

	tests := []struct {
		name      string
		url       string
		content   *string
		err       *string
		extracted map[Syntax]any
		errs      []error
	}{
		{
			name:      "testServer index page",
			url:       server.URL,
			content:   nil,
			err:       pointerOfString("received HTTP status 404"),
			extracted: map[Syntax]any{},
			errs:      []error{errors.New("received HTTP status 404")},
		},
		{
			name:    "page with no structured data",
			url:     server.URL,
			content: pointerOfString("<html>error</p></html>"),
			err:     nil,
			extracted: map[Syntax]any{
				"opengraph": nil,
				"xcards":    nil,
				"json-ld":   []map[string]any(nil),
				"microdata": []extract.MicrodataItem(nil),
			},
			errs: nil,
		},
		{
			name:    "test-01-opengraph-minimal",
			url:     fmt.Sprintf("%s/test-01-opengraph-minimal.html", server.URL),
			content: nil,
			err:     nil,
			extracted: map[Syntax]any{
				"opengraph": &extract.OpenGraph{
					Type:  "website",
					Title: "go-microdata-extract",
					URL:   "https://github.com/aafeher/go-microdata-extract",
				},
				"xcards": &extract.XCards{
					Type:  "website",
					Title: "go-microdata-extract",
					URL:   "https://github.com/aafeher/go-microdata-extract",
				},
				"json-ld":   []map[string]any(nil),
				"microdata": []extract.MicrodataItem(nil),
			},
			errs: nil,
		},
		{
			name:    "test-02-opengraph-optional",
			url:     fmt.Sprintf("%s/test-02-opengraph-optional.html", server.URL),
			content: nil,
			err:     nil,
			extracted: map[Syntax]any{
				"opengraph": &extract.OpenGraph{
					Type:        `website`,
					Title:       `go-microdata-extract`,
					URL:         `https://github.com/aafeher/go-microdata-extract`,
					Description: `OpenGraph with optional metadata`,
					Determiner:  "the",
					OpenGraphImage: []extract.OpenGraphImage{
						{
							URL: "https://picsum.photos/200/300",
						},
						{
							URL: "https://picsum.photos/210/310",
						},
					},
					Locale: "en_GB",
					LocaleAlternate: []string{
						"hu_HU",
						"fr_FR",
					},
					SiteName: "go-microdata-extract",
					OpenGraphAudio: []extract.OpenGraphAudio{
						{
							URL: "https://example.com/bond/theme.mp3",
						},
					},
					OpenGraphVideo: []extract.OpenGraphVideo{
						{
							URL: "https://example.com/bond/trailer.swf",
						},
					},
				},
				"xcards": &extract.XCards{
					Type:        `website`,
					Title:       `go-microdata-extract`,
					URL:         `https://github.com/aafeher/go-microdata-extract`,
					Description: `OpenGraph with optional metadata`,
					Determiner:  "the",
					OpenGraphImage: []extract.OpenGraphImage{
						{
							URL: "https://picsum.photos/200/300",
						},
						{
							URL: "https://picsum.photos/210/310",
						},
					},
					Locale: "en_GB",
					LocaleAlternate: []string{
						"hu_HU",
						"fr_FR",
					},
					SiteName: "go-microdata-extract",
					OpenGraphAudio: []extract.OpenGraphAudio{
						{
							URL: "https://example.com/bond/theme.mp3",
						},
					},
					OpenGraphVideo: []extract.OpenGraphVideo{
						{
							URL: "https://example.com/bond/trailer.swf",
						},
					},
				},
				"json-ld":   []map[string]any(nil),
				"microdata": []extract.MicrodataItem(nil),
			},
			errs: nil,
		},
		{
			name:    "test-03-opengraph-image",
			url:     fmt.Sprintf("%s/test-03-opengraph-image.html", server.URL),
			content: nil,
			err:     nil,
			extracted: map[Syntax]any{
				"opengraph": &extract.OpenGraph{
					Type:        `website`,
					Title:       `go-microdata-extract`,
					URL:         `https://github.com/aafeher/go-microdata-extract`,
					Description: `OpenGraph with image`,
					OpenGraphImage: []extract.OpenGraphImage{
						{
							URL: "https://picsum.photos/200/300",
						},
						{
							URL:       "https://picsum.photos/210/310",
							SecureURL: "https://picsum.photos/210/310",
							Type:      "image/jpeg",
							Width:     210,
							Height:    310,
							Alt:       "image for testing",
						},
					},
				},
				"xcards": &extract.XCards{
					Type:        `website`,
					Title:       `go-microdata-extract`,
					URL:         `https://github.com/aafeher/go-microdata-extract`,
					Description: `OpenGraph with image`,
					OpenGraphImage: []extract.OpenGraphImage{
						{
							URL: "https://picsum.photos/200/300",
						},
						{
							URL:       "https://picsum.photos/210/310",
							SecureURL: "https://picsum.photos/210/310",
							Type:      "image/jpeg",
							Width:     210,
							Height:    310,
							Alt:       "image for testing",
						},
					},
				},
				"json-ld":   []map[string]any(nil),
				"microdata": []extract.MicrodataItem(nil),
			},
			errs: nil,
		},
		{
			name:    "test-04-opengraph-video",
			url:     fmt.Sprintf("%s/test-04-opengraph-video.html", server.URL),
			content: nil,
			err:     nil,
			extracted: map[Syntax]any{
				"opengraph": &extract.OpenGraph{
					Type:        `website`,
					Title:       `go-microdata-extract`,
					URL:         `https://github.com/aafeher/go-microdata-extract`,
					Description: `OpenGraph with video`,
					OpenGraphVideo: []extract.OpenGraphVideo{
						{
							URL: "https://example.com/movie.mp4",
						},
						{
							URL:       "https://example.com/movie2.mp4",
							SecureURL: "https://secure.example.com/movie2.mp4",
							Type:      "video/mp4",
							Width:     400,
							Height:    300,
						},
					},
				},
				"xcards": &extract.XCards{
					Type:        `website`,
					Title:       `go-microdata-extract`,
					URL:         `https://github.com/aafeher/go-microdata-extract`,
					Description: `OpenGraph with video`,
					OpenGraphVideo: []extract.OpenGraphVideo{
						{
							URL: "https://example.com/movie.mp4",
						},
						{
							URL:       "https://example.com/movie2.mp4",
							SecureURL: "https://secure.example.com/movie2.mp4",
							Type:      "video/mp4",
							Width:     400,
							Height:    300,
						},
					},
				},
				"json-ld":   []map[string]any(nil),
				"microdata": []extract.MicrodataItem(nil),
			},
			errs: nil,
		},
		{
			name:    "test-05-opengraph-audio",
			url:     fmt.Sprintf("%s/test-05-opengraph-audio.html", server.URL),
			content: nil,
			err:     nil,
			extracted: map[Syntax]any{
				"opengraph": &extract.OpenGraph{
					Type:        `website`,
					Title:       `go-microdata-extract`,
					URL:         `https://github.com/aafeher/go-microdata-extract`,
					Description: `OpenGraph with audio`,
					OpenGraphAudio: []extract.OpenGraphAudio{
						{
							URL: "https://example.com/sound.mp3",
						},
						{
							URL:       "https://example.com/sound2.mp3",
							SecureURL: "https://secure.example.com/sound2.mp3",
							Type:      "audio/mpeg",
						},
					},
				},
				"xcards": &extract.XCards{
					Type:        `website`,
					Title:       `go-microdata-extract`,
					URL:         `https://github.com/aafeher/go-microdata-extract`,
					Description: `OpenGraph with audio`,
					OpenGraphAudio: []extract.OpenGraphAudio{
						{
							URL: "https://example.com/sound.mp3",
						},
						{
							URL:       "https://example.com/sound2.mp3",
							SecureURL: "https://secure.example.com/sound2.mp3",
							Type:      "audio/mpeg",
						},
					},
				},
				"json-ld":   []map[string]any(nil),
				"microdata": []extract.MicrodataItem(nil),
			},
			errs: nil,
		},
		{
			name:    "test-06-opengraph-music-song",
			url:     fmt.Sprintf("%s/test-06-opengraph-music-song.html", server.URL),
			content: nil,
			err:     nil,
			extracted: map[Syntax]any{
				"opengraph": &extract.OpenGraph{
					Type:     `music.song`,
					Title:    `Under Pressure`,
					URL:      `http://open.spotify.com/track/2aSFLiDPreOVP6KHiWk4lF`,
					SiteName: "Spotify",
					OpenGraphImage: []extract.OpenGraphImage{
						{
							URL: "http://o.scdn.co/image/e4c7b06c20c17156e46bbe9a71eb0703281cf345",
						},
					},
					Music: &extract.Music{
						Album:      "http://open.spotify.com/album/7rq68qYz66mNdPfidhIEFa",
						AlbumDisc:  1,
						AlbumTrack: 2,
						Duration:   236,
						Musician: []string{
							"http://open.spotify.com/artist/1dfeR4HaWDbWqFHLkxsg1d",
							"http://open.spotify.com/artist/0oSGxfWSnnOXhD2fKuz2Gy",
						},
					},
				},
				"xcards": &extract.XCards{
					Type:     `music.song`,
					Title:    `Under Pressure`,
					URL:      `http://open.spotify.com/track/2aSFLiDPreOVP6KHiWk4lF`,
					SiteName: "Spotify",
					OpenGraphImage: []extract.OpenGraphImage{
						{
							URL: "http://o.scdn.co/image/e4c7b06c20c17156e46bbe9a71eb0703281cf345",
						},
					},
					Music: &extract.Music{
						Album:      "http://open.spotify.com/album/7rq68qYz66mNdPfidhIEFa",
						AlbumDisc:  1,
						AlbumTrack: 2,
						Duration:   236,
						Musician: []string{
							"http://open.spotify.com/artist/1dfeR4HaWDbWqFHLkxsg1d",
							"http://open.spotify.com/artist/0oSGxfWSnnOXhD2fKuz2Gy",
						},
					},
				},
				"json-ld":   []map[string]any(nil),
				"microdata": []extract.MicrodataItem(nil),
			},
			errs: nil,
		},
		{
			name:    "test-07-opengraph-music-album",
			url:     fmt.Sprintf("%s/test-07-opengraph-music-album.html", server.URL),
			content: nil,
			err:     nil,
			extracted: map[Syntax]any{
				"opengraph": &extract.OpenGraph{
					Type:        `music.album`,
					Title:       `Greatest Hits II`,
					URL:         `http://open.spotify.com/album/7rq68qYz66mNdPfidhIEFa`,
					Description: `Greatest Hits II, an album by Queen on Spotify.`,
					SiteName:    "Spotify",
					OpenGraphImage: []extract.OpenGraphImage{
						{
							URL: "http://o.scdn.co/image/e4c7b06c20c17156e46bbe9a71eb0703281cf345",
						},
					},
					Music: &extract.Music{
						Musician: []string{
							"http://open.spotify.com/artist/1dfeR4HaWDbWqFHLkxsg1d",
						},
						Song: []extract.MusicSong{
							{
								URL:   "http://open.spotify.com/track/0pfHfdUNVwlXA0WDXznm2C",
								Disc:  1,
								Track: 1,
							},
							{
								URL:   "http://open.spotify.com/track/2aSFLiDPreOVP6KHiWk4lF",
								Disc:  1,
								Track: 2,
							},
						},
						ReleaseDate: "2011-04-19",
					},
				},
				"xcards": &extract.XCards{
					Type:        `music.album`,
					Title:       `Greatest Hits II`,
					URL:         `http://open.spotify.com/album/7rq68qYz66mNdPfidhIEFa`,
					Description: `Greatest Hits II, an album by Queen on Spotify.`,
					SiteName:    "Spotify",
					OpenGraphImage: []extract.OpenGraphImage{
						{
							URL: "http://o.scdn.co/image/e4c7b06c20c17156e46bbe9a71eb0703281cf345",
						},
					},
					Music: &extract.Music{
						Musician: []string{
							"http://open.spotify.com/artist/1dfeR4HaWDbWqFHLkxsg1d",
						},
						Song: []extract.MusicSong{
							{
								URL:   "http://open.spotify.com/track/0pfHfdUNVwlXA0WDXznm2C",
								Disc:  1,
								Track: 1,
							},
							{
								URL:   "http://open.spotify.com/track/2aSFLiDPreOVP6KHiWk4lF",
								Disc:  1,
								Track: 2,
							},
						},
						ReleaseDate: "2011-04-19",
					},
				},
				"json-ld":   []map[string]any(nil),
				"microdata": []extract.MicrodataItem(nil),
			},
			errs: nil,
		},
		{
			name:    "test-08-opengraph-music-playlist",
			url:     fmt.Sprintf("%s/test-08-opengraph-music-playlist.html", server.URL),
			content: nil,
			err:     nil,
			extracted: map[Syntax]any{
				"opengraph": &extract.OpenGraph{
					Type:     `music.playlist`,
					Title:    `on repeat`,
					URL:      `http://open.spotify.com/user/austinhaugen/playlist/1a8444uyNXVOpwtFdgakhv`,
					SiteName: "Spotify",
					OpenGraphImage: []extract.OpenGraphImage{
						{
							URL: "http://o.scdn.co/300/756df3afcb3d14cb362448b68ed2f5506479f313",
						},
					},
					Music: &extract.Music{
						Creator: []string{
							"http://open.spotify.com/user/austinhaugen",
						},
					},
				},
				"xcards": &extract.XCards{
					Type:     `music.playlist`,
					Title:    `on repeat`,
					URL:      `http://open.spotify.com/user/austinhaugen/playlist/1a8444uyNXVOpwtFdgakhv`,
					SiteName: "Spotify",
					OpenGraphImage: []extract.OpenGraphImage{
						{
							URL: "http://o.scdn.co/300/756df3afcb3d14cb362448b68ed2f5506479f313",
						},
					},
					Music: &extract.Music{
						Creator: []string{
							"http://open.spotify.com/user/austinhaugen",
						},
					},
				},
				"json-ld":   []map[string]any(nil),
				"microdata": []extract.MicrodataItem(nil),
			},
			errs: nil,
		},
		{
			name:    "test-09-opengraph-video-movie",
			url:     fmt.Sprintf("%s/test-09-opengraph-video-movie.html", server.URL),
			content: nil,
			err:     nil,
			extracted: map[Syntax]any{
				"opengraph": &extract.OpenGraph{
					Type:     `video.movie`,
					Title:    `OpenGraph Video Movie Title`,
					URL:      `https://www.example.com/videos/video-movie-title`,
					SiteName: "SiteName",
					Video: &extract.Video{
						Actor: []extract.VideoActor{
							{
								URL:  "https://www.example.com/actors/@firstnameA-lastnameA",
								Role: "ant",
							},
							{
								URL:  "https://www.example.com/actors/@firstnameB-lastnameB",
								Role: "bear",
							},
						},
						Director: []string{
							"https://www.example.com/actors/@firstnameA-lastnameA",
							"https://www.example.com/actors/@firstnameB-lastnameB",
						},
						Writer: []string{
							"https://www.example.com/actors/@firstnameA-lastnameA",
							"https://www.example.com/actors/@firstnameB-lastnameB",
						},
						Duration:    42,
						ReleaseDate: time.Date(2024, 10, 31, 0, 0, 0, 0, time.UTC),
						Tag: []string{
							"tag A",
							"tag B",
						},
					},
				},
				"xcards": &extract.XCards{
					Type:     `video.movie`,
					Title:    `OpenGraph Video Movie Title`,
					URL:      `https://www.example.com/videos/video-movie-title`,
					SiteName: "SiteName",
					Video: &extract.Video{
						Actor: []extract.VideoActor{
							{
								URL:  "https://www.example.com/actors/@firstnameA-lastnameA",
								Role: "ant",
							},
							{
								URL:  "https://www.example.com/actors/@firstnameB-lastnameB",
								Role: "bear",
							},
						},
						Director: []string{
							"https://www.example.com/actors/@firstnameA-lastnameA",
							"https://www.example.com/actors/@firstnameB-lastnameB",
						},
						Writer: []string{
							"https://www.example.com/actors/@firstnameA-lastnameA",
							"https://www.example.com/actors/@firstnameB-lastnameB",
						},
						Duration:    42,
						ReleaseDate: time.Date(2024, 10, 31, 0, 0, 0, 0, time.UTC),
						Tag: []string{
							"tag A",
							"tag B",
						},
					},
				},
				"json-ld":   []map[string]any(nil),
				"microdata": []extract.MicrodataItem(nil),
			},
			errs: nil,
		},
		{
			name:    "test-10-opengraph-video-episode",
			url:     fmt.Sprintf("%s/test-10-opengraph-video-episode.html", server.URL),
			content: nil,
			err:     nil,
			extracted: map[Syntax]any{
				"opengraph": &extract.OpenGraph{
					Type:     `video.episode`,
					Title:    `OpenGraph Video Episode Title`,
					URL:      `https://www.example.com/videos/video-episode-title`,
					SiteName: "SiteName",
					Video: &extract.Video{
						Actor: []extract.VideoActor{
							{
								URL:  "https://www.example.com/actors/@firstnameA-lastnameA",
								Role: "ant",
							},
							{
								URL:  "https://www.example.com/actors/@firstnameB-lastnameB",
								Role: "bear",
							},
						},
						Director: []string{
							"https://www.example.com/actors/@firstnameA-lastnameA",
							"https://www.example.com/actors/@firstnameB-lastnameB",
						},
						Writer: []string{
							"https://www.example.com/actors/@firstnameA-lastnameA",
							"https://www.example.com/actors/@firstnameB-lastnameB",
						},
						Duration:    42,
						ReleaseDate: time.Date(2024, 10, 31, 0, 0, 0, 0, time.UTC),
						Tag: []string{
							"tag A",
							"tag B",
						},
						Series: "Video Series",
					},
				},
				"xcards": &extract.XCards{
					Type:     `video.episode`,
					Title:    `OpenGraph Video Episode Title`,
					URL:      `https://www.example.com/videos/video-episode-title`,
					SiteName: "SiteName",
					Video: &extract.Video{
						Actor: []extract.VideoActor{
							{
								URL:  "https://www.example.com/actors/@firstnameA-lastnameA",
								Role: "ant",
							},
							{
								URL:  "https://www.example.com/actors/@firstnameB-lastnameB",
								Role: "bear",
							},
						},
						Director: []string{
							"https://www.example.com/actors/@firstnameA-lastnameA",
							"https://www.example.com/actors/@firstnameB-lastnameB",
						},
						Writer: []string{
							"https://www.example.com/actors/@firstnameA-lastnameA",
							"https://www.example.com/actors/@firstnameB-lastnameB",
						},
						Duration:    42,
						ReleaseDate: time.Date(2024, 10, 31, 0, 0, 0, 0, time.UTC),
						Tag: []string{
							"tag A",
							"tag B",
						},
						Series: "Video Series",
					},
				},
				"json-ld":   []map[string]any(nil),
				"microdata": []extract.MicrodataItem(nil),
			},
			errs: nil,
		},
		{
			name:    "test-11-opengraph-article",
			url:     fmt.Sprintf("%s/test-11-opengraph-article.html", server.URL),
			content: nil,
			err:     nil,
			extracted: map[Syntax]any{
				"opengraph": &extract.OpenGraph{
					Type:     `article`,
					Title:    `OpenGraph Article Title`,
					URL:      `https://www.example.com/article/article-title`,
					SiteName: "SiteName",
					Article: &extract.Article{
						PublishedTime:  time.Date(2024, 10, 01, 0, 0, 0, 0, time.UTC),
						ModifiedTime:   time.Date(2024, 10, 31, 0, 0, 0, 0, time.UTC),
						ExpirationTime: time.Date(2024, 11, 01, 0, 0, 0, 0, time.UTC),
						Author: []string{
							"https://www.example.com/profileAuthorA.html",
							"https://www.example.com/profileAuthorB.html",
						},
						Section: "Front page",
						Tag: []string{
							"tag A",
							"tag B",
						},
					},
				},
				"xcards": &extract.XCards{
					Type:     `article`,
					Title:    `OpenGraph Article Title`,
					URL:      `https://www.example.com/article/article-title`,
					SiteName: "SiteName",
					Article: &extract.Article{
						PublishedTime:  time.Date(2024, 10, 01, 0, 0, 0, 0, time.UTC),
						ModifiedTime:   time.Date(2024, 10, 31, 0, 0, 0, 0, time.UTC),
						ExpirationTime: time.Date(2024, 11, 01, 0, 0, 0, 0, time.UTC),
						Author: []string{
							"https://www.example.com/profileAuthorA.html",
							"https://www.example.com/profileAuthorB.html",
						},
						Section: "Front page",
						Tag: []string{
							"tag A",
							"tag B",
						},
					},
				},
				"json-ld":   []map[string]any(nil),
				"microdata": []extract.MicrodataItem(nil),
			},
			errs: nil,
		},
		{
			name:    "test-12-opengraph-book",
			url:     fmt.Sprintf("%s/test-12-opengraph-book.html", server.URL),
			content: nil,
			err:     nil,
			extracted: map[Syntax]any{
				"opengraph": &extract.OpenGraph{
					Type:     `book`,
					Title:    `OpenGraph Book Title`,
					URL:      `https://www.example.com/book/book-title`,
					SiteName: "SiteName",
					Book: &extract.Book{
						Author: []string{
							"https://www.example.com/profileAuthorA.html",
							"https://www.example.com/profileAuthorB.html",
						},
						ReleaseDate: time.Date(2024, 10, 31, 0, 0, 0, 0, time.UTC),
						ISBN:        "9871234567890",
						Tag: []string{
							"tag A",
							"tag B",
						},
					},
				},
				"xcards": &extract.XCards{
					Type:     `book`,
					Title:    `OpenGraph Book Title`,
					URL:      `https://www.example.com/book/book-title`,
					SiteName: "SiteName",
					Book: &extract.Book{
						Author: []string{
							"https://www.example.com/profileAuthorA.html",
							"https://www.example.com/profileAuthorB.html",
						},
						ReleaseDate: time.Date(2024, 10, 31, 0, 0, 0, 0, time.UTC),
						ISBN:        "9871234567890",
						Tag: []string{
							"tag A",
							"tag B",
						},
					},
				},
				"json-ld":   []map[string]any(nil),
				"microdata": []extract.MicrodataItem(nil),
			},
			errs: nil,
		},
		{
			name:    "test-13-opengraph-profile",
			url:     fmt.Sprintf("%s/test-13-opengraph-profile.html", server.URL),
			content: nil,
			err:     nil,
			extracted: map[Syntax]any{
				"opengraph": &extract.OpenGraph{
					Type:     `profile`,
					Title:    `OpenGraph Profile Title`,
					URL:      `https://www.example.com/profiles/profile-title`,
					SiteName: "SiteName",
					Profile: &extract.Profile{
						FirstName: "John",
						LastName:  "Doe",
						Username:  "johndoe",
						Gender:    "male",
					},
				},
				"xcards": &extract.XCards{
					Type:     `profile`,
					Title:    `OpenGraph Profile Title`,
					URL:      `https://www.example.com/profiles/profile-title`,
					SiteName: "SiteName",
					Profile: &extract.Profile{
						FirstName: "John",
						LastName:  "Doe",
						Username:  "johndoe",
						Gender:    "male",
					},
				},
				"json-ld":   []map[string]any(nil),
				"microdata": []extract.MicrodataItem(nil),
			},
			errs: nil,
		},
		{
			name:    "test-14-opengraph-errors",
			url:     fmt.Sprintf("%s/test-14-opengraph-errors.html", server.URL),
			content: nil,
			err:     nil,
			extracted: map[Syntax]any{
				"opengraph": &extract.OpenGraph{
					Type:     `video.movie`,
					Title:    `OpenGraph Errors Title`,
					URL:      `https://www.example.com/videos/video-movie-title`,
					SiteName: "SiteName",
					Video: &extract.Video{
						Duration:    0,
						ReleaseDate: time.Time{},
					},
				},
				"xcards": &extract.XCards{
					Type:     `video.movie`,
					Title:    `OpenGraph Errors Title`,
					URL:      `https://www.example.com/videos/video-movie-title`,
					SiteName: "SiteName",
					Video: &extract.Video{
						Duration:    0,
						ReleaseDate: time.Time{},
					},
				},
				"json-ld":   []map[string]any(nil),
				"microdata": []extract.MicrodataItem(nil),
			},
			errs: nil,
		},
		{
			name:    "test-15-xcards-minimal",
			url:     fmt.Sprintf("%s/test-15-xcards-minimal.html", server.URL),
			content: nil,
			err:     nil,
			extracted: map[Syntax]any{
				"opengraph": nil,
				"xcards": &extract.XCards{
					Card:    "summary",
					Site:    "@examplesite",
					Creator: "@creator",
					Type:    `website`,
					Title:   `go-microdata-extract`,
					URL:     `https://github.com/aafeher/go-microdata-extract`,
				},
				"json-ld":   []map[string]any(nil),
				"microdata": []extract.MicrodataItem(nil),
			},
			errs: nil,
		},
		{
			name:    "test-16-xcards-optional",
			url:     fmt.Sprintf("%s/test-16-xcards-optional.html", server.URL),
			content: nil,
			err:     nil,
			extracted: map[Syntax]any{
				"opengraph": nil,
				"xcards": &extract.XCards{
					Card:        "summary",
					Site:        "@examplesite",
					Creator:     "@creator",
					Type:        `website`,
					Title:       `go-microdata-extract`,
					URL:         `https://github.com/aafeher/go-microdata-extract`,
					Description: `X Cards with optional metadata`,
					Determiner:  "the",
					XCardsImage: []extract.XCardsImage{
						{
							URL: "https://picsum.photos/200/300",
						},
						{
							URL: "https://picsum.photos/210/310",
						},
					},
					Locale: "en_GB",
					LocaleAlternate: []string{
						"hu_HU",
						"fr_FR",
					},
					SiteName: "go-microdata-extract",
					XCardsAudio: []extract.XCardsAudio{
						{
							URL: "https://example.com/bond/theme.mp3",
						},
					},
					XCardsVideo: []extract.XCardsVideo{
						{
							URL: "https://example.com/bond/trailer.swf",
						},
					},
				},
				"json-ld":   []map[string]any(nil),
				"microdata": []extract.MicrodataItem(nil),
			},
			errs: nil,
		},
		{
			name:    "test-17-xcards-image",
			url:     fmt.Sprintf("%s/test-17-xcards-image.html", server.URL),
			content: nil,
			err:     nil,
			extracted: map[Syntax]any{
				"opengraph": nil,
				"xcards": &extract.XCards{
					Type:        `website`,
					Title:       `go-microdata-extract`,
					URL:         `https://github.com/aafeher/go-microdata-extract`,
					Description: `X Cards with image`,
					XCardsImage: []extract.XCardsImage{
						{
							URL: "https://picsum.photos/200/300",
						},
						{
							URL:       "https://picsum.photos/210/310",
							SecureURL: "https://picsum.photos/210/310",
							Type:      "image/jpeg",
							Width:     210,
							Height:    310,
							Alt:       "image for testing",
						},
					},
				},
				"json-ld":   []map[string]any(nil),
				"microdata": []extract.MicrodataItem(nil),
			},
			errs: nil,
		},
		{
			name:    "test-18-xcards-video",
			url:     fmt.Sprintf("%s/test-18-xcards-video.html", server.URL),
			content: nil,
			err:     nil,
			extracted: map[Syntax]any{
				"opengraph": nil,
				"xcards": &extract.XCards{
					Type:        `website`,
					Title:       `go-microdata-extract`,
					URL:         `https://github.com/aafeher/go-microdata-extract`,
					Description: `X Cards with video`,
					XCardsVideo: []extract.XCardsVideo{
						{
							URL: "https://example.com/movie.mp4",
						},
						{
							URL:       "https://example.com/movie2.mp4",
							SecureURL: "https://secure.example.com/movie2.mp4",
							Type:      "video/mp4",
							Width:     400,
							Height:    300,
						},
					},
				},
				"json-ld":   []map[string]any(nil),
				"microdata": []extract.MicrodataItem(nil),
			},
			errs: nil,
		},
		{
			name:    "test-19-xcards-audio",
			url:     fmt.Sprintf("%s/test-19-xcards-audio.html", server.URL),
			content: nil,
			err:     nil,
			extracted: map[Syntax]any{
				"opengraph": nil,
				"xcards": &extract.XCards{
					Type:        `website`,
					Title:       `go-microdata-extract`,
					URL:         `https://github.com/aafeher/go-microdata-extract`,
					Description: `X Cards with audio`,
					XCardsAudio: []extract.XCardsAudio{
						{
							URL: "https://example.com/sound.mp3",
						},
						{
							URL:       "https://example.com/sound2.mp3",
							SecureURL: "https://secure.example.com/sound2.mp3",
							Type:      "audio/mpeg",
						},
					},
				},
				"json-ld":   []map[string]any(nil),
				"microdata": []extract.MicrodataItem(nil),
			},
			errs: nil,
		},
		{
			name:    "test-20-xcards-music-song",
			url:     fmt.Sprintf("%s/test-20-xcards-music-song.html", server.URL),
			content: nil,
			err:     nil,
			extracted: map[Syntax]any{
				"opengraph": nil,
				"xcards": &extract.XCards{
					Type:     `music.song`,
					Title:    `Under Pressure`,
					URL:      `http://open.spotify.com/track/2aSFLiDPreOVP6KHiWk4lF`,
					SiteName: "Spotify",
					XCardsImage: []extract.XCardsImage{
						{
							URL: "http://o.scdn.co/image/e4c7b06c20c17156e46bbe9a71eb0703281cf345",
						},
					},
					Music: &extract.Music{
						Album:      "http://open.spotify.com/album/7rq68qYz66mNdPfidhIEFa",
						AlbumDisc:  1,
						AlbumTrack: 2,
						Duration:   236,
						Musician: []string{
							"http://open.spotify.com/artist/1dfeR4HaWDbWqFHLkxsg1d",
							"http://open.spotify.com/artist/0oSGxfWSnnOXhD2fKuz2Gy",
						},
					},
				},
				"json-ld":   []map[string]any(nil),
				"microdata": []extract.MicrodataItem(nil),
			},
			errs: nil,
		},
		{
			name:    "test-21-xcards-music-album",
			url:     fmt.Sprintf("%s/test-21-xcards-music-album.html", server.URL),
			content: nil,
			err:     nil,
			extracted: map[Syntax]any{
				"opengraph": nil,
				"xcards": &extract.XCards{
					Type:        `music.album`,
					Title:       `Greatest Hits II`,
					URL:         `http://open.spotify.com/album/7rq68qYz66mNdPfidhIEFa`,
					Description: `Greatest Hits II, an album by Queen on Spotify.`,
					SiteName:    "Spotify",
					XCardsImage: []extract.XCardsImage{
						{
							URL: "http://o.scdn.co/image/e4c7b06c20c17156e46bbe9a71eb0703281cf345",
						},
					},
					Music: &extract.Music{
						Musician: []string{
							"http://open.spotify.com/artist/1dfeR4HaWDbWqFHLkxsg1d",
						},
						Song: []extract.MusicSong{
							{
								URL:   "http://open.spotify.com/track/0pfHfdUNVwlXA0WDXznm2C",
								Disc:  1,
								Track: 1,
							},
							{
								URL:   "http://open.spotify.com/track/2aSFLiDPreOVP6KHiWk4lF",
								Disc:  1,
								Track: 2,
							},
						},
						ReleaseDate: "2011-04-19",
					},
				},
				"json-ld":   []map[string]any(nil),
				"microdata": []extract.MicrodataItem(nil),
			},
			errs: nil,
		},
		{
			name:    "test-22-xcards-music-playlist",
			url:     fmt.Sprintf("%s/test-22-xcards-music-playlist.html", server.URL),
			content: nil,
			err:     nil,
			extracted: map[Syntax]any{
				"opengraph": nil,
				"xcards": &extract.XCards{
					Type:     `music.playlist`,
					Title:    `on repeat`,
					URL:      `http://open.spotify.com/user/austinhaugen/playlist/1a8444uyNXVOpwtFdgakhv`,
					SiteName: "Spotify",
					XCardsImage: []extract.XCardsImage{
						{
							URL: "http://o.scdn.co/300/756df3afcb3d14cb362448b68ed2f5506479f313",
						},
					},
					Music: &extract.Music{
						Creator: []string{
							"http://open.spotify.com/user/austinhaugen",
						},
					},
				},
				"json-ld":   []map[string]any(nil),
				"microdata": []extract.MicrodataItem(nil),
			},
			errs: nil,
		},
		{
			name:    "test-23-xcards-video-movie",
			url:     fmt.Sprintf("%s/test-23-xcards-video-movie.html", server.URL),
			content: nil,
			err:     nil,
			extracted: map[Syntax]any{
				"opengraph": nil,
				"xcards": &extract.XCards{
					Type:     `video.movie`,
					Title:    `X Cards Video Movie Title`,
					URL:      `https://www.example.com/videos/video-movie-title`,
					SiteName: "SiteName",
					Video: &extract.Video{
						Actor: []extract.VideoActor{
							{
								URL:  "https://www.example.com/actors/@firstnameA-lastnameA",
								Role: "ant",
							},
							{
								URL:  "https://www.example.com/actors/@firstnameB-lastnameB",
								Role: "bear",
							},
						},
						Director: []string{
							"https://www.example.com/actors/@firstnameA-lastnameA",
							"https://www.example.com/actors/@firstnameB-lastnameB",
						},
						Writer: []string{
							"https://www.example.com/actors/@firstnameA-lastnameA",
							"https://www.example.com/actors/@firstnameB-lastnameB",
						},
						Duration:    42,
						ReleaseDate: time.Date(2024, 10, 31, 0, 0, 0, 0, time.UTC),
						Tag: []string{
							"tag A",
							"tag B",
						},
					},
				},
				"json-ld":   []map[string]any(nil),
				"microdata": []extract.MicrodataItem(nil),
			},
			errs: nil,
		},
		{
			name:    "test-24-xcards-video-episode",
			url:     fmt.Sprintf("%s/test-24-xcards-video-episode.html", server.URL),
			content: nil,
			err:     nil,
			extracted: map[Syntax]any{
				"opengraph": nil,
				"xcards": &extract.XCards{
					Type:     `video.episode`,
					Title:    `X Cards Video Episode Title`,
					URL:      `https://www.example.com/videos/video-episode-title`,
					SiteName: "SiteName",
					Video: &extract.Video{
						Actor: []extract.VideoActor{
							{
								URL:  "https://www.example.com/actors/@firstnameA-lastnameA",
								Role: "ant",
							},
							{
								URL:  "https://www.example.com/actors/@firstnameB-lastnameB",
								Role: "bear",
							},
						},
						Director: []string{
							"https://www.example.com/actors/@firstnameA-lastnameA",
							"https://www.example.com/actors/@firstnameB-lastnameB",
						},
						Writer: []string{
							"https://www.example.com/actors/@firstnameA-lastnameA",
							"https://www.example.com/actors/@firstnameB-lastnameB",
						},
						Duration:    42,
						ReleaseDate: time.Date(2024, 10, 31, 0, 0, 0, 0, time.UTC),
						Tag: []string{
							"tag A",
							"tag B",
						},
						Series: "Video Series",
					},
				},
				"json-ld":   []map[string]any(nil),
				"microdata": []extract.MicrodataItem(nil),
			},
			errs: nil,
		},
		{
			name:    "test-25-xcards-article",
			url:     fmt.Sprintf("%s/test-25-xcards-article.html", server.URL),
			content: nil,
			err:     nil,
			extracted: map[Syntax]any{
				"opengraph": nil,
				"xcards": &extract.XCards{
					Type:     `article`,
					Title:    `X Cards Article Title`,
					URL:      `https://www.example.com/article/article-title`,
					SiteName: "SiteName",
					Article: &extract.Article{
						PublishedTime:  time.Date(2024, 10, 01, 0, 0, 0, 0, time.UTC),
						ModifiedTime:   time.Date(2024, 10, 31, 0, 0, 0, 0, time.UTC),
						ExpirationTime: time.Date(2024, 11, 01, 0, 0, 0, 0, time.UTC),
						Author: []string{
							"https://www.example.com/profileAuthorA.html",
							"https://www.example.com/profileAuthorB.html",
						},
						Section: "Front page",
						Tag: []string{
							"tag A",
							"tag B",
						},
					},
				},
				"json-ld":   []map[string]any(nil),
				"microdata": []extract.MicrodataItem(nil),
			},
			errs: nil,
		},
		{
			name:    "test-26-xcards-book",
			url:     fmt.Sprintf("%s/test-26-xcards-book.html", server.URL),
			content: nil,
			err:     nil,
			extracted: map[Syntax]any{
				"opengraph": nil,
				"xcards": &extract.XCards{
					Type:     `book`,
					Title:    `X Cards Book Title`,
					URL:      `https://www.example.com/book/book-title`,
					SiteName: "SiteName",
					Book: &extract.Book{
						Author: []string{
							"https://www.example.com/profileAuthorA.html",
							"https://www.example.com/profileAuthorB.html",
						},
						ReleaseDate: time.Date(2024, 10, 31, 0, 0, 0, 0, time.UTC),
						ISBN:        "9871234567890",
						Tag: []string{
							"tag A",
							"tag B",
						},
					},
				},
				"json-ld":   []map[string]any(nil),
				"microdata": []extract.MicrodataItem(nil),
			},
			errs: nil,
		},
		{
			name:    "test-27-xcards-profile",
			url:     fmt.Sprintf("%s/test-27-xcards-profile.html", server.URL),
			content: nil,
			err:     nil,
			extracted: map[Syntax]any{
				"opengraph": nil,
				"xcards": &extract.XCards{
					Type:     `profile`,
					Title:    `X Cards Profile Title`,
					URL:      `https://www.example.com/profiles/profile-title`,
					SiteName: "SiteName",
					Profile: &extract.Profile{
						FirstName: "John",
						LastName:  "Doe",
						Username:  "johndoe",
						Gender:    "male",
					},
				},
				"json-ld":   []map[string]any(nil),
				"microdata": []extract.MicrodataItem(nil),
			},
			errs: nil,
		},
		{
			name:    "test-28-xcards-errors",
			url:     fmt.Sprintf("%s/test-28-xcards-errors.html", server.URL),
			content: nil,
			err:     nil,
			extracted: map[Syntax]any{
				"opengraph": nil,
				"xcards": &extract.XCards{
					Type:     `video.movie`,
					Title:    `X Cards Errors Title`,
					URL:      `https://www.example.com/videos/video-movie-title`,
					SiteName: "SiteName",
					Video: &extract.Video{
						Duration:    0,
						ReleaseDate: time.Time{},
					},
				},
				"json-ld":   []map[string]any(nil),
				"microdata": []extract.MicrodataItem(nil),
			},
			errs: nil,
		},
		{
			name:    "test-29-ldjson-object",
			url:     fmt.Sprintf("%s/test-29-ldjson-object.html", server.URL),
			content: nil,
			err:     nil,
			extracted: map[Syntax]any{
				"opengraph": nil,
				"xcards":    nil,
				"json-ld": []map[string]any{
					{
						"@context": "https://schema.org",
						"address": map[string]any{
							"@type":           "PostalAddress",
							"addressLocality": "Colorado Springs",
							"addressRegion":   "CO",
							"postalCode":      "80840",
							"streetAddress":   "100 Main Street",
						},
						"email":       "info@example.com",
						"jobTitle":    "Research Assistant",
						"image":       "janedoe.jpg",
						"name":        "Jane Doe",
						"alumniOf":    "Dartmouth",
						"birthPlace":  "Philadelphia, PA",
						"birthDate":   "1979-10-12",
						"height":      "72 inches",
						"gender":      "female",
						"memberOf":    "Republican Party",
						"nationality": "Albanian",
						"telephone":   "(123) 456-6789",
						"url":         "http://www.example.com",
						"@type":       "Person",
						"colleague": []any{
							"http://www.example.com/JohnColleague.html",
							"http://www.example.com/JameColleague.html",
						},
						"sameAs": []any{
							"https://www.facebook.com/",
							"https://www.linkedin.com/",
							"http://twitter.com/",
							"http://instagram.com/",
							"https://plus.google.com/",
						},
					},
				},
				"microdata": []extract.MicrodataItem(nil),
			},
			errs: nil,
		},
		{
			name:    "test-30-ldjson-array",
			url:     fmt.Sprintf("%s/test-30-ldjson-array.html", server.URL),
			content: nil,
			err:     nil,
			extracted: map[Syntax]any{
				"opengraph": nil,
				"xcards":    nil,
				"json-ld": []map[string]any{
					{
						"@context": "https://schema.org",
						"address": map[string]any{
							"@type":           "PostalAddress",
							"addressLocality": "Colorado Springs",
							"addressRegion":   "CO",
							"postalCode":      "80840",
							"streetAddress":   "100 Main Street",
						},
						"email":       "info@example.com",
						"jobTitle":    "Research Assistant",
						"image":       "janedoe.jpg",
						"name":        "Jane Doe",
						"alumniOf":    "Dartmouth",
						"birthPlace":  "Philadelphia, PA",
						"birthDate":   "1979-10-12",
						"height":      "72 inches",
						"gender":      "female",
						"memberOf":    "Republican Party",
						"nationality": "Albanian",
						"telephone":   "(123) 456-6789",
						"url":         "http://www.example.com",
						"@type":       "Person",
						"colleague": []any{
							"http://www.example.com/JohnColleague.html",
							"http://www.example.com/JameColleague.html",
						},
						"sameAs": []any{
							"https://www.facebook.com/",
							"https://www.linkedin.com/",
							"http://twitter.com/",
							"http://instagram.com/",
							"https://plus.google.com/",
						},
					},
				},
				"microdata": []extract.MicrodataItem(nil),
			},
			errs: nil,
		},
		{
			name:    "test-31-ldjson-multiple-objects",
			url:     fmt.Sprintf("%s/test-31-ldjson-multiple-objects.html", server.URL),
			content: nil,
			err:     nil,
			extracted: map[Syntax]any{
				"opengraph": nil,
				"xcards":    nil,
				"json-ld": []map[string]any{
					{
						"@context": "https://schema.org",
						"name":     "John Doe",
						"@type":    "Person",
					},
					{
						"@context": "https://schema.org",
						"name":     "Jane Doe",
						"@type":    "Person",
					},
				},
				"microdata": []extract.MicrodataItem(nil),
			},
			errs: nil,
		},
		{
			name:    "test-32-ldjson-errors",
			url:     fmt.Sprintf("%s/test-32-ldjson-errors.html", server.URL),
			content: nil,
			err:     nil,
			extracted: map[Syntax]any{
				"opengraph": nil,
				"xcards":    nil,
				"json-ld":   []map[string]any(nil),
				"microdata": []extract.MicrodataItem(nil),
			},
			errs: []error{
				func() error {
					var jsonData []map[string]any
					jsonLD := `[
        {
            "@context": "https://schema.org",
            "@type": "Person",
            "name": "John Doe",
        #}
    ]`
					if err := json.Unmarshal([]byte(jsonLD), &jsonData); err != nil {
						return err
					}
					return nil
				}(),
				func() error {
					var jsonData []map[string]any
					jsonLD := `{
        "@context": "https://schema.org",
        "@type": "Person",
        "name": "John Doe",
    }]`
					if err := json.Unmarshal([]byte(jsonLD), &jsonData); err != nil {
						return err
					}
					return nil
				}(),
			},
		},
		{
			name:    "test-33-w3cmicrodata-simple",
			url:     fmt.Sprintf("%s/test-33-w3cmicrodata-simple.html", server.URL),
			content: nil,
			err:     nil,
			extracted: map[Syntax]any{
				"opengraph": nil,
				"xcards":    nil,
				"json-ld":   []map[string]any(nil),
				"microdata": []extract.MicrodataItem{
					{
						Type: "https://schema.org/SoftwareApplication",
						Properties: map[string]any{
							"name":                "Angry Birds",
							"operatingSystem":     "ANDROID",
							"applicationCategory": "",
							"aggregateRating": &extract.MicrodataItem{
								Type: "https://schema.org/AggregateRating",
								ID:   nil,
								Properties: map[string]any{
									"ratingValue": "4.6",
									"ratingCount": "8864",
								},
							},
							"offers": &extract.MicrodataItem{
								Type: "https://schema.org/Offer",
								ID:   nil,
								Properties: map[string]any{
									"price":         "1.00",
									"priceCurrency": "",
								},
							},
						},
					},
				},
			},
			errs: nil,
		},
		{
			name:    "test-34-w3cmicrodata-extended",
			url:     fmt.Sprintf("%s/test-34-w3cmicrodata-extended.html", server.URL),
			content: nil,
			err:     nil,
			extracted: map[Syntax]any{
				"opengraph": nil,
				"xcards":    nil,
				"json-ld":   []map[string]any(nil),
				"microdata": []extract.MicrodataItem{
					{
						Type: "https://schema.org/SoftwareApplication",
						Properties: map[string]any{
							"name":                "Angry Birds",
							"operatingSystem":     "ANDROID",
							"downloadUrl":         fmt.Sprintf("%s/download", server.URL),
							"applicationCategory": "",
							"aggregateRating": &extract.MicrodataItem{
								Type: "https://schema.org/AggregateRating",
								ID:   nil,
								Properties: map[string]any{
									"ratingValue": "4.6",
									"ratingCount": "8864",
								},
							},
							"offers": &extract.MicrodataItem{
								Type: "https://schema.org/Offer",
								ID:   nil,
								Properties: map[string]any{
									"price":         "1.00",
									"priceCurrency": "",
								},
							},
						},
					},
				},
			},
			errs: nil,
		},
		{
			name:    "test-35-w3cmicrodata-book",
			url:     fmt.Sprintf("%s/test-35-w3cmicrodata-book.html", server.URL),
			content: nil,
			err:     nil,
			extracted: map[Syntax]any{
				"opengraph": nil,
				"xcards":    nil,
				"json-ld":   []map[string]any(nil),
				"microdata": []extract.MicrodataItem{
					{
						ID: pointerOfString("urn:isbn:0-374-22848-5\u003c"),
						Properties: map[string]any{
							"author":        "Jonathan C Slaght",
							"datePublished": "2020-08-04",
							"title":         "Owls of the Eastern Ice",
							"discussionUrl": "//www.example.com/book/discussion",
						},
						Type: "https://schema.org/Book",
					},
				},
			},
			errs: nil,
		},
		{
			name:    "test-36-w3cmicrodata-organization",
			url:     fmt.Sprintf("%s/test-36-w3cmicrodata-organization.html", server.URL),
			content: nil,
			err:     nil,
			extracted: map[Syntax]any{
				"opengraph": nil,
				"xcards":    nil,
				"json-ld":   []map[string]any(nil),
				"microdata": []extract.MicrodataItem{
					{
						ID: pointerOfString("http://example.com/org/1"),
						Properties: map[string]any{
							"employee": &extract.MicrodataItem{
								Type: "http://schema.org/Person",
								ID:   pointerOfString("http://example.com/person/1"),
								Properties: map[string]any{
									"name": "John Doe",
								},
							},
							"name": "Example Organization",
						},
						Type: "http://schema.org/Organization",
					},
				},
			},
			errs: nil,
		},
		{
			name:    "test-37-w3cmicrodata-product",
			url:     fmt.Sprintf("%s/test-37-w3cmicrodata-product.html", server.URL),
			content: nil,
			err:     nil,
			extracted: map[Syntax]any{
				"opengraph": nil,
				"xcards":    nil,
				"json-ld":   []map[string]any(nil),
				"microdata": []extract.MicrodataItem{
					{
						Type: "http://schema.org/Product",
						Properties: map[string]any{
							"aggregateRating": &extract.MicrodataItem{
								Type: "http://schema.org/AggregateRating",
								Properties: map[string]any{
									"ratingValue": "3.5",
									"reviewCount": "11",
								},
							},
							"name":       "Panasonic White 60L Refrigerator",
							"product-id": "9678AOU879",
						},
					},
				},
			},
			errs: nil,
		},
		{
			name:    "test-38-w3cmicrodata-multiple-itemprop",
			url:     fmt.Sprintf("%s/test-38-w3cmicrodata-multiple-itemprop.html", server.URL),
			content: nil,
			err:     nil,
			extracted: map[Syntax]any{
				"opengraph": nil,
				"xcards":    nil,
				"json-ld":   []map[string]any(nil),
				"microdata": []extract.MicrodataItem{
					{
						Properties: map[string]any{
							"flavor": []any{
								"Lemon sorbet",
								"Apricot sorbet",
							},
							"color": []any{
								"yellow",
								"green",
								"purple",
							},
						},
					},
				},
			},
			errs: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			e := New()
			e, err := e.Extract(test.url, test.content)
			if err != nil {
				if err.Error() != *test.err {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			extracted := e.GetExtracted()

			if extracted == nil {
				t.Fatal("Expected no nil map, but got nil")
			}
			if e.url != test.url {
				t.Fatalf("Expected URL to be %s, but got %s", test.url, e.url)
			}

			if !reflect.DeepEqual(extracted, test.extracted) {
				extractedJSON, _ := json.MarshalIndent(extracted, "", "  ")
				testExtractedJSON, _ := json.MarshalIndent(test.extracted, "", "  ")
				_ = extractedJSON
				_ = testExtractedJSON
				t.Error("extracted is not equal to expected value")
			}
			if !reflect.DeepEqual(e.errs, test.errs) {
				t.Error("errs is not equal to expected value")
			}
		})
	}
}

func TestExtractor_setContent(t *testing.T) {
	server := testServer()
	defer server.Close()

	tests := []struct {
		name           string
		setup          func() *Extractor
		attrURLContent *string
		wantURLContent string
		wantErr        error
	}{
		{
			name: "setContent_with_urlContent",
			setup: func() *Extractor {
				return &Extractor{
					url: fmt.Sprintf("%s/example", server.URL),
				}
			},
			attrURLContent: pointerOfString("URL Content"),
			wantURLContent: "URL Content",
			wantErr:        nil,
		},
		{
			name: "setContent_without_urlContent",
			setup: func() *Extractor {
				return &Extractor{
					url: fmt.Sprintf("%s/example", server.URL),
				}
			},
			attrURLContent: nil,
			wantURLContent: "example content\n",
			wantErr:        nil,
		},
		{
			name: "setContent_without_urlContent_with_invalid_mainURL",
			setup: func() *Extractor {
				return &Extractor{
					url: fmt.Sprintf("%s/404", server.URL),
				}
			},
			attrURLContent: nil,
			wantURLContent: "",
			wantErr:        fmt.Errorf("received HTTP status 404"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := test.setup()
			retURLContent, err := s.setContent(test.attrURLContent)
			if retURLContent != test.wantURLContent {
				t.Errorf("unexpected urlContent: got %v, want %v", retURLContent, test.wantURLContent)
			}
			if err != nil && test.wantErr != nil {
				if err.Error() != test.wantErr.Error() {
					t.Errorf("unexpected err: got %v, want %v", err, test.wantErr)
				}
			} else if err != nil && test.wantErr == nil {
				t.Errorf("unexpected err: got %v, want %v", err, test.wantErr)
			} else if err == nil && test.wantErr != nil {
				t.Errorf("unexpected err: got %v, want %v", err, test.wantErr)
			}
		})
	}
}

func TestExtractor_fetch(t *testing.T) {
	server := testServer()
	defer server.Close()

	e := Extractor{cfg: config{fetchTimeout: 3}}
	type fields struct {
		cfg config
	}
	tests := []struct {
		name    string
		fields  fields
		url     string
		wantErr bool
	}{
		{
			name:    "Empty URL",
			fields:  fields{e.cfg},
			url:     "",
			wantErr: true,
		},
		{
			name:    "Invalid URL",
			fields:  fields{e.cfg},
			url:     "https:bad_domain",
			wantErr: true,
		},
		{
			name:    "404 HTTP response",
			fields:  fields{e.cfg},
			url:     fmt.Sprintf("%s/404", server.URL),
			wantErr: true,
		},
		{
			name:    "Expected HTTP Response",
			fields:  fields{e.cfg},
			url:     fmt.Sprintf("%s/test-01-opengraph-minimal.html", server.URL),
			wantErr: false,
		},
		{
			name:    "Timeout URL",
			fields:  fields{config{fetchTimeout: 0}},
			url:     fmt.Sprintf("%s/test-01-opengraph-minimal.html", server.URL),
			wantErr: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			e := &Extractor{
				cfg: test.fields.cfg,
			}
			_, err := e.fetch(test.url)
			if (err != nil) != test.wantErr {
				t.Errorf("fetch() error = %v, wantErr %v", err, test.wantErr)
				return
			}
		})
	}
}

func TestExtractor_GetExtracted(t *testing.T) {
	tests := []struct {
		name  string
		setup func() *Extractor
		want  map[Syntax]any
	}{
		{
			name: "extracted map initialized",
			setup: func() *Extractor {
				return &Extractor{
					extracted: map[Syntax]any{
						"key1": "value1",
						"key2": "value2",
					},
				}
			},
			want: map[Syntax]any{
				"key1": "value1",
				"key2": "value2",
			},
		},
		{
			name: "extracted map not initialized",
			setup: func() *Extractor {
				return &Extractor{}
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := tt.setup()
			if got := e.GetExtracted(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Extractor.GetExtracted() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractor_GetExtractedJSON(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *Extractor
		want    json.RawMessage
		wantErr bool
	}{
		{
			name: "extracted map initialized",
			setup: func() *Extractor {
				tmp := &Extractor{
					extracted: map[Syntax]any{
						"key1": "value1",
						"key2": "value2",
					},
				}
				return tmp
			},
			want: json.RawMessage(`{
  "key1": "value1",
  "key2": "value2"
}`),
			wantErr: false,
		},
		{
			name: "empty extracted map",
			setup: func() *Extractor {
				tmp := &Extractor{
					extracted: map[Syntax]any{},
				}
				return tmp
			},
			want:    json.RawMessage("{}"),
			wantErr: false,
		},
		{
			name: "nil extracted map",
			setup: func() *Extractor {
				return &Extractor{}
			},
			want:    json.RawMessage("null"),
			wantErr: false,
		},
		{
			name: "error",
			setup: func() *Extractor {
				return &Extractor{
					extracted: map[Syntax]any{
						"key1": struct {
							Channel chan int
							Name    string
						}{
							Channel: make(chan int),
							Name:    "John",
						},
					},
				}
			},
			want:    json.RawMessage(nil),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := tt.setup()
			if got := e.GetExtractedJSON(); !bytes.Equal(got, tt.want) {
				t.Errorf("Extractor.GetExtractedJSON() = %v, want %v", string(got), string(tt.want))
			}
			if len(e.errs) == 0 && tt.wantErr {
				t.Errorf("Extractor.GetExtractedJSON() error = %v, wantErr %v", e.errs, tt.wantErr)
			}
			if len(e.errs) > 0 && !tt.wantErr {
				t.Errorf("Extractor.GetExtractedJSON() error = %v, wantErr %v", e.errs, tt.wantErr)
			}
		})
	}
}

func Test_index(t *testing.T) {
	tests := []struct {
		name string
		s    []int
		v    int
		want int
	}{
		{
			name: "element found",
			s:    []int{1, 2, 3, 4, 5},
			v:    3,
			want: 2,
		},
		{
			name: "element not found",
			s:    []int{1, 2, 3, 4, 5},
			v:    6,
			want: -1,
		},
		{
			name: "empty slice",
			s:    []int{},
			v:    1,
			want: -1,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := index(test.s, test.v)
			if got != test.want {
				t.Errorf("Name: %s, Expected: %v, Got: %v", test.name, test.want, got)
			}
		})
	}
}

func Test_contains(t *testing.T) {
	tests := []struct {
		name string
		s    []int
		v    int
		want bool
	}{
		{
			name: "element found",
			s:    []int{1, 2, 3, 4, 5},
			v:    3,
			want: true,
		},
		{
			name: "element not found",
			s:    []int{1, 2, 3, 4, 5},
			v:    6,
			want: false,
		},
		{
			name: "empty slice",
			s:    []int{},
			v:    1,
			want: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := contains(test.s, test.v)
			if got != test.want {
				t.Errorf("Name: %s, Expected: %v, Got: %v", test.name, test.want, got)
			}
		})
	}
}

func pointerOfString(str string) *string {
	return &str
}

func areSyntaxSlicesEqual(slice1, slice2 []Syntax) bool {
	if len(slice1) != len(slice2) {
		return false
	}

	for i := range slice1 {
		if slice1[i] != slice2[i] {
			return false
		}
	}

	return true
}
