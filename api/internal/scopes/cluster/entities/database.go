package entities

type DatabasesFilter struct {
	Sort  string
	Order string
}

type DatabaseDetails struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	Owner             string `json:"owner"`
	Encoding          string `json:"encoding"`
	Collate           string `json:"collate"`
	Ctype             string `json:"ctype"`
	IsTemplate        bool   `json:"isTemplate"`
	AllowConnections  bool   `json:"allowConnections"`
	ConnectionLimit   int    `json:"connectionLimit"`
	SizeBytes         int64  `json:"sizeBytes"`
	SizePretty        string `json:"sizePretty"`
	ActiveConnections int    `json:"activeConnections"`
}
