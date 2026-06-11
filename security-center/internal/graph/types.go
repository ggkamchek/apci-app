package graph

const (
	LinkSameDevice       = "same_device"
	LinkSharedContact    = "shared_contact"
	LinkCommunicatesWith = "communicates_with"

	WeightSameDevice       = 100
	WeightSharedContact    = 50
	WeightCommunicatesWith = 50
)

type Link struct {
	AccountA   string `json:"accountA"`
	AccountB   string `json:"accountB"`
	LinkType   string `json:"linkType"`
	Weight     int    `json:"weight"`
	DetectedAt string `json:"detectedAt"`
}
