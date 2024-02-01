package nginx

type NginxStats struct {
	ConnectionsActive float64
	Connections       []Connections
}

type Connections struct {
	Type  string
	Total float64
}
