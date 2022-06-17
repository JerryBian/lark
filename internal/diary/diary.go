package internal

type Diary struct {
	Id             int64          `json:"id"`
	Contents       []DiaryContent `json:"contents"`
	CreatedAt      int64          `json:"created_at"`
	LastModifiedAt int64          `json:"last_modified_at"`

	Title string `json:"-"`
}

type DiaryContent struct {
	Id        int64  `json:"id"`
	DiaryId   int64  `json:"diary_id"`
	Content   string `json:"content"`
	Comment   string `json:"comment"`
	CreatedAt int64  `json:"created_at"`
}