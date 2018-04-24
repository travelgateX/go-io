package elastic

import (
	"bytes"
	"encoding/json"
	"time"

	uuid "github.com/satori/go.uuid"

	"github.com/travelgateX/go-io/log"
)

const indexLayout = "2006.01.02"
const timestampLayout = time.RFC3339

type BulkFormatter struct {
	Index string
	Type  string
}

func formatFields(fields log.Fields) string {
	if len(fields) == 0 {
		return ""
	}
	b, err := json.Marshal(fields)
	if err != nil {
		return ""
	}
	if len(b) < 2 {
		return ""
	}

	return "," + string(b[1:len(b)-1]) // remove first and last '{' '}'
}

func (f *BulkFormatter) Format(b *bytes.Buffer, m string, lvl log.Level, fields log.Fields) {
	t := time.Now().UTC()

	u, err := uuid.NewV4()
	var id string
	if err == nil {
		// elastic creates an ID if the log doesn't contains one
		id = `, "_id":"` + u.String() + `"`
	}

	// format dates for each log, the index must contain a date at the end and a timestamps also required
	const timeBufferSize = len(timestampLayout) + 10 // timestampLayout is bigger than indexLayout
	var buf [timeBufferSize]byte

	timebuf := buf[:0]
	day := t.AppendFormat(timebuf, indexLayout)
	b.WriteString(`{"index":{"_index":"` + f.Index + string(day) + `","_type":"` + f.Type + `"` + id + `}}`)
	b.WriteByte('\n')

	// buf has been copied to b, its reusable
	timebuf = buf[:0]
	ts := t.AppendFormat(timebuf, timestampLayout)
	b.WriteString(`{"@Timestamp":"` + string(ts) + `","level":"` + lvl.String() + `","message":"` + m + `"` + formatFields(fields) + `}`)
	b.WriteByte('\n')
}
