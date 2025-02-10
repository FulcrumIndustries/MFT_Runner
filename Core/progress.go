package Core

import (
	"fmt"
	"strings"
	"time"
)

type ProgressBar struct {
	total     int
	current   int
	startTime time.Time
	width     int
}

func NewProgressBar(total int) *ProgressBar {
	if total < 1 {
		total = 1 // Prevent division by zero
	}
	return &ProgressBar{
		total:     total,
		startTime: time.Now(),
		width:     50,
	}
}

func (p *ProgressBar) Update(current int) {
	if current < 0 {
		current = 0
	} else if current > p.total {
		current = p.total
	}
	p.current = current
	p.draw()
}

func (p *ProgressBar) draw() {
	percent := float64(p.current) / float64(p.total)
	bar := strings.Repeat("=", int(percent*float64(p.width)))
	spaces := strings.Repeat(" ", p.width-len(bar))
	elapsed := time.Since(p.startTime).Round(time.Second)
	
	fmt.Printf("\r[%s%s] %d/%d (%.1f%%) %v elapsed", 
		bar, spaces, 
		p.current, p.total, 
		percent*100, 
		elapsed)
}

func (p *ProgressBar) Finish() {
	fmt.Println() // New line after completion
} 