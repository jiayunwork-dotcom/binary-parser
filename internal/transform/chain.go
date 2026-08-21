package transform

import "binary-parser/internal/format"

// Step 是变换管道中的一步操作。
type Step func(c *format.Container) *format.Container

// Chain 将多个变换步骤串联为一个管道，按顺序依次执行。
type Chain struct {
	steps []Step
}

// NewChain 创建空的变换管道。
func NewChain() *Chain {
	return &Chain{}
}

// Then 追加一步变换。
func (ch *Chain) Then(step Step) *Chain {
	ch.steps = append(ch.steps, step)
	return ch
}

// ThenSort 追加排序步骤。
func (ch *Chain) ThenSort(less SortFunc) *Chain {
	return ch.Then(func(c *format.Container) *format.Container {
		return SortBy(c, less)
	})
}

// ThenDedup 追加去重步骤。
func (ch *Chain) ThenDedup() *Chain {
	return ch.Then(func(c *format.Container) *format.Container {
		return Dedup(c)
	})
}

// ThenTake 追加取前 n 条步骤。
func (ch *Chain) ThenTake(n int) *Chain {
	return ch.Then(func(c *format.Container) *format.Container {
		return Take(c, n)
	})
}

// ThenSkip 追加跳过前 n 条步骤。
func (ch *Chain) ThenSkip(n int) *Chain {
	return ch.Then(func(c *format.Container) *format.Container {
		return Skip(c, n)
	})
}

// ThenReverse 追加反转步骤。
func (ch *Chain) ThenReverse() *Chain {
	return ch.Then(func(c *format.Container) *format.Container {
		return Reverse(c)
	})
}

// ThenMap 追加映射步骤。
func (ch *Chain) ThenMap(fn MapFunc) *Chain {
	return ch.Then(func(c *format.Container) *format.Container {
		return Map(c, fn)
	})
}

// ThenReID 追加重编号步骤。
func (ch *Chain) ThenReID(startID uint32) *Chain {
	return ch.Then(func(c *format.Container) *format.Container {
		return ReID(c, startID)
	})
}

// Apply 对容器执行整个管道。
func (ch *Chain) Apply(c *format.Container) *format.Container {
	for _, step := range ch.steps {
		if c == nil {
			return nil
		}
		c = step(c)
	}
	return c
}

// Len 返回管道步骤数。
func (ch *Chain) Len() int {
	return len(ch.steps)
}
