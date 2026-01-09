package progressReader

import (
	"bytes"
	"io"
	"testing"
)

func TestNewProgressManager(t *testing.T) {
	// 測試初始化是否回傳 nil
	pm := NewProgressManager()
	if pm == nil {
		t.Error("NewProgressManager() 回傳了 nil")
	}

	// 驗證內部 mpb.Progress 是否已初始化
	if pm.p == nil {
		t.Error("NewProgressManager() 初始化的 mpb.Progress 為 nil")
	}
}

func TestWait(t *testing.T) {
	pm := NewProgressManager()
	if pm == nil {
		t.Fatal("NewProgressManager() returned nil")
	}
	// Wait should not panic (minimal test)
	pm.Wait()
}

func TestNewBar(t *testing.T) {
	pm := NewProgressManager()
	bar := pm.NewBar(100, "test-file")
	if bar == nil {
		t.Error("NewBar() returned nil")
	}
}

func TestWrapReader(t *testing.T) {
	pm := NewProgressManager()
	content := []byte("test content")
	reader := bytes.NewReader(content)

	// bar := pm.NewBar(int64(len(content)), "test.txt")
	// To avoid actual bar logic which might require mpb internal state or interfere with visualization,
	// we just test that WrapReader returns a reader that works.
	// However WrapReader calls bar.ProxyReader(r), so we need a bar.

	bar := pm.NewBar(int64(len(content)), "test.txt")
	wrapped := pm.WrapReader(bar, reader)

	buf := make([]byte, len(content))
	n, err := wrapped.Read(buf)
	if err != nil && err != io.EOF {
		t.Errorf("Read failed: %v", err)
	}
	if n != len(content) {
		t.Errorf("Read count mismatch: got %d, want %d", n, len(content))
	}
	if string(buf) != "test content" {
		t.Errorf("Content mismatch: got %s, want %s", string(buf), "test content")
	}
	wrapped.Close()
}


