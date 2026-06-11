package graph

type Stats struct {
	TotalAccounts int `json:"total_accounts"`
	TotalDevices  int `json:"total_devices"`
	TotalLinks    int `json:"total_links"`
	Hourly        int `json:"hourly"`
	UniqueDevices int `json:"unique_devices"`
}

type DeviceClone struct {
	DeviceHash string `json:"device_hash"`
	Count      int    `json:"count"`
	LastSeen   string `json:"last_seen"`
}

type ActivityPoint struct {
	Hour            string `json:"hour"`
	Total           int    `json:"total"`
	Links           int    `json:"links"`
	UniqueAccounts  int    `json:"unique_accounts"`
}

type DeviceRegistryItem struct {
	DeviceHash   string `json:"device_hash"`
	AccountCount int    `json:"account_count"`
	RecordCount  int    `json:"record_count"`
	LastSeen     string `json:"last_seen"`
	RiskLevel    string `json:"risk_level"`
}

type DeviceInfo struct {
	DeviceHash  string `json:"device_hash"`
	FirstSeenAt string `json:"first_seen_at"`
	LastSeenAt  string `json:"last_seen_at"`
}

type AccountProfile struct {
	UserID            string       `json:"user_id"`
	FirstSeenAt       string       `json:"first_seen_at"`
	LastSeenAt        string       `json:"last_seen_at"`
	Devices           []DeviceInfo `json:"devices"`
	PrimaryDeviceHash string       `json:"primary_device_hash"`
}

type RelatedAccount struct {
	UserID      string `json:"user_id"`
	DeviceHash  string `json:"device_hash"`
	MatchReason string `json:"match_reason"`
	LinkType    string `json:"link_type"`
	Weight      int    `json:"weight"`
}

type RiskVerdict struct {
	Level  string `json:"level"`
	Reason string `json:"reason"`
}

type RelatedResult struct {
	Related     []RelatedAccount `json:"related"`
	RiskVerdict RiskVerdict      `json:"risk_verdict"`
}

type DiffResult struct {
	AccountA AccountProfile `json:"account_a"`
	AccountB AccountProfile `json:"account_b"`
	Shared   DiffShared     `json:"shared"`
	Links    []Link         `json:"links"`
	Verdict  RiskVerdict    `json:"verdict"`
}

type DiffShared struct {
	Devices []string `json:"devices"`
}

type DeviceLookupResult struct {
	DeviceHash string           `json:"device_hash"`
	MatchType  string           `json:"match_type"`
	Accounts   []AccountProfile `json:"accounts"`
}
