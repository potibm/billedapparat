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

func parse(logger *slog.Logger, reader io.Reader) ([]contracts.IngestSlideRequest, error) {
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return nil, err
	}

	var result []contracts.IngestSlideRequest

	doc.Find("#pouetbox_onelinerview li").Each(func(i int, s *goquery.Selection) {
		if s.HasClass("day") {
			return
		}

		createdString := s.Find("time").AttrOr("datetime", "")
		if createdString == "" {
			logger.Warn("Missing datetime attribute in time element, skipping item", "index", i)

			return
		}

		created, err := time.Parse(time.DateTime, createdString)
		if err != nil {
			logger.Warn("Invalid datetime format in time element, skipping item", "index", i)

			return
		}

		userTag := s.Find("a.usera")
		if userTag.Length() == 0 {
			logger.Warn("Missing user link element, skipping item", "index", i)

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

	return strings.TrimSpace(s.Text())
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
