package seeder

import (
	"context"
	"log/slog"
	"os"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/potibm/billedapparat/internal/app/config"
	"github.com/potibm/billedapparat/internal/app/domain"
	"github.com/potibm/billedapparat/internal/app/repository"
)

const (
	numberOfSponsorSlides     = 10
	numberOfSceneFriendSlides = 25
	numberOfNewsSlides        = 10
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

	if err := os.MkdirAll(config.MediaDirname, config.DataDirPerm); err != nil {
		return err
	}

	_ = gofakeit.Seed(0)

	if err := s.generateAndSave(ctx, numberOfSponsorSlides, s.buildSponsorSlide); err != nil {
		return err
	}

	if err := s.generateAndSave(ctx, numberOfSceneFriendSlides, s.buildSceneFriendSlide); err != nil {
		return err
	}

	if err := s.generateAndSave(ctx, numberOfNewsSlides, s.buildNewsSlide); err != nil {
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
			Type: domain.TypeSponsor,
		},
		MediaURLOriginal: imageURL,
		Status:           "active",
		DisplayOptions: domain.DisplayOptions{
			AllowSocialOverlay: false,
			//nolint:mnd // mnd: sponsor slides between 3 and 10, randomized to create some variation in the seed data
			Priority: gofakeit.Number(3, 10),
			IsUrgent: false,
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
			Type: domain.TypeSponsor,
		},
		MediaURLOriginal: imageURL,
		Status:           "active",
		DisplayOptions: domain.DisplayOptions{
			AllowSocialOverlay: true,
			Priority:           1,
			IsUrgent:           false,
		},
	}, nil
}

func (s *Seeder) buildNewsSlide(id int64) (domain.Slide, error) {

	text := gofakeit.Paragraph(1)

	// second level heading 
	if gofakeit.Bool() {
		text += "\n\n"
		text += "## " + gofakeit.Sentence(5) + "\n\n" + gofakeit.Paragraph(1)
	}

	// some bullet points
	if gofakeit.Bool() {
		text += "\n\n"
		text += "- " + gofakeit.Sentence(5) + "\n"
		text += "- " + gofakeit.Sentence(5) + "\n" 
		text += "- " + gofakeit.Sentence(5)
	}

	return domain.Slide{
		ID: id,
		Content: domain.Content{
			Title: gofakeit.Sentence(5),
			Body:  text,
			Type:  domain.TypeNews,
		},
		Status:           "active",
		MediaURLOriginal: "",
		DisplayOptions: domain.DisplayOptions{
			AllowSocialOverlay: true,
			Priority:           1,
			IsUrgent:           false,
		},
	}, nil
}
