package graph

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"
)

func (r *Repository) GetStats(ctx context.Context) (Stats, error) {
	var stats Stats
	err := r.pool.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*) FROM account_nodes),
			(SELECT COUNT(*) FROM (
				SELECT device_hash
				FROM user_devices
				GROUP BY device_hash
				HAVING COUNT(DISTINCT user_id) > 1
			) clone_devices),
			(SELECT COUNT(*) FROM account_links WHERE link_type <> 'same_ip'),
			(SELECT COUNT(*) FROM account_links
				WHERE link_type <> 'same_ip'
				  AND detected_at >= now() - interval '1 hour'),
			(SELECT COUNT(*) FROM (
				SELECT device_hash
				FROM user_devices
				GROUP BY device_hash
				HAVING COUNT(DISTINCT user_id) > 1
			) clone_devices)
	`).Scan(&stats.TotalAccounts, &stats.TotalDevices, &stats.TotalLinks, &stats.Hourly, &stats.UniqueDevices)
	if err != nil {
		return Stats{}, fmt.Errorf("get stats: %w", err)
	}
	return stats, nil
}

func (r *Repository) TopDeviceClones(ctx context.Context, limit int) ([]DeviceClone, error) {
	if limit <= 0 {
		limit = 10
	}

	rows, err := r.pool.Query(ctx, `
		SELECT device_hash, user_id::text, last_seen_at
		FROM user_devices
		WHERE device_hash IN (
			SELECT device_hash
			FROM user_devices
			GROUP BY device_hash
			HAVING COUNT(DISTINCT user_id) > 1
		)
		ORDER BY last_seen_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("top device clones: %w", err)
	}
	defer rows.Close()

	type deviceMeta struct {
		hash     string
		lastSeen time.Time
		users    map[string]struct{}
	}

	devices := make(map[string]*deviceMeta)
	userToDevices := make(map[string]map[string]struct{})

	for rows.Next() {
		var hash, userID string
		var lastSeen time.Time
		if err := rows.Scan(&hash, &userID, &lastSeen); err != nil {
			return nil, fmt.Errorf("scan device clone row: %w", err)
		}

		meta, ok := devices[hash]
		if !ok {
			meta = &deviceMeta{hash: hash, users: make(map[string]struct{})}
			devices[hash] = meta
		}
		meta.users[userID] = struct{}{}
		if lastSeen.After(meta.lastSeen) {
			meta.lastSeen = lastSeen
		}

		if userToDevices[userID] == nil {
			userToDevices[userID] = make(map[string]struct{})
		}
		userToDevices[userID][hash] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate device clone rows: %w", err)
	}

	parent := make(map[string]string, len(devices))
	var find func(string) string
	find = func(hash string) string {
		if parent[hash] == "" {
			parent[hash] = hash
		}
		if parent[hash] != hash {
			parent[hash] = find(parent[hash])
		}
		return parent[hash]
	}
	union := func(a, b string) {
		rootA := find(a)
		rootB := find(b)
		if rootA != rootB {
			parent[rootB] = rootA
		}
	}

	for _, linkedDevices := range userToDevices {
		var hashes []string
		for hash := range linkedDevices {
			hashes = append(hashes, hash)
		}
		for i := 1; i < len(hashes); i++ {
			union(hashes[0], hashes[i])
		}
	}

	clusters := make(map[string]*DeviceClone)
	clusterUsers := make(map[string]map[string]struct{})
	for hash, meta := range devices {
		root := find(hash)
		if clusterUsers[root] == nil {
			clusterUsers[root] = make(map[string]struct{})
		}
		cluster, ok := clusters[root]
		if !ok {
			cluster = &DeviceClone{DeviceHash: hash}
			clusters[root] = cluster
		}

		if meta.lastSeen.After(parseRFC3339(cluster.LastSeen)) {
			cluster.DeviceHash = hash
			cluster.LastSeen = meta.lastSeen.UTC().Format(time.RFC3339)
		}

		for userID := range meta.users {
			clusterUsers[root][userID] = struct{}{}
		}
	}

	var clones []DeviceClone
	for root, cluster := range clusters {
		cluster.Count = len(clusterUsers[root])
		clones = append(clones, *cluster)
	}

	sort.Slice(clones, func(i, j int) bool {
		if clones[i].Count != clones[j].Count {
			return clones[i].Count > clones[j].Count
		}
		return clones[i].LastSeen > clones[j].LastSeen
	})

	if len(clones) > limit {
		clones = clones[:limit]
	}
	return clones, nil
}

func parseRFC3339(value string) time.Time {
	if value == "" {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}
	}
	return t
}

func (r *Repository) ActivitySeries(ctx context.Context, hours int) ([]ActivityPoint, error) {
	if hours <= 0 {
		hours = 24
	}
	rows, err := r.pool.Query(ctx, `
		WITH hours AS (
			SELECT generate_series(
				date_trunc('hour', now() - make_interval(hours => $1)),
				date_trunc('hour', now()),
				interval '1 hour'
			) AS hour
		)
		SELECT
			h.hour,
			COALESCE((SELECT COUNT(*) FROM ingested_events e WHERE date_trunc('hour', e.received_at) = h.hour), 0),
			COALESCE((SELECT COUNT(*) FROM account_links l
				WHERE link_type <> 'same_ip'
				  AND date_trunc('hour', l.detected_at) = h.hour), 0),
			COALESCE((SELECT COUNT(DISTINCT user_id) FROM user_devices WHERE date_trunc('hour', last_seen_at) = h.hour), 0)
		FROM hours h
		ORDER BY h.hour
	`, hours)
	if err != nil {
		return nil, fmt.Errorf("activity series: %w", err)
	}
	defer rows.Close()

	var series []ActivityPoint
	for rows.Next() {
		var point ActivityPoint
		var hour time.Time
		if err := rows.Scan(&hour, &point.Total, &point.Links, &point.UniqueAccounts); err != nil {
			return nil, fmt.Errorf("scan activity point: %w", err)
		}
		point.Hour = hour.UTC().Format(time.RFC3339)
		series = append(series, point)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate activity series: %w", err)
	}
	return series, nil
}

func deviceRiskLevel(accountCount int) string {
	switch {
	case accountCount > 1:
		return "clone"
	case accountCount == 1:
		return "clean"
	default:
		return "clean"
	}
}

func (r *Repository) ListDeviceRegistry(ctx context.Context, filter string, limit int) ([]DeviceRegistryItem, error) {
	if limit <= 0 {
		limit = 100
	}

	having := ""
	switch filter {
	case "clone":
		having = "HAVING COUNT(DISTINCT user_id) > 1"
	case "suspicious":
		having = "HAVING COUNT(DISTINCT user_id) = 1 AND COUNT(*) > 2"
	case "clean":
		having = "HAVING COUNT(DISTINCT user_id) = 1 AND COUNT(*) <= 2"
	case "all":
		having = ""
	default:
		having = "HAVING COUNT(DISTINCT user_id) > 1"
		filter = "clone"
	}

	query := fmt.Sprintf(`
		SELECT device_hash,
			COUNT(DISTINCT user_id) AS account_count,
			COUNT(*) AS record_count,
			MAX(last_seen_at) AS last_seen
		FROM user_devices
		GROUP BY device_hash
		%s
		ORDER BY account_count DESC, last_seen DESC
		LIMIT $1
	`, having)

	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("list device registry: %w", err)
	}
	defer rows.Close()

	var items []DeviceRegistryItem
	for rows.Next() {
		var item DeviceRegistryItem
		var lastSeen time.Time
		if err := rows.Scan(&item.DeviceHash, &item.AccountCount, &item.RecordCount, &lastSeen); err != nil {
			return nil, fmt.Errorf("scan device registry: %w", err)
		}
		item.LastSeen = lastSeen.UTC().Format(time.RFC3339)
		item.RiskLevel = deviceRiskLevel(item.AccountCount)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate device registry: %w", err)
	}
	return items, nil
}

func (r *Repository) GetAccount(ctx context.Context, userID string) (*AccountProfile, error) {
	var profile AccountProfile
	var firstSeen, lastSeen time.Time
	err := r.pool.QueryRow(ctx, `
		SELECT user_id::text, first_seen_at, last_seen_at
		FROM account_nodes
		WHERE user_id = $1::uuid
	`, userID).Scan(&profile.UserID, &firstSeen, &lastSeen)
	if err != nil {
		return nil, fmt.Errorf("get account: %w", err)
	}
	profile.FirstSeenAt = firstSeen.UTC().Format(time.RFC3339)
	profile.LastSeenAt = lastSeen.UTC().Format(time.RFC3339)

	devices, err := r.listAccountDevices(ctx, userID)
	if err != nil {
		return nil, err
	}
	profile.Devices = devices
	if len(devices) > 0 {
		profile.PrimaryDeviceHash = devices[0].DeviceHash
	}

	return &profile, nil
}

func (r *Repository) listAccountDevices(ctx context.Context, userID string) ([]DeviceInfo, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT device_hash, first_seen_at, last_seen_at
		FROM user_devices
		WHERE user_id = $1::uuid
		ORDER BY last_seen_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list account devices: %w", err)
	}
	defer rows.Close()

	var devices []DeviceInfo
	for rows.Next() {
		var d DeviceInfo
		var firstSeen, lastSeen time.Time
		if err := rows.Scan(&d.DeviceHash, &firstSeen, &lastSeen); err != nil {
			return nil, fmt.Errorf("scan device: %w", err)
		}
		d.FirstSeenAt = firstSeen.UTC().Format(time.RFC3339)
		d.LastSeenAt = lastSeen.UTC().Format(time.RFC3339)
		devices = append(devices, d)
	}
	return devices, rows.Err()
}

func (r *Repository) FindAccountByDevice(ctx context.Context, deviceHash string) (*AccountProfile, error) {
	result, err := r.FindAccountsByDeviceQuery(ctx, deviceHash)
	if err != nil {
		return nil, err
	}
	if len(result.Accounts) == 0 {
		return nil, fmt.Errorf("find account by device: not found")
	}
	return &result.Accounts[0], nil
}

func normalizeDeviceQuery(query string) string {
	query = strings.TrimSpace(strings.ToLower(query))
	query = strings.ReplaceAll(query, " ", "")
	if idx := strings.Index(query, "..."); idx >= 0 {
		query = query[:idx]
	}
	if idx := strings.Index(query, "…"); idx >= 0 {
		query = query[:idx]
	}
	return strings.TrimSpace(query)
}

func (r *Repository) FindAccountsByDeviceQuery(ctx context.Context, query string) (*DeviceLookupResult, error) {
	query = normalizeDeviceQuery(query)
	if query == "" {
		return nil, fmt.Errorf("empty device query")
	}

	hashes, matchType, err := r.matchDeviceHashes(ctx, query)
	if err != nil {
		return nil, err
	}

	userIDs, err := r.listUsersForDeviceHashes(ctx, hashes)
	if err != nil {
		return nil, err
	}

	var accounts []AccountProfile
	for _, userID := range userIDs {
		profile, err := r.GetAccount(ctx, userID)
		if err != nil {
			continue
		}
		accounts = append(accounts, *profile)
	}
	if len(accounts) == 0 {
		return nil, fmt.Errorf("find accounts by device: not found")
	}

	return &DeviceLookupResult{
		DeviceHash: hashes[0],
		MatchType:  matchType,
		Accounts:   accounts,
	}, nil
}

func (r *Repository) matchDeviceHashes(ctx context.Context, query string) ([]string, string, error) {
	hashes, err := r.listDeviceHashesExact(ctx, query)
	if err != nil {
		return nil, "", err
	}
	if len(hashes) > 0 {
		return hashes, "exact", nil
	}

	hashes, err = r.listDeviceHashesPrefix(ctx, query)
	if err != nil {
		return nil, "", err
	}
	if len(hashes) > 0 {
		return hashes, "prefix", nil
	}

	return nil, "", fmt.Errorf("device hash not found")
}

func (r *Repository) listDeviceHashesExact(ctx context.Context, query string) ([]string, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT DISTINCT device_hash
		FROM user_devices
		WHERE LOWER(device_hash) = $1
	`, query)
	if err != nil {
		return nil, fmt.Errorf("list exact device hashes: %w", err)
	}
	defer rows.Close()
	return scanStrings(rows)
}

func (r *Repository) listDeviceHashesPrefix(ctx context.Context, query string) ([]string, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT DISTINCT device_hash
		FROM user_devices
		WHERE LOWER(device_hash) LIKE $1
		ORDER BY device_hash
	`, query+"%")
	if err != nil {
		return nil, fmt.Errorf("list prefix device hashes: %w", err)
	}
	defer rows.Close()
	return scanStrings(rows)
}

func (r *Repository) listUsersForDeviceHashes(ctx context.Context, hashes []string) ([]string, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT DISTINCT user_id::text
		FROM user_devices
		WHERE device_hash = ANY($1)
		ORDER BY user_id::text
	`, hashes)
	if err != nil {
		return nil, fmt.Errorf("list users for device hashes: %w", err)
	}
	defer rows.Close()
	return scanStrings(rows)
}

func scanStrings(rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
}) ([]string, error) {
	var values []string
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			return nil, fmt.Errorf("scan string: %w", err)
		}
		values = append(values, value)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return values, nil
}

func (r *Repository) sharedDeviceHash(ctx context.Context, userA, userB string) string {
	var hash string
	err := r.pool.QueryRow(ctx, `
		SELECT ud1.device_hash
		FROM user_devices ud1
		INNER JOIN user_devices ud2 ON ud1.device_hash = ud2.device_hash
		WHERE ud1.user_id = $1::uuid AND ud2.user_id = $2::uuid
		ORDER BY GREATEST(ud1.last_seen_at, ud2.last_seen_at) DESC
		LIMIT 1
	`, userA, userB).Scan(&hash)
	if err != nil {
		return ""
	}
	return hash
}

func (r *Repository) GetRelatedAccounts(ctx context.Context, userID string) (RelatedResult, error) {
	links, err := r.ListLinksForAccount(ctx, userID)
	if err != nil {
		return RelatedResult{}, err
	}

	seen := make(map[string]bool)
	var related []RelatedAccount

	for _, link := range links {
		other := link.AccountB
		if other == userID {
			other = link.AccountA
		}
		if seen[other] {
			continue
		}
		seen[other] = true

		reason := linkReasonLabel(link.LinkType)
		deviceHash := ""
		if link.LinkType == LinkSameDevice {
			deviceHash = r.sharedDeviceHash(ctx, userID, other)
		}
		if deviceHash == "" {
			if profile, err := r.GetAccount(ctx, other); err == nil && profile.PrimaryDeviceHash != "" {
				deviceHash = profile.PrimaryDeviceHash
			}
		}

		related = append(related, RelatedAccount{
			UserID:      other,
			DeviceHash:  deviceHash,
			MatchReason: reason,
			LinkType:    link.LinkType,
			Weight:      link.Weight,
		})
	}

	verdict := computeVerdict(related)
	return RelatedResult{Related: related, RiskVerdict: verdict}, nil
}

func linkReasonLabel(linkType string) string {
	switch linkType {
	case LinkSameDevice:
		return "same_device"
	case LinkSharedContact:
		return "shared_contact"
	case LinkCommunicatesWith:
		return "communicates_with"
	default:
		return linkType
	}
}

func computeVerdict(related []RelatedAccount) RiskVerdict {
	for _, r := range related {
		if r.LinkType == LinkSameDevice {
			return RiskVerdict{Level: "clone", Reason: "shared device"}
		}
	}
	if len(related) > 0 {
		return RiskVerdict{Level: "suspicious", Reason: "linked accounts"}
	}
	return RiskVerdict{Level: "clean", Reason: "no links"}
}

func (r *Repository) ListLinksForAccount(ctx context.Context, userID string) ([]Link, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT account_a::text, account_b::text, link_type, weight, detected_at
		FROM account_links
		WHERE (account_a = $1::uuid OR account_b = $1::uuid)
		  AND link_type <> 'same_ip'
		ORDER BY weight DESC, detected_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list links for account: %w", err)
	}
	defer rows.Close()

	var links []Link
	for rows.Next() {
		var link Link
		var detectedAt time.Time
		if err := rows.Scan(&link.AccountA, &link.AccountB, &link.LinkType, &link.Weight, &detectedAt); err != nil {
			return nil, fmt.Errorf("scan link: %w", err)
		}
		link.DetectedAt = detectedAt.UTC().Format(time.RFC3339)
		links = append(links, link)
	}
	return links, rows.Err()
}

func (r *Repository) DiffAccounts(ctx context.Context, userA, userB string) (*DiffResult, error) {
	profileA, err := r.GetAccount(ctx, userA)
	if err != nil {
		return nil, err
	}
	profileB, err := r.GetAccount(ctx, userB)
	if err != nil {
		return nil, err
	}

	sharedDevices := intersectHashes(deviceHashes(profileA.Devices), deviceHashes(profileB.Devices))

	links, err := r.ListLinksForAccount(ctx, userA)
	if err != nil {
		return nil, err
	}
	var between []Link
	for _, link := range links {
		if (link.AccountA == userA && link.AccountB == userB) || (link.AccountA == userB && link.AccountB == userA) {
			between = append(between, link)
		}
	}

	verdict := RiskVerdict{Level: "clean", Reason: "different accounts"}
	if len(sharedDevices) > 0 {
		verdict = RiskVerdict{Level: "clone", Reason: "shared device"}
	} else if len(between) > 0 {
		verdict = RiskVerdict{Level: "suspicious", Reason: "graph link"}
	}

	return &DiffResult{
		AccountA: *profileA,
		AccountB: *profileB,
		Shared: DiffShared{
			Devices: sharedDevices,
		},
		Links:   between,
		Verdict: verdict,
	}, nil
}

func deviceHashes(devices []DeviceInfo) []string {
	out := make([]string, len(devices))
	for i, d := range devices {
		out[i] = d.DeviceHash
	}
	return out
}

func intersectHashes(a, b []string) []string {
	set := make(map[string]bool, len(a))
	for _, v := range a {
		set[v] = true
	}
	var out []string
	for _, v := range b {
		if set[v] {
			out = append(out, v)
		}
	}
	return out
}
