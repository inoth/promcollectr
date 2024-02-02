package nginx

type NginxStats struct {
	Interval float64

	ConnectionsActive float64
	Connections       []Connections
}

type Connections struct {
	Type  string
	Total float64
}
