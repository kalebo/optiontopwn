package common

type Record struct {
	Victim      string `json:"victim"`
	Perpetrator string `json:"perpetrator"`
	Host        string `json:"host"`
}
