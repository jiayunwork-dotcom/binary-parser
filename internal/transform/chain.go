package transform

import "binary-parser/internal/format"

type Step func(c *format.Container) *format.Container

type Chain struct {
	steps []Step
}

func NewChain() *Chain {
	return &Chain{}
}

func (ch *Chain) Then(step Step) *Chain {
	ch.steps = append(ch.steps, step)
	return ch
}

func (ch *Chain) ThenSort(less SortFunc) *Chain {
	return ch.Then(func(c *format.Container) *format.Container {
		return SortBy(c, less)
	})
}

func (ch *Chain) ThenDedup() *Chain {
	return ch.Then(func(c *format.Container) *format.Container {
		return Dedup(c)
	})
}

func (ch *Chain) ThenTake(n int) *Chain {
	return ch.Then(func(c *format.Container) *format.Container {
		return Take(c, n)
	})
}

func (ch *Chain) ThenSkip(n int) *Chain {
	return ch.Then(func(c *format.Container) *format.Container {
		return Skip(c, n)
	})
}

func (ch *Chain) ThenReverse() *Chain {
	return ch.Then(func(c *format.Container) *format.Container {
		return Reverse(c)
	})
}

func (ch *Chain) ThenMap(fn MapFunc) *Chain {
	return ch.Then(func(c *format.Container) *format.Container {
		return Map(c, fn)
	})
}

func (ch *Chain) ThenReID(startID uint32) *Chain {
	return ch.Then(func(c *format.Container) *format.Container {
		return ReID(c, startID)
	})
}

func (ch *Chain) Apply(c *format.Container) *format.Container {
	for _, step := range ch.steps {
		if c == nil {
			return nil
		}
		c = step(c)
	}
	return c
}

func (ch *Chain) Len() int {
	return len(ch.steps)
}
