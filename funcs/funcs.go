package funcs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

var rq int64 = 0
var dumpCurl bool = false

// DumpCurl Dumps the request received in a CURL statement. In files named /tmp/here-x.req --
// It needs the DUMP_AS_CURL variable with the value "yes" in environment. It makes things
// slower.
func DumpCurl(r *http.Request, bodyBytes []byte) {
	if !dumpCurl {
		return
	}

	rq = rq + 1

	s := fmt.Sprintf("curl -X %s ${NGSILD_ADDRESS}%s \\\n", r.Method, r.URL.Path)
	for key, value := range r.Header {
		s = fmt.Sprintf("%s -H \"%s: %s\" \\\n", s, key, value[0])
	}
	s = fmt.Sprintf("%s-d '%s'", s, string(bodyBytes))

	log.Printf("\n%s\n", s)

	tmpFile, err := os.OpenFile(fmt.Sprintf("/tmp/here-%d.req", rq), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return
	}
	defer tmpFile.Close()

	// Write the byte array to the file
	if _, err := tmpFile.WriteString(s); err != nil {
		return
	}
}

// Function takes a json and compacts it removing all undesired extra characters, saving some space in
// the databases.
func CompactJson(bodyBytes []uint8) ([]byte, error) {
	var buf bytes.Buffer
	err := json.Compact(&buf, bodyBytes)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func Config() {
	dumpCurl = strings.ToLower(os.Getenv("DUMP_AS_CURL")) == "yes"
}
