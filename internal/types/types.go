package types

type GetResponse struct {
	Value     string `json:"value"`
	Exists    bool   `json:"exists"`
	OnNode    string `json:"onNode"`
	CacheSize int    `json:"cacheSize"`
}

type PutResponse struct {
	Ok        bool   `json:"ok"`
	OnNode    string `json:"onNode"`
	CacheSize int    `json:"cacheSize"`
}

type RepairResponse struct {
	Data map[string]string `json:"data"`
}
