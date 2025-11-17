package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
)

const fileChunkSize = 10000

var dirBatchSize int

func init() {
	flag.IntVar(&dirBatchSize, "batch", 300, "directory batch size")
}

func main() {
	flag.Parse()
	r := bufio.NewReaderSize(os.Stdin, 64*1024)

	fileIndex := 1
	var w *bufio.Writer
	var f *os.File
	var lineNum int
	var objCount int

	openFile := func() {
		fileBatchNum := int(
			math.Ceil(float64(fileIndex) / float64(dirBatchSize)),
		)
		name := fmt.Sprintf(
			"data/out/%d/out_%05d.rdf",
			fileBatchNum,
			fileIndex,
		)

		// create parent directory
		dir := filepath.Dir(name)
		if err := os.MkdirAll(dir, 0755); err != nil {
			panic(err)
		}

		fh, err := os.Create(name)
		if err != nil {
			panic(err)
		}
		f = fh
		w = bufio.NewWriterSize(f, 64*1024)
		objCount = 0
	}

	closeFile := func() {
		if w == nil {
			return
		}
		w.Flush()
		f.Close()
		w = nil
	}

	openFile()

	for {
		lineNum++

		if lineNum >= 15 {
			break
		}

		line, err := r.ReadBytes('\n')
		if len(line) <= 0 {
			if err != nil {
				break
			}
			continue
		}

		if lineNum == 1 {
			continue // skip header
		}

		line = bytes.TrimRight(line, "\r\n")
		fields := bytes.Split(line, []byte{'\t'})
		if len(fields) < 10 {
			continue
		}

		edgeId := fields[0]
		node1 := fields[1]
		relation := fields[2]
		node2 := fields[3]
		node1Label := fields[4]
		node2Label := fields[5]
		relationLabel := string(fields[6])
		source := fields[8]
		sentence := fields[9]

		writeObj := func(buf *bytes.Buffer, args ...interface{}) {
			if objCount > 0 {
				fmt.Fprintf(w, "\n")
			}
			// fmt.Fprintf(w, format, args...)
			w.Write(buf.Bytes())
			objCount++
		}

		buf := bytes.NewBufferString("")

		fmt.Fprintf(buf, `<_:%s> <uri> "%s" .`, escapeStr(node1), escapeStr(node1))

		fmt.Fprintf(buf, "\n")
		fmt.Fprintf(buf, `<_:%s> <label> "%s" .`, escapeStr(node1), escapeStr(node1Label))

		fmt.Fprintf(buf, "\n")

		fmt.Fprintf(buf, `<_:%s> <uri> "%s" .`, escapeStr(node2), escapeStr(node2))

		fmt.Fprintf(buf, "\n")
		fmt.Fprintf(buf, `<_:%s> <label> "%s" .`, escapeStr(node2), escapeStr(node2Label))

		fmt.Fprintf(buf, "\n")

		rdfRelation := []byte(strings.Join(strings.Split(relationLabel, string(' ')), string('_')))

		fmt.Fprintf(buf, `<_:%s> <%s> <_:%s> (edge_id="%s", relation="%s", source="%s", sentence="%s") .`,
			escapeStr(node1),
			escapeStr(rdfRelation),
			escapeStr(node2),
			escapeStr(edgeId),
			escapeStr(relation),
			escapeStr(source),
			escapeStr(sentence),
		)

		fmt.Fprintf(buf, "\n")

		// node1
		writeObj(
			buf,
			// `<_:%s> <>  .`,
			// `{"uid":"_:%s","dgraph.type":"Concept","uri":"%s","label":"%s"}`,
			// escapeStr(node1),
			//  escapeStr(node1),
			//  escapeStr(node1Label),

		)

		// // node2
		// writeObj(
		// 	`{"uid":"_:%s","dgraph.type":"Concept","uri":"%s","label":"%s"}`,
		// 	escapeStr(node2),
		//  escapeStr(node2),
		//  escapeStr(node2Label),

		// )

		// // relation
		// writeObj(
		// 	`{"uid":"_:%s","rel":[{"uid":"_:%s","rel|edge_id":"%s","rel|relation":"%s","rel|relation_label":"%s","rel|source":"%s","rel|sentence":"%s"}]}`,
		// 	escapeStr(node1),

		// 	escapeStr(node2),

		// 	escapeStr(id),

		// 	escapeStr(relation),

		// 	escapeStr(relationLabel),

		// 	escapeStr(source),

		// 	escapeStr(sentence),

		// )

		if objCount >= fileChunkSize {
			closeFile()
			fileIndex++
			openFile()
		}
	}

	closeFile()
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
