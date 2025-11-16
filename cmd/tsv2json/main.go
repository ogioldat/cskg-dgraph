package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
)

func main() {
	r := bufio.NewReaderSize(os.Stdin, 64*1024)
	w := bufio.NewWriterSize(os.Stdout, 64*1024)

	// fmt.Fprintf(w, "[\n")

	defer w.Flush()
	// defer fmt.Fprintf(w, "]\n")
	var lineNum int

	for {
		lineNum++

		if lineNum == 1 {
			continue // skip header
		}

		line, err := r.ReadBytes('\n')
		if len(line) <= 0 {
			if err != nil {
				break
			}
			continue
		}

		line = bytes.TrimRight(line, "\r\n")
		fields := bytes.Split(line, []byte{'\t'})
		if len(fields) < 10 {
			continue
		}

		id := fields[0]
		node1 := fields[1]
		relation := fields[2]
		node2 := fields[3]
		node1Label := fields[4]
		node2Label := fields[5]
		relationLabel := fields[6]
		source := fields[8]
		sentence := fields[9]

		// Concept node1
		fmt.Fprintf(w,
			`{"uid":"%s","dgraph.type":"Concept","uri":"%s","label":"%s"},`+"\n",
			escapeStr(node1), escapeStr(node1), escapeStr(node1Label),
		)

		// Concept node2
		fmt.Fprintf(w,
			`{"uid":"%s","dgraph.type":"Concept","uri":"%s","label":"%s"},`+"\n",
			escapeStr(node2), escapeStr(node2), escapeStr(node2Label),
		)

		// Relation edge (facets)
		fmt.Fprintf(w,
			`{"uid":"%s","rel":[{"uid":"%s","rel|edge_id":"%s","rel|relation":"%s","rel|relation_label":"%s","rel|source":"%s","rel|sentence":"%s"}]},`+"\n",
			escapeStr(node1),
			escapeStr(node2),
			escapeStr(id),
			escapeStr(relation),
			escapeStr(relationLabel),
			escapeStr(source),
			escapeStr(sentence),
		)
	}
}

func escapeStr(b []byte) string {
	out := make([]byte, 0, len(b))
	for _, c := range b {
		switch c {
		case '\\':
			out = append(out, '\\', '\\')
		case '"':
			out = append(out, '\\', '"')
		default:
			out = append(out, c)
		}
	}
	return string(out)
}
