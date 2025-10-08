package udev

type Configuration struct {
	Subsystem string   `json:"subsystem"`
	Action    string   `json:"action"`
	Filters   []Filter `json:"filters"`
}

func DefaultConfiguration() Configuration {
	return Configuration{
		Subsystem: "block",
		Action:    "add",
		Filters:   []Filter{PrefixFilter{Prefix: "/dev/vdb"}},
	}
}
