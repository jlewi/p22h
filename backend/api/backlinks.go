package api

type BackLinkList struct {
	Items []BackLink `json:"items"`
}

// TODO(jeremy): Don't think this is the right data structure. What's the proper way to return BackLinks?
// We'd like information about the source document (e.g. its URI as well as tne target document).
type BackLink struct {
	Text  string `json:"text"`
	DocId string `json:"docId"`
}
