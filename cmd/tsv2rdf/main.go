package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func main() {
	r := bufio.NewReaderSize(os.Stdin, 64*1024)

	var w *bufio.Writer
	var f *os.File
	var lineNum int
	var objCount int
	nodeCache := make(map[[32]byte]bool)
	relationCache := make(map[string]bool)

	name := "data/out/data.rdf"

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

	for {
		lineNum++

		line, err := r.ReadBytes('\n')

		if len(line) <= 0 {
			if err == io.EOF {
				break
			}
			continue
		}

		if err != nil && err != io.EOF {
			fmt.Printf("read error at line %d: %v\n", lineNum, err)
		}

		if lineNum == 1 {
			continue // skip header
		}

		line = bytes.TrimRight(line, "\r\n")
		fields := bytes.Split(line, []byte{'\t'})
		if len(fields) < 10 {
			fmt.Println("Wrong row")
			continue
		}

		edgeId := fields[0]
		node1 := fields[1]
		relation := fields[2]
		node2 := fields[3]
		node1Label := fields[4]
		node2Label := fields[5]
		relationLabel := fields[6]
		source := fields[8]
		sentence := fields[9]

		node1Hash := sha256.Sum256(node1)
		node2Hash := sha256.Sum256(node2)

		buf := bytes.NewBufferString("")

		_, hasNode1 := nodeCache[node1Hash]
		if !hasNode1 {
			fmt.Fprintf(buf, `<_:%s> <uri> "%s" .`, escapeStr(node1, true), escapeStr(node1, false))

			fmt.Fprintf(buf, "\n")
			fmt.Fprintf(buf, `<_:%s> <label> "%s" .`, escapeStr(node1, true), escapeStr(node1Label, false))

			fmt.Fprintf(buf, "\n")
			fmt.Fprintf(buf, `<_:%s> <dgraph.type> "Concept" .`, escapeStr(node1, true))
			// fmt.Fprintf(buf, "\n")
			// fmt.Fprintf(buf, `<_:%s> <dgraph.type> "%s" .`, escapeStr(node1, true), "Concept")

			fmt.Fprintf(buf, "\n")

			nodeCache[node1Hash] = true
		}

		_, hasNode2 := nodeCache[node2Hash]
		if !hasNode2 {
			fmt.Fprintf(buf, `<_:%s> <uri> "%s" .`, escapeStr(node2, true), escapeStr(node2, false))

			fmt.Fprintf(buf, "\n")
			fmt.Fprintf(buf, `<_:%s> <label> "%s" .`, escapeStr(node2, true), escapeStr(node2Label, false))

			fmt.Fprintf(buf, "\n")
			fmt.Fprintf(buf, `<_:%s> <dgraph.type> "Concept" .`, escapeStr(node2, true))
			// fmt.Fprintf(buf, "\n")
			// fmt.Fprintf(buf, `<_:%s> <dgraph.type> "%s" .`, escapeStr(node2, true), "Concept")

			fmt.Fprintf(buf, "\n")

			nodeCache[node2Hash] = true
		}

		if len(relationLabel) < 1 {
			continue
		}

		// Original relation labels
		// rdfRelation := []byte(
		// 	strings.ReplaceAll(
		// 		strings.ReplaceAll(string(relationLabel),
		// 			" ", "_"),
		// 		"|",
		// 		"_",
		// 	),
		// )
		// Synthetic label
		// rdfRelation := []byte("rel")

		fmt.Fprintf(buf, `<_:%s> <rel> <_:%s> (edge_id="%s", relation="%s", label="%s", source="%s", sentence="%s") .`,
			escapeStr(node1, true),
			escapeStr(node2, true),
			escapeStr(edgeId, false),
			escapeStr(relation, false),
			escapeStr(relationLabel, false),
			escapeStr(source, false),

			escapeStr(sentence, false),
		)
		// fmt.Fprintf(buf, "\n")

		// fmt.Fprintf(buf, `<_:%s> <rel_label> "%s" .`,
		// 	escapeStr(node1, true),
		// 	escapeStr(relationLabel, true),
		// )

		if objCount > 0 {
			if _, err := fmt.Fprintf(w, "\n"); err != nil {
				fmt.Fprintf(os.Stderr, "write newline error at line %d: %v\n", lineNum, err)
			}
		}

		if _, err := w.Write(buf.Bytes()); err != nil {
			fmt.Fprintf(os.Stderr, "write buffer error at line %d: %v\n", lineNum, err)
		}
		objCount++
	}

	if w == nil {
		return
	}
	if err := w.Flush(); err != nil {
		fmt.Fprintf(os.Stderr, "flush error: %v\n", err)
	}
	if err := f.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "file close error: %v\n", err)
	}
	w = nil

	fmt.Println("Completed successfully, lines", lineNum)

	if lineNum != 6001533 {
		fmt.Println("Expected lines", 6001533)
	}

	schemaRelationsDefBuf := bytes.NewBufferString("")
	schemaRelationsBuf := bytes.NewBufferString("")
	for k := range relationCache {
		fmt.Fprintf(schemaRelationsDefBuf, "%s: [uid] @reverse .\n", k)
		fmt.Fprintf(schemaRelationsBuf, "\t%s\n", k)
	}

	schema := bytes.NewBufferString("")

	fmt.Fprintf(schema, "uri: string @index(exact) .\n")
	fmt.Fprintf(schema, "label: string @index(fulltext, term, exact) .\n")
	fmt.Fprintf(schema, "%s\n", schemaRelationsDefBuf)

	fmt.Fprintf(schema, `type Concept {
    uri
    label`)
	fmt.Fprintf(schema, "\n%s", schemaRelationsBuf)
	fmt.Fprintf(schema, "\n}\n")

	os.WriteFile("data/schema.dql", schema.Bytes(), 0644)

}

func escapeStr(b []byte, predicate bool) string {
	out := make([]byte, 0, len(b))
	for _, c := range b {
		switch c {
		case '\\':
			if predicate {
				out = append(out, '%', '5', 'C')
			} else {
				out = append(out, '\\', '\\')
			}
		case '"':
			if predicate {
				continue
			} else {
				out = append(out, '\\', '"')
			}
		case '`':
			out = append(out, '\'')
		case '>':
			if predicate {
				out = append(out, '%', '3', 'E')
			}
		case '<':
			if predicate {
				out = append(out, '%', '3', 'C')
			}
		default:
			out = append(out, c)
		}
	}
	return string(out)
}
