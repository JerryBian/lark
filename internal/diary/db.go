package internal

import (
	"fmt"

	C "github.com/JerryBian/lark/internal/config"
	"golang.org/x/exp/slices"

	"database/sql"
	"errors"
	"log"
	"os"
	"time"
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

func (s *Db) GetDiaryById(id int) (Diary, error) {
	database, _ := sql.Open("sqlite3", s.Conf.Database.ConnStr)
	defer database.Close()

	rows, err := database.Query("SELECT id, created_at, last_modified_at FROM diary WHERE id = ?", id)
	if err != nil {
		return Diary{}, err
	}

	defer rows.Close()
	var d Diary
	for rows.Next() {
		if err := rows.Scan(&d.Id, &d.CreatedAt, &d.LastModifiedAt); err != nil {
			return Diary{}, err
		}
	}

	d.Title = time.UnixMicro(d.CreatedAt).Format("2006年01月02日 15时04分05秒")
	rows, err = database.Query("SELECT id, created_at, diary_id, content, comment FROM diary_content WHERE diary_id = ? ORDER BY created_at DESC", d.Id)
	if err != nil {
		return Diary{}, err
	}

	var cs []DiaryContent
	for rows.Next() {
		var c DiaryContent
		if err := rows.Scan(&c.Id, &c.CreatedAt, &c.DiaryId, &c.Content, &c.Comment); err != nil {
			return Diary{}, err
		}

		c.Title = time.UnixMicro(d.CreatedAt).Format("15时04分05秒")
		cs = append(cs, c)
	}

	d.Contents = cs
	return d, nil
}

func (s *Db) GetDiaries(year int, month int, day int) ([]Diary, error) {
	database, _ := sql.Open("sqlite3", s.Conf.Database.ConnStr)
	defer database.Close()

	start := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC).UnixMicro()
	end := time.Date(year, time.Month(month), day, 23, 59, 59, 999, time.UTC).UnixMicro()
	rows, err := database.Query("SELECT id FROM diary WHERE created_at BETWEEN ? AND ? ORDER BY created_at DESC", start, end)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	var r []Diary
	for rows.Next() {
		var diaryId int
		if err := rows.Scan(&diaryId); err != nil {
			return nil, err
		}

		d, err := s.GetDiaryById(diaryId)
		if err != nil {
			return nil, err
		}

		r = append(r, d)
	}

	return r, nil
}

func (s *Db) GetDiaryNavs() ([]DiaryNav, error) {
	database, _ := sql.Open("sqlite3", s.Conf.Database.ConnStr)
	defer database.Close()

	rows, err := database.Query("SELECT id, created_at FROM diary ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	r := []DiaryNav{}
	for rows.Next() {
		var d Diary
		if err := rows.Scan(&d.Id, &d.CreatedAt); err != nil {
			return nil, err
		}

		createdAt := time.UnixMicro(d.CreatedAt)
		day := createdAt.Format("02")
		year := createdAt.Format("2006")
		month := createdAt.Format("01")

		yKey := fmt.Sprintf("k%s", year)
		idx := slices.IndexFunc(r, func(x DiaryNav) bool {
			return x.Key == yKey
		})
		if idx == -1 {
			r = append(r, DiaryNav{
				Title: fmt.Sprintf("%s年", year),
				Key:   yKey,
			})

			idx = slices.IndexFunc(r, func(x DiaryNav) bool {
				return x.Key == yKey
			})
		}

		mKey := fmt.Sprintf("k%s%s", year, month)
		idy := slices.IndexFunc(r[idx].Children, func(x DiaryNav) bool {
			return x.Key == mKey
		})

		if idy == -1 {
			r[idx].Children = append(r[idx].Children, DiaryNav{
				Title: fmt.Sprintf("%s月", month),
				Key:   mKey,
			})

			idy = slices.IndexFunc(r[idx].Children, func(x DiaryNav) bool {
				return x.Key == mKey
			})
		}

		dKey := fmt.Sprintf("%s%s%s", year, month, day)
		idz := slices.IndexFunc(r[idx].Children[idy].Children, func(x DiaryNav) bool {
			return x.Key == dKey
		})

		if idz == -1 {
			r[idx].Children[idy].Children = append(r[idx].Children[idy].Children, DiaryNav{
				Title: fmt.Sprintf("%s日", day),
				Link:  fmt.Sprintf("/diary/%s/%s/%s", year, month, day),
				Key:   dKey,
			})
		}
	}

	return r, nil
}
