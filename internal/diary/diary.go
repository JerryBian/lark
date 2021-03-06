package internal

import (
	"html/template"
)

type Diary struct {
	Id             int64          `json:"id"`
	Contents       []DiaryContent `json:"contents"`
	CreatedAt      int64          `json:"created_at"`
	LastModifiedAt int64          `json:"last_modified_at"`
	
	Revisions int `json:"-"`
	DayLink string `json:"-"`
	DayString string `json:"-"`
	CreatedAtString string `json:"-"`
	TimeOnlyStr string `json:"-"`
}

type DiaryContent struct {
	Id        int64  `json:"id"`
	DiaryId   int64  `json:"diary_id"`
	Content   string `json:"content"`
	Comment   string `json:"comment"`
	CreatedAt int64  `json:"created_at"`

	CreatedAtString string `json:"-"`
	HtmlContent template.HTML `json:"-"`
	ContentLen int `json:"-"`
}

type DiaryView struct {
	NextLink string `json:"-"`
	PreviousLink string `json:"-"`
	Diaries []Diary
}

type DiaryNav struct {
	Title string `json:"title"`
	Link string  `json:"link"`
	Tooltip string `json:"tooltip"`
	Children []DiaryNav `json:"children"`
	Key string `json:"-"`
}