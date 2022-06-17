package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"time"

	R "github.com/JerryBian/lark/internal/diary"
)

func main() {
	const v1file string = "../_/v1.json"
	j, err := ioutil.ReadFile(v1file)
	if err != nil {
		panic(err)
	}
	var words []Word
	err = json.Unmarshal([]byte(j), &words)
	if err != nil {
		panic(err)
	} else {
		var diaries []R.Diary
		sort.Slice(words, func(i, j int) bool {
			return words[i].Created_At < words[j].Created_At
		})
		for _, item := range words {
			diary := R.Diary {}
			t, err := time.Parse("2006-01-02 15:04:05", item.Created_At)
			if err != nil {
				panic(err)
			}

			fmt.Println(t)
			fmt.Println(t.Add(time.Hour * time.Duration(-8)))
			diary.CreatedAt = t.Add(time.Hour * time.Duration(-8)).UnixMicro()
			fmt.Println(diary.CreatedAt)
			diary.LastModifiedAt = time.Now().UTC().UnixMicro()
			fmt.Println(diary.LastModifiedAt)
			diary.Contents = append(diary.Contents, R.DiaryContent{ Content: item.Content, Comment: "Init version", CreatedAt: diary.CreatedAt })
			diaries = append(diaries, diary)
		}

		b, err := json.Marshal(diaries)
		if err != nil {
			panic(err)
		}

		fmt.Println(string(b))
		os.WriteFile("../_/v1_u.json", b, os.ModePerm)
	}
}

type Word struct {
	Id         int64
	Content    string
	Created_At string
}