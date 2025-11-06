package metadata

type Meta struct {
	Name    string `json:"name"`
	Devices []Disk `json:"devices"`
}

type Disk struct {
	Type    string   `json:"type"`
	Address string   `json:"address"`
	Bus     string   `json:"bus"`
	Serial  string   `json:"serial"`
	Tags    []string `json:"tags"`
}
