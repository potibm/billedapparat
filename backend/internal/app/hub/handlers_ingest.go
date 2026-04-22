package hub

import (
	"time"

	"github.com/potibm/billedapparat/internal/app/domain"
)

type IngestRequest struct {
	Source          string         `json:"source"            binding:"required"`
	ExternalID      string         `json:"external_id"       binding:"required"`
	Author          domain.Author  `json:"author"            binding:"required"`
	Content         domain.Content `json:"content"           binding:"required"`
	OriginCreatedAt time.Time      `json:"origin_created_at" binding:"required"`
}

/*
func (s *Server) collectorIngestSlide(c *gin.Context) {
	var req IngestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// 1. Duplicate Check: Haben wir Source + ExternalID + ExternalSubID schon?
	exists, err := s.db.SlideExists(req.Source, req.ExternalID, req.ExternalSubID)
	if err != nil {
		c.JSON(500, gin.H{"error": "database error"})
		return
	}
	if exists {
		// Kein Fehler, wir sagen dem Collector einfach "Alles gut, hab ich schon"
		c.JSON(200, gin.H{"status": "skipped", "reason": "duplicate"})
		return
	}

	// 2. Blacklist Check: Enthält req.Content.Text böse Wörter oder ist der Author gebannt?
	if s.blacklist.IsBanned(req) {
		c.JSON(200, gin.H{"status": "skipped", "reason": "blacklisted"})
		return
	}

	// 3. Mapping: Aus dem DTO einen echten domain.Slide machen
	slide := domain.Slide{
		Source:          req.Source,
		ExternalID:      req.ExternalID,
		ExternalSubID:   req.ExternalSubID,
		Author:          req.Author,
		Content:         req.Content,
		Status:          "pending", // Geht erst live, wenn Bilder da sind oder Admin freigibt
		OriginCreatedAt: req.OriginCreatedAt,
		CreatedAt:       time.Now(),
	}

	// 4. In die Datenbank speichern
	if err := s.db.SaveSlide(&slide); err != nil {
		c.JSON(500, gin.H{"error": "failed to save slide"})
		return
	}

	// 5. Asynchronen Download der Bilder anstoßen (Avatar & Content-Media)
	// Dazu gleich unten eine Frage an dich!
	go s.mediaDownloader.ProcessSlideMedia(slide.ID)

	// 6. Collector glücklich machen
	c.JSON(201, gin.H{"status": "ingested", "id": slide.ID})
}
*/
