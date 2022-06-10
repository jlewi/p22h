package api

type RequestStatus struct {
	Kind    string `json:"kind"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}
