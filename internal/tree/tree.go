// Package tree 把已解析的容器渲染为可阅读的树状文本。
package tree

import (
	"fmt"
	"strings"

	"binary-parser/internal/format"
)

// Render 渲染容器结构；每条记录标注 CRC 校验状态。
func Render(c *format.Container) string {
	var b strings.Builder
	fmt.Fprintf(&b, "container (version=%d, records=%d)\n", c.Header.Version, c.Header.Count)
	for i, rec := range c.Records {
		state := "ok"
		if !rec.ChecksumOK() {
			state = "BAD"
		}
		fmt.Fprintf(&b, "  [%d] type=%d id=%d len=%d crc=%s\n", i, rec.Type, rec.ID, len(rec.Payload), state)
	}
	return b.String()
}
