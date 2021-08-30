package response

import (
	"net/http"
	"time"

	"github.com/go-camp/httpc"
)

type mdDateKey struct{}

func GetDate(md httpc.Metadata) (t time.Time) {
	v := md.Get(mdDateKey{})
	t, _ = v.(time.Time)
	return t
}

// DateDeserializer parses the value of reponse header Date and
// sets the parsed date to metadata.
type DateDeserializer struct {
	ParseTime func(string) (time.Time, error)
}

func (d DateDeserializer) Deserializer(deserialize httpc.DeserializeFunc) httpc.DeserializeFunc {
	return func(req *http.Request) (httpc.DeserializeOutput, httpc.Metadata, error) {
		return d.deserialize(req, deserialize)
	}
}

func (d DateDeserializer) parseTime() func(string) (time.Time, error) {
	if d.ParseTime == nil {
		return http.ParseTime
	}
	return d.ParseTime
}

func (d DateDeserializer) deserialize(req *http.Request, deserialize httpc.DeserializeFunc) (
	output httpc.DeserializeOutput, md httpc.Metadata, err error,
) {
	output, md, err = deserialize(req)

	resp := output.Response
	if resp == nil {
		return
	}

	dh := resp.Header.Get(headerDate)
	if dh == "" {
		return
	}

	date, parseErr := d.parseTime()(dh)
	if parseErr != nil {
		return
	}
	md.Set(mdDateKey{}, date)

	return
}
