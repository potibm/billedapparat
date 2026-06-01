package bluesky

import (
	"fmt"
	"strings"
)

type JetstreamEvent struct {
	Kind   string           `json:"kind"`
	Did    string           `json:"did"`
	Commit *JetstreamCommit `json:"commit,omitempty"`
}

type JetstreamCommit struct {
	Operation  string      `json:"operation"`
	Collection string      `json:"collection"`
	Record     *PostRecord `json:"record,omitempty"`
	Rkey       string      `json:"rkey"`
}

type PostRecord struct {
	Type      string     `json:"$type"`
	Text      string     `json:"text"`
	CreatedAt string     `json:"createdAt"`
	Langs     []string   `json:"langs,omitempty"`
	Embed     *PostEmbed `json:"embed,omitempty"`
	Facets    []Facet    `json:"facets,omitempty"`
}

type Facet struct {
	Features []FacetFeature `json:"features"`
}

type FacetFeature struct {
	Type string `json:"$type"`
	Tag  string `json:"tag,omitempty"`
}

type PostEmbed struct {
	Type   string       `json:"$type"`
	Images []EmbedImage `json:"images,omitempty"`
	Video  *MediaBlob   `json:"video,omitempty"`
}

type EmbedImage struct {
	Alt   string    `json:"alt"`
	Image MediaBlob `json:"image"`
}

type MediaBlob struct {
	Ref ImageRef `json:"ref"`
}

type ImageRef struct {
	Link string `json:"$link"`
}

type ProfileResponse struct {
	Handle      string `json:"handle"`
	DisplayName string `json:"displayName"`
	Avatar      string `json:"avatar"`
}

func (p *PostRecord) FirstLanguage() string {
	if len(p.Langs) > 0 {
		return p.Langs[0]
	}

	return ""
}

func (e *JetstreamEvent) ExtractImageURLs() []string {
	var urls []string

	if e.Commit == nil || e.Commit.Record == nil || e.Commit.Record.Embed == nil {
		return urls
	}

	if e.Commit.Record.Embed.Type != "app.bsky.embed.images" {
		return urls
	}

	for _, img := range e.Commit.Record.Embed.Images {
		cid := img.Image.Ref.Link
		if cid != "" {
			url := fmt.Sprintf("https://cdn.bsky.app/img/feed_fullsize/plain/%s/%s@jpeg", e.Did, cid)
			urls = append(urls, url)
		}
	}

	return urls
}

func (e *JetstreamEvent) ExtractVideoURLs() (streamURL, thumbnailURL string, hasVideo bool) {
	if e.Commit == nil || e.Commit.Record == nil || e.Commit.Record.Embed == nil {
		return "", "", false
	}

	if e.Commit.Record.Embed.Type != "app.bsky.embed.video" || e.Commit.Record.Embed.Video == nil {
		return "", "", false
	}

	cid := e.Commit.Record.Embed.Video.Ref.Link
	if cid == "" {
		return "", "", false
	}

	baseURL := fmt.Sprintf("https://video.bsky.app/watch/%s/%s", e.Did, cid)

	streamURL = fmt.Sprintf("%s/playlist.m3u8", baseURL)
	thumbnailURL = fmt.Sprintf("%s/thumbnail.jpg", baseURL)

	return streamURL, thumbnailURL, true
}

func (p *PostRecord) Hashtags() []string {
	var tags []string

	for _, facet := range p.Facets {
		for _, feature := range facet.Features {
			if feature.Type == "app.bsky.richtext.facet#tag" && feature.Tag != "" {
				tags = append(tags, strings.ToLower(feature.Tag))
			}
		}
	}

	return tags
}
