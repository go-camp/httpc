package response

import (
	"net/http"
	"testing"
	"time"

	"github.com/go-camp/httpc"
)

func TestDateDeserializer(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	deserialize := DateDeserializer{}.Deserializer(
		func(req *http.Request) (output httpc.DeserializeOutput, md httpc.Metadata, err error) {
			output.Response = &http.Response{
				Header: http.Header{
					"Date": []string{
						now.Format(http.TimeFormat),
					},
				},
			}
			return output, md, nil
		},
	)
	_, md, err := deserialize(newNopHTTPRequest())
	if err != nil {
		t.Fatalf("expect no err, got %s", err)
	}
	date := GetDate(md)
	if !date.Equal(now) {
		t.Fatalf("expect date is %s, got %s", now, date)
	}
}
