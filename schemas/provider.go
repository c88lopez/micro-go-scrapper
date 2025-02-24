package schemas

type Provider struct {
	ID     int      `json:"id"`
	Name   string   `json:"name"`
	Link   string   `json:"link"`
	Phone  string   `json:"phone"`
	Place  string   `json:"place"`
	Source string   `json:"source"`
	Pics   []string `json:"pics"`
}
