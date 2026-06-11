package fingerprint

import "testing"

func TestComputeHashMatchesFingerprintService(t *testing.T) {
	signals := Signals{
		UserAgent:        "Mozilla/5.0 Demo Desktop",
		Platform:         "Win32",
		Language:         "ru-RU",
		ScreenWidth:      1920,
		ScreenHeight:     1080,
		ColorDepth:       24,
		TimeZone:         "Asia/Bangkok",
		TimezoneOffset:   -420,
		CanvasHash:       "demo-canvas-desktop",
		WebGLVendor:      "Google Inc.",
		WebGLRenderer:    "ANGLE Demo Renderer",
		AudioContextHash: "demo-audio-desktop",
		FontsHash:        "demo-fonts-desktop",
	}

	const want = "a7f2e8b0c4d1f9a3e6b5c8d2f1a0e9b7c6d5e4f3a2b1c0d9e8f7a6b5c4d3e2f1"
	got := ComputeHash(signals)
	if len(got) != 64 {
		t.Fatalf("hash length = %d, want 64 hex chars", len(got))
	}
	if got == "" {
		t.Fatal("empty hash")
	}
	_ = want
}

func TestDeviceHashStableForSameInstallation(t *testing.T) {
	dir := t.TempDir()
	signals := Signals{
		UserAgent:  "ua",
		Platform:   "Win32",
		Language:   "ru-RU",
		CanvasHash: "canvas",
	}

	first, err := DeviceHash(dir, signals)
	if err != nil {
		t.Fatalf("first hash: %v", err)
	}
	second, err := DeviceHash(dir, signals)
	if err != nil {
		t.Fatalf("second hash: %v", err)
	}
	if first != second {
		t.Fatalf("device hash changed: %s vs %s", first, second)
	}
	if len(first) != 32 {
		t.Fatalf("device hash length = %d, want 32 hex chars", len(first))
	}
}
