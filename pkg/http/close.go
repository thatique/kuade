package http

import (
	"io"
	"io/ioutil"
)

// Drainbody close non nil response with any response Body
// convenient wrapper to drain any remaining data on response body
func DrainBody(respBody io.ReadCloser) {
	if respBody != nil {
		defer respBody.Close()
		io.Copy(ioutil.Discard, respBody)
	}
}
