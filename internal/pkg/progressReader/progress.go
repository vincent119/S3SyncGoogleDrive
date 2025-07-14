package progressReader

import (
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
	"io"
)

type ProgressManager struct {
	p *mpb.Progress
}

func NewProgressManager() *ProgressManager {
	return &ProgressManager{
		p: mpb.New(mpb.WithWidth(64)),
	}
}

func (pm *ProgressManager) Wait() {
	pm.p.Wait()
}

func (pm *ProgressManager) NewBar(total int64, fileName string) *mpb.Bar {
	return pm.p.AddBar(total,
		mpb.PrependDecorators(
			decor.Name("[Uploading] "+fileName+" "),
			decor.CountersKibiByte("% .2f / % .2f"),
		),
		mpb.AppendDecorators(decor.Percentage()),
	)
}

func (pm *ProgressManager) WrapReader(bar *mpb.Bar, r io.Reader) io.ReadCloser {
	return bar.ProxyReader(r)
}
