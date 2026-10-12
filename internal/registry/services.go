package registry

type Service struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Status  string `json:"status"`
	Command string `json:"command"`
}

func Services() []Service {
	return []Service{
		{ID: "hast", Name: "Hast AIP", Status: "active", Command: "hast"},
		{ID: "paseo-bridge", Name: "Paseo Bridge", Status: "active", Command: "paseo-bridge"},
	}
}
