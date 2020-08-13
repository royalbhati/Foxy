package progress

import (
	"fmt"
)

type Bar struct {
	percent int64  // progress percentage
	cur     int64  // current progress
	total   int64  // total value for progress
	rate    string // the actual progress bar to be printed
	graph   string // the fill value for progress bar
}

func (bar *Bar) NewOption(start, total int64) {
	bar.cur = start
	bar.total = total
	if bar.graph == "" {
		bar.graph = "█"
	}
	bar.percent = bar.getPercent()
	for i := 0; i < int(bar.percent); i += 2 {
		bar.rate += bar.graph // initial progress position
	}
}
func (bar *Bar) getPercent() int64 {
	return int64(float32(bar.cur) / float32(bar.total) * 100)
}

func (bar *Bar) NewOptionWithGraph(start, total int64, graph string) {
	bar.graph = graph
	bar.NewOption(start, total)
}

func (bar *Bar) Play(cur int64) {
	bar.cur = cur
	last := bar.percent
	bar.percent = bar.getPercent()
	if bar.percent != last && bar.percent%2 == 0 {
		calc := int(100 / (bar.total * 2))
		for x := 0; x <= calc; x++ {
			bar.rate = bar.rate + bar.graph
		}
	}

	fmt.Printf("\r\u001b[32m[%-50s]%3d%% %8d/%d \033[5;36m", bar.rate, bar.percent, bar.cur, bar.total)
}

func (bar *Bar) Finish() {
	fmt.Println()
}
