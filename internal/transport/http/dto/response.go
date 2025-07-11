package dto

type SuccessResponse struct {
	Response any `json:"response"`
}

type ErrorResponse struct {
	Error Error `json:"error"`
}

type DataResponse struct {
	Data any `json:"data"`
}

type Error struct {
	Code int    `json:"code"`
	Text string `json:"text"`
}

type DocsResponse struct {
	JSON map[string]any `json:"json,omitempty"`
	File string         `json:"file,omitempty"`
}
