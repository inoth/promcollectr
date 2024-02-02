package nginx

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type NginxStats struct {
	Interval float64

	ConnectionsActive float64
	Connections       []Connections
}

type Connections struct {
	Type  string
	Total float64
}

func ScanBasicStats(r io.Reader) ([]NginxStats, error) {
	s := bufio.NewScanner(r)

	var stats []NginxStats
	var conns []Connections
	var nginxStats NginxStats

	for s.Scan() {
		fields := strings.Fields(string(s.Bytes()))

		if len(fields) == 3 && fields[0] == "Active" {
			c, err := strconv.ParseFloat(fields[2], 64)
			if err != nil {
				return nil, fmt.Errorf("%w: strconv.ParseFloat failed", err)
			}
			nginxStats.ConnectionsActive = c
		}

		if fields[0] == "Reading:" {

			reading, _ := strconv.ParseFloat(fields[1], 64)
			writing, _ := strconv.ParseFloat(fields[3], 64)
			waiting, _ := strconv.ParseFloat(fields[5], 64)

			readingConns := Connections{Type: "reading", Total: reading}
			writingConns := Connections{Type: "writing", Total: writing}
			waitingConns := Connections{Type: "waiting", Total: waiting}

			conns = append(conns, readingConns, writingConns, waitingConns)
			nginxStats.Connections = conns
		}
	}

	stats = append(stats, nginxStats)

	if err := s.Err(); err != nil {
		return nil, fmt.Errorf("%w: failed to read metrics", err)
	}
	return stats, nil
}
