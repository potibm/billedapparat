package pouet

import (
	"log/slog"
	"os"
	"testing"

	"github.com/potibm/billedapparat/internal/app/contracts"
	"github.com/stretchr/testify/assert"
)

func TestParse_FromRealHTML(t *testing.T) {
	file, err := os.Open("testdata/20260530-oneliner.html")
	if err != nil {
		t.Fatalf("Unable to open file: %v", err)
	}
	defer file.Close()

	result, err := parse(slog.Default(), file)
	if err != nil {
		t.Fatalf("Error parsing HTML: %v", err)
	}

	assert.Nil(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 19)

	const externalURLPrefix = "https://www.pouet.net/oneliner.php#"

	expectedSecondNoHTML := contracts.IngestSlideRequest{
		Source:          collectorName,
		ExternalID:      externalURLPrefix + "733fc74806d629c29d75eac9164432020e55d2400210eefa1b062e85902c9d52",
		Body:            "Ha...  ha.......",
		Language:        "en",
		MediaURLs:       nil,
		OriginCreatedAt: result[1].OriginCreatedAt,
		Author: &contracts.IngestSlideRequestAuthor{
			ExternalID:        "https://pouet.net/user.php?who=3153",
			DisplayName:       "sim",
			Username:          "sim",
			AvatarExternalURL: "https://content.pouet.net/avatars/budda5.gif",
		},
	}

	assert.Equal(t, expectedSecondNoHTML, result[1])

	expectedFirstWithHTML := contracts.IngestSlideRequest{
		Source:     collectorName,
		ExternalID: externalURLPrefix + "275596469a22cd65b23f854bef25dfb79fe1af030543be16454a4b634148dfc3",
		Body: "Just a wonderful SID tune: https://deepsid.chordian.net/" +
			"?file=/MUSICIANS/N/Nygaard_Richard/Thats_the_Wave_I t_Is.sid",
		Language:        "en",
		MediaURLs:       nil,
		OriginCreatedAt: result[0].OriginCreatedAt,
		Author: &contracts.IngestSlideRequestAuthor{
			ExternalID:        "https://pouet.net/user.php?who=102874",
			DisplayName:       "G-Fellow",
			Username:          "G-Fellow",
			AvatarExternalURL: "https://content.pouet.net/avatars/g-f5.gif",
		},
	}
	assert.Equal(t, expectedFirstWithHTML, result[0])
}
