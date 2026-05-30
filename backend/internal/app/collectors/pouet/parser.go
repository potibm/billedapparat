package pouet

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/potibm/billedapparat/internal/app/contracts"
)

func parse(reader io.Reader) ([]contracts.IngestSlideRequest, error) {
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return nil, err
	}

	var result []contracts.IngestSlideRequest

	doc.Find("#pouetbox_onelinerview li").Each(func(i int, s *goquery.Selection) {
		if s.HasClass("day") {
			slog.Warn("Skipping item with 'date' class, likely a date separator", "index", i)

			return
		}

		createdString := s.Find("time").AttrOr("datetime", "")
		if createdString == "" {
			slog.Warn("Missing datetime attribute in time element, skipping item")

			return
		}

		created, err := time.Parse(time.DateTime, createdString)
		if err != nil {
			slog.Warn("Invalid datetime format in time element, skipping item")

			return
		}

		userTag := s.Find("a.usera")
		if userTag.Length() == 0 {
			slog.Warn("Missing user link element, skipping item")

			return
		}

		author := getAuthor(userTag)

		slide := contracts.IngestSlideRequest{
			Source:          pouetCollectorName,
			ExternalID:      getExternalID(author.ExternalID, createdString),
			Body:            getMessageHTML(s),
			Language:        "en-EN",
			OriginCreatedAt: created,
			Author:          author,
		}

		result = append(result, slide)
	})

	return result, nil
}

func getMessageHTML(s *goquery.Selection) string {
	// clone the selection to avoid modifying the original document
	s = s.Clone()

	s.Find("time").Remove()
	s.Find("a.usera").Remove()

	content, err := s.Html()
	if err != nil {
		slog.Warn("Error extracting HTML content, falling back to text", "error", err)

		content = s.Text()
	}

	return strings.TrimSpace(content)
}

func getAuthor(s *goquery.Selection) *contracts.IngestSlideRequestAuthor {
	author := contracts.IngestSlideRequestAuthor{
		ExternalID:        "https://pouet.net/" + s.AttrOr("href", ""),
		DisplayName:       s.AttrOr("title", ""),
		Username:          s.AttrOr("title", ""),
		AvatarExternalURL: s.Find("img").AttrOr("src", ""),
	}

	return &author
}

func getExternalID(authorID, createdString string) string {
	hash := sha256.Sum256([]byte(authorID + createdString))

	return fmt.Sprintf("%s#%x", pouetOnelinerURL, hash)
}
