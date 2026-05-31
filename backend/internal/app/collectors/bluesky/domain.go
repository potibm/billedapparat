package bluesky

import "fmt"

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
	Type  string     `json:"$type"`
	Text  string     `json:"text"`
	Langs []string   `json:"langs,omitempty"`
	Embed *PostEmbed `json:"embed,omitempty"`
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
