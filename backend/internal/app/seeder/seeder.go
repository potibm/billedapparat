//nolint:mnd // we want to keep the numbers in the seeder for better control over the generated data
package seeder

import (
	"context"
	"log/slog"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/potibm/billedapparat/internal/app/domain"
	"github.com/potibm/billedapparat/internal/app/repository"
)

const (
	numberOfSponsorSlides     = 10
	numberOfSceneFriendSlides = 25
)

type Seeder struct {
	slides    []domain.Slide
	currentID int64
	repo      repository.SlideRepository
}

func NewSeeder(repo repository.SlideRepository) *Seeder {
	return &Seeder{
		slides:    []domain.Slide{},
		currentID: 0,
		repo:      repo,
	}
}

func (s *Seeder) Run() error {
	slog.Info("Starting DB Purge & Seed...")

	ctx := context.Background()

	_ = gofakeit.Seed(0)

	if err := s.generateAndSave(ctx, numberOfSponsorSlides, s.buildSponsorSlide); err != nil {
		return err
	}

	if err := s.generateAndSave(ctx, numberOfSceneFriendSlides, s.buildSceneFriendSlide); err != nil {
		return err
	}

	slog.Info("Seeding finished successfully", "total_slides", s.currentID)

	return nil
}

func (s *Seeder) nextID() int64 {
	s.currentID++

	return s.currentID
}

func (s *Seeder) generateAndSave(ctx context.Context, count int, builder func(id int64) (domain.Slide, error)) error {
	for i := 0; i < count; i++ {
		id := s.nextID()

		slide, err := builder(id)
		if err != nil {
			return err
		}

		if err := s.repo.Save(ctx, &slide); err != nil {
			slog.Error("Failed to save slide", "error", err, "slide_id", slide.ID)

			return err
		}
	}

	return nil
}

func (s *Seeder) buildSponsorSlide(id int64) (domain.Slide, error) {
	imageURL, err := s.getFakeImageURL(id, false)
	if err != nil {
		return domain.Slide{}, err
	}

	title := gofakeit.Company()

	return domain.Slide{
		ID: id,
		Content: domain.Content{
			Title: title,
			Type:  domain.TypeSponsor,
			Media: &domain.Media{
				LocalURL: imageURL,
				MimeType: "image/webp",
			},
		},
		Status: "active",
		DisplayOptions: domain.DisplayOptions{
			AllowSocialOverlay: false,
			Priority:           gofakeit.Number(3, 10),
			IsUrgent:           false,
		},
	}, nil
}

func (s *Seeder) buildSceneFriendSlide(id int64) (domain.Slide, error) {
	imageURL, err := s.getFakeImageURL(id, true)
	if err != nil {
		return domain.Slide{}, err
	}

	title := gofakeit.Gamertag()

	return domain.Slide{
		ID: id,
		Content: domain.Content{
			Title: title,
			Type:  domain.TypeSponsor,
			Media: &domain.Media{
				LocalURL: imageURL,
				MimeType: "image/webp",
			},
		},
		Status: "active",
		DisplayOptions: domain.DisplayOptions{
			AllowSocialOverlay: true,
			Priority:           1,
			IsUrgent:           false,
		},
	}, nil
}
