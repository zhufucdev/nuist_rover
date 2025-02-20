package helper

import (
	"golang.org/x/text/encoding/ianaindex"
	"golang.org/x/text/transform"
	"io"
	"net/http"
)

func GetBody(response *http.Response) io.Reader {
	encoding, err := ianaindex.MIME.Encoding(GetCharset(response.Header.Get("Content-Type")))
	if err != nil {
		panic(err)
	}

	return transform.NewReader(response.Body, encoding.NewDecoder())
}
