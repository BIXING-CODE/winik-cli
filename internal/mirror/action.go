package mirror

// ActionRequest 对应 mirror 的 POST /v1/user/action 请求体（camelCase，空值省略）。
// startAt 与 timeRangeDesc 二选一；startAt 线格式 "yyyy-MM-dd HH:mm:ss"。
// mode: 0=OFFLINE 1=ONLINE；visibleStatus: 0=self 1=public。
type ActionRequest struct {
	ID            *int     `json:"id,omitempty"`
	VisibleStatus *int     `json:"visibleStatus,omitempty"`
	Title         string   `json:"title,omitempty"`
	Content       string   `json:"content,omitempty"`
	CoverImage    string   `json:"coverImage,omitempty"`
	Location      string   `json:"location,omitempty"`
	Lat           *float64 `json:"lat,omitempty"`
	Lng           *float64 `json:"lng,omitempty"`
	Price         string   `json:"price,omitempty"`
	TimeRangeDesc string   `json:"timeRangeDesc,omitempty"`
	StartAt       string   `json:"startAt,omitempty"`
	FragmentIDs   []int    `json:"fragmentIds,omitempty"`
	Mode          *int     `json:"mode,omitempty"`
}

type ActionResponse struct {
	ID            *int    `json:"id"`
	Title         string  `json:"title"`
	Content       string  `json:"content"`
	CoverImage    string  `json:"coverImage"`
	Location      string  `json:"location"`
	City          string  `json:"city"`
	Price         float64 `json:"price"`
	ActionStatus  string  `json:"actionStatus"`
	VisibleStatus any     `json:"visibleStatus"`
}

// PublishAction 发布（或编辑，带 ID 时）一条行动。
func (c *Client) PublishAction(req *ActionRequest) (*ActionResponse, error) {
	var out ActionResponse
	if err := c.do("POST", "/user/action", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
