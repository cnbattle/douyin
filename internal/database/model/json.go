package model

type Data struct {
	StatusCode int    `json:"status_code"`
	MinCursor  int    `json:"min_cursor"`
	MaxCursor  int    `json:"max_cursor"`
	HasMore    int    `json:"has_more"`
	AwemeList  []Item `json:"aweme_list"`
}

type Item struct {
	AwemeId      string      `json:"aweme_id"`
	Desc         string      `json:"desc"`
	CreateTime   int         `json:"create_time"`
	Author       author      `json:"author"`
	Video        video       `json:"video"`
	IsAds        bool        `json:"is_ads"`
	Duration     int         `json:"duration"`
	GroupId      string      `json:"group_id"`
	AuthorUserId int64       `json:"author_user_id"`
	LongVideo    []longVideo `json:"long_video"`
	Statistics   statistics  `json:"statistics"`
	ShareInfo    shareInfo   `json:"share_info"`
}

type author struct {
	Uid         string `json:"uid"`
	Nickname    string `json:"nickname"`
	Gender      int    `json:"gender"`
	UniqueId    string `json:"unique_id"`
	AvatarThumb uriStr `json:"avatar_thumb"`
}

type video struct {
	PlayAddr uriStr `json:"play_addr"`
	Cover    uriStr `json:"cover"`
}

type longVideo struct {
	Video            video `json:"video"`
	TrailerStartTime int   `json:"trailer_start_time"`
}

type uriStr struct {
	Uri     string   `json:"uri"`
	UrlList []string `json:"url_list"`
	Width   int      `json:"width"`
	Height  int      `json:"height"`
}

type statistics struct {
	AwemeId           string `json:"aweme_id"`
	CommentCount      int    `json:"comment_count"`
	DiggCount         int    `json:"digg_count"`
	DownloadCount     int    `json:"download_count"`
	PlayCount         int    `json:"play_count"`
	ShareCount        int    `json:"share_count"`
	ForwardCount      int    `json:"forward_count"`
	LoseCount         int    `json:"lose_count"`
	LoseComment_count int    `json:"lose_comment_count"`
}

type shareInfo struct {
	ShareUrl   string `json:"share_url"`
	ShareTitle string `json:"share_title"`
}
