package main

type Edge struct {
	u int
	v int
}

type CSRGraph struct {
	N         int // vertices
	M         int // undirected edges
	Deg       []int
	Offsets   []int
	Neighbors []int
}

type RunRecord struct {
	Type       string  `json:"type"` // "run"
	Scale      int     `json:"scale"`
	EdgeFactor int     `json:"edge_factor"`
	RunIndex   int     `json:"run_index"` // 0-based BFS iteration
	Root       int     `json:"root"`
	TimeS      float64 `json:"time_s"`
	Nedge      float64 `json:"nedge"`
	TEPS       float64 `json:"teps"`
}

type SummaryRecord struct {
	Type           string  `json:"type"` // "summary"
	Scale          int     `json:"scale"`
	EdgeFactor     int     `json:"edge_factor"`
	Nodes          int     `json:"nodes"`
	TotalEdges     int     `json:"total_edges"`
	NBFS           int     `json:"nbfs"`
	ConstructTimeS float64 `json:"construct_time_s"`

	TimeMin    float64 `json:"time_min_s"`
	TimeMedian float64 `json:"time_median_s"`
	TimeMax    float64 `json:"time_max_s"`
	TimeMean   float64 `json:"time_mean_s"`
	TimeStddev float64 `json:"time_stddev_s"`

	NedgeMin    float64 `json:"nedge_min"`
	NedgeMedian float64 `json:"nedge_median"`
	NedgeMax    float64 `json:"nedge_max"`
	NedgeMean   float64 `json:"nedge_mean"`
	NedgeStddev float64 `json:"nedge_stddev"`

	TEPSMin      float64 `json:"teps_min"`
	TEPSMedian   float64 `json:"teps_median"`
	TEPSMax      float64 `json:"teps_max"`
	TEPSHarmonic float64 `json:"teps_harmonic_mean"`
}
