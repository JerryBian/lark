package internal

import (
	C "github.com/JerryBian/lark/internal/config"

	"database/sql"
	"errors"
	"log"
	"time"
	"os"
)

type Db struct {
	Conf *C.Config
}

func init() {
	log.SetOutput(os.Stdout)
}

func (s *Db) CountDiaries() (int64, error) {
	database, _ := sql.Open("sqlite3", s.Conf.Database.ConnStr)
	defer database.Close()

	rows, err := database.Query(`
		SELECT COUNT(*) FROM diary
	`)

	if err != nil {
		return 0, err
	}

	var count int64
	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(&count); err != nil {
			return 0, err
		}
	}

	return count, nil
}

func (s *Db) AddDiary(d Diary) (*Diary, error) {
	log.Println("add new diary ...")
	if len(d.Contents) == 0 {
		return nil, errors.New("No content in diary.")
	}

	now := time.Now().UTC().UnixMicro()
	createdAt := now
	lastModifedAt := now
	if d.CreatedAt != 0 {
		createdAt = d.CreatedAt
	}

	if d.LastModifiedAt != 0 {
		lastModifedAt = d.LastModifiedAt
	}
	database, _ := sql.Open("sqlite3", s.Conf.Database.ConnStr)
	defer database.Close()

	if d.Id > 0 {
		var exist bool
		row := database.QueryRow("SELECT EXISTS(SELECT id FROM diary WHERE id=? LIMIT 1)", d.Id)
		if err := row.Scan(&exist); err != nil {
			return nil, err
		}

		_, err := database.Exec("UPDATE diary SET last_modified_at=? WHERE id=?", lastModifedAt, d.Id)
		if err != nil {
			return nil, err
		}
	} else {
		res, err := database.Exec("INSERT INTO diary(created_at, last_modified_at) VALUES(?,?)", createdAt, lastModifedAt)
		if err != nil {
			return nil, err
		}

		id, err := res.LastInsertId()
		if err != nil {
			return nil, err
		}

		d.Id = id
	}


	for _, content := range d.Contents {
		createdAt = now
		if content.CreatedAt != 0 {
			createdAt = content.CreatedAt
		}
		res, err := database.Exec("INSERT INTO diary_content(diary_id, content, comment, created_at) VALUES(?,?,?,?)", d.Id, content.Content, content.Comment, createdAt)
		if err != nil {
			return nil, err
		}
	
		id, err := res.LastInsertId()
		if err != nil {
			return nil, err
		}
	
		content.Id = id
	}
	
	s.Conf.Runtime.LastModifiedAt = time.Now().Local()
	log.Println("finished add new diary")
	return &d, nil
}

// This is very expensive, use at your risk
func (s *Db) Dump() ([]Diary, error) {
	database, _ := sql.Open("sqlite3", s.Conf.Database.ConnStr)
	defer database.Close()

	rows, err := database.Query("SELECT id, created_at, last_modified_at FROM diary ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var diaries []Diary
	for rows.Next() {
		var d Diary
		if err := rows.Scan(&d.Id, &d.CreatedAt, &d.LastModifiedAt); err != nil {
			return nil, err
		}

		diaries = append(diaries, d)
	}

	for i := range diaries {
		diaries[i].Title = time.UnixMicro(diaries[i].CreatedAt).Format("2006年01月02日 15时04分05秒")

		rows, err := database.Query("SELECT id, diary_id, content, comment, created_at FROM diary_content WHERE diary_id = ? ORDER BY created_at DESC", diaries[i].Id)
		if err != nil {
			return nil, err
		}

		defer rows.Close()

		var contents []DiaryContent
		for rows.Next() {
			var content DiaryContent
			if err := rows.Scan(&content.Id, &content.DiaryId, &content.Content, &content.Comment, &content.CreatedAt); err != nil {
				return nil, err
			}

			contents = append(contents, content)
		}

		diaries[i].Contents = contents
	}

	return diaries, nil
}
