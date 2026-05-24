package gorm

import "github.com/potibm/billedapparat/internal/app/domain"

type dbNews struct {
	GormModel

	Source     string `gorm:"uniqueIndex:news_idx_ext"`
	ExternalID string `gorm:"uniqueIndex:news_idx_ext"`

	Title       string
	Body        string
	IsUrgent    bool
	ExternalURL string
	IsHidden    bool
}

func (dbNews) TableName() string {
	return "news"
}

func fromDomainNews(n *domain.News) *dbNews {
	db := &dbNews{
		GormModel: GormModel{ID: n.ID},

		Source:     n.Source,
		ExternalID: n.ExternalID,

		Title:       n.Title,
		Body:        n.Body,
		IsUrgent:    n.IsUrgent,
		ExternalURL: n.ExternalURL,
		IsHidden:    n.IsHidden,
	}

	return db
}

func (n *dbNews) toDomain() *domain.News {
	result := &domain.News{
		ID: n.ID,

		Source:     n.Source,
		ExternalID: n.ExternalID,

		Title:       n.Title,
		Body:        n.Body,
		IsUrgent:    n.IsUrgent,
		ExternalURL: n.ExternalURL,
		IsHidden:    n.IsHidden,
	}

	return result
}

func toDomainNewsList(news []dbNews) []domain.News {
	result := make([]domain.News, len(news))
	for i, s := range news {
		result[i] = *s.toDomain()
	}

	return result
}
