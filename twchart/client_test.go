package twchart

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestJSON(t *testing.T) {
	rawJSON := "{\"id\":\"d4kdisifn76c73dkrju0\",\"Session\":{\"Name\":\"Test Bean\",\"Date\":\"2025-11-27T16:06:26.504207-07:00\",\"StartTime\":\"0001-01-01T00:00:00Z\",\"Probes\":[{\"Name\":\"Ambient\",\"Position\":1},{\"Name\":\"Roaster\",\"Position\":2}],\"Stages\":null,\"Events\":null,\"Data\":null},\"UploadedAt\":\"2025-11-27T23:06:26.60698014Z\"}"
	var s session
	err := json.Unmarshal([]byte(rawJSON), &s)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	fmt.Println(s)
}
