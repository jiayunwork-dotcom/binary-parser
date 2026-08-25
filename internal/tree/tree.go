package tree

import (
	"fmt"
	"strings"

	"binary-parser/internal/format"
)

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
