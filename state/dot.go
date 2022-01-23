package state

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/jakub-m/formaggo/log"
)

func (g StateGraph) ExportToDotFile(path string) error {
	log.Debugf("Exporting state graph of size %d to file %s\n", g.NumStates(), path)
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	g.ExportToDot(f)
	return nil
}

func (g StateGraph) ExportToDot(w io.Writer) error {
	_, err := io.WriteString(w, "digraph D {\n")
	if err != nil {
		return err
	}

	for h, s := range g.hashToState {
		jb, err := json.MarshalIndent(s, "", " ")
		if err != nil {
			return err
		}
		label := strconv.Quote(string(jb))
		_, err = io.WriteString(w, fmt.Sprintf("s%d [label=%s]\n", h, label))
		if err != nil {
			return err
		}
	}

	for h, tt := range g.hashGraph {
		for _, t := range tt {
			_, err = io.WriteString(w, fmt.Sprintf("s%d -> s%d\n", h, t))
			if err != nil {
				return err
			}
		}
	}

	_, err = io.WriteString(w, "}")
	if err != nil {
		return err
	}
	return nil
}
