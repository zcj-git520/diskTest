package diskTest

import "fmt"

// Progress 进图结构体
type Progress struct {
	name    string  // 名字
	percent int64  // 百分比
	current int64  // 当前进度
	total   int64  // 总量
	rate    string // 进度条
	graph   string // 进度符号
}

// NewProgress 初始化方法
func NewProgress(name string, start, total int64) *Progress {
	p := new(Progress)
	p.name = name
	p.current = start
	p.total = total
	p.graph = "###" // 这里设置进度条的样式
	p.percent = p.GetPercent()
	return p
}

// GetPercent 获取百分比
func (p *Progress) GetPercent() int64 {
	return int64(float32(p.current) / float32(p.total) * 100)
}

// Add 增加进度
func (p *Progress) Add(i int64) {

	p.current += i

	if p.current > p.total {
		return
	}

	last := p.percent
	p.percent = p.GetPercent()

	if p.percent != last && p.percent%2 == 0 {
		p.rate += p.graph
	}

	fmt.Printf("\r%s:[%-50s]%8d%% %8d/%d", p.name, p.rate, p.percent, p.current, p.total)
	// %-50s 左对齐, 占50个字符位置, 打印string
	// %3d   右对齐, 占3个字符位置 打印int

	if p.current == p.total {
		p.Done()
	}
}

// Done 完毕
func (p *Progress) Done() {
	fmt.Println()
}
