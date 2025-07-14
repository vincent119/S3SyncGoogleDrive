package drive

import (
	"fmt"
	"io"
	"strings"
)

type ProgressReader struct {
	io.Reader
	Total     int64
	ReadBytes int64
	FileName  string
	LastPrint int64
}

func (pr *ProgressReader) Read(p []byte) (int, error) {
	n, err := pr.Reader.Read(p)
	if n > 0 {
		pr.ReadBytes += int64(n)
		pr.printProgress()
	}
	return n, err
}

func (pr *ProgressReader) printProgress() {
	percent := float64(pr.ReadBytes) / float64(pr.Total) * 100
	if int(percent)-int(pr.LastPrint) >= 1 || percent == 100 { // 每1%刷新一次
		pr.LastPrint = int64(percent)
		bar := renderProgressBar(percent)
		fmt.Printf("\r[Uploading] %s %s %.0f%%", pr.FileName, bar, percent)
		if percent == 100 {
			fmt.Println() // 換行
		}
	}
}

func renderProgressBar(percent float64) string {
	totalBars := 20
	filled := int(percent / 100 * float64(totalBars))
	bar := "[" + strings.Repeat("#", filled) + strings.Repeat("-", totalBars-filled) + "]"
	return bar
}
