package generator

import (
	"context"
	"log/slog"

	"github.com/potibm/billedapparat/internal/app/domain"
	"github.com/potibm/billedapparat/internal/app/repository"
)

type newsGenerator struct {
	newsRepo repository.NewsRepository
	logger   *slog.Logger
}

func NewNewsGenerator(newsRepo repository.NewsRepository, logger *slog.Logger) *newsGenerator {
	return &newsGenerator{
		newsRepo: newsRepo,
		logger:   logger,
	}
}

func (g *newsGenerator) Name() string {
	return "news-generator"
}

func (g *newsGenerator) Generate(ctx context.Context) ([]domain.Slide, error) {
	newsList, err := g.newsRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	var slides []domain.Slide

	for _, n := range newsList {
		slide := g.newsToSlide(n)
		slides = append(slides, slide)
	}

	g.logger.Info("Generated slides from news", "count", len(slides))

	return slides, nil
}

func (g *newsGenerator) newsToSlide(n domain.News) domain.Slide {
	status := domain.StatusActive
	if n.IsHidden {
		status = domain.StatusInactive
	}

	subID := 0

	return domain.Slide{
		Content: domain.Content{
			Title: n.Title,
			Body:  n.Body,
			Type:  domain.TypeNews,
			Media: nil,
		},
		Status: status,
		DisplayOptions: domain.DisplayOptions{
			AllowSocialOverlay: true,
			Priority:           1,
			IsUrgent:           n.IsUrgent,
		},
		Source:        g.Name(),
		ExternalID:    n.ExternalID,
		ExternalSubID: &subID,
		Author:        nil,
	}
}
