package internal

import (
	"encoding/json"
	"log"
	"os"
	"sort"
	"time"

	C "github.com/JerryBian/lark/internal/config"
	R "github.com/JerryBian/lark/internal/diary"
)

func init() {
	log.SetOutput(os.Stdout)
}

type AzTask struct {
	Config         *C.Config
	Stopping       bool
}

func (at *AzTask) SpinWorker() {
	if at.Config.Az.ConnStr == "" {
		log.Println("No ENV_AZ_CONNSTR provided, so skip az worker.")
		return
	}

	repo := R.Db{Conf: at.Config}
	now := time.Now()
	at.Config.Runtime.LastBackupAt = now
	at.Config.Runtime.LastModifiedAt = now
	CreateContainerIfNotExists(at.Config)

	for !at.Stopping {
		if !at.Stopping {
			time.Sleep(time.Minute * time.Duration(at.Config.Az.BackupIntervalInMinutes))
		}

		if at.Config.Runtime.LastBackupAt.After(at.Config.Runtime.LastModifiedAt) || at.Config.Runtime.LastBackupAt.Equal(at.Config.Runtime.LastModifiedAt) {
			continue
		}

		log.Println("Detect changes, running az backup...")
		start := time.Now()
		words, err := repo.Dump()
		if err != nil {
			log.Println(err)
		}

		if err != nil {
			log.Println(err)
		} else {
			b, err := json.Marshal(words)
			if err != nil {
				log.Println(err)
			} else {
				err = Save("", b, at.Config)
				if err != nil {
					log.Println(err)
				} else {
					at.Config.Runtime.LastBackupAt = time.Now().Local()
					log.Println("JSON saved successfully.")
				}

				err = Save("-"+time.Now().Local().Format("20060102150405"), b, at.Config)
				if err != nil {
					log.Println(err)
				} else {
					at.Config.Runtime.LastBackupAt = time.Now().Local()
					log.Println("JSON snapshot saved successfully.")
				}
			}
		}
		
		elapsed := time.Since(start)
		log.Printf("Ran az backup completed, elasped: %s.\n", elapsed)
	}
}

func (at *AzTask) TryRestore() {
	if at.Config.Az.ConnStr == "" {
		log.Println("No ENV_AZ_CONNSTR provided, so skip az restore.")
		return
	}

	log.Println("Running az restore...")
	start := time.Now()
	repo := R.Db{Conf: at.Config}
	count, err := repo.CountDiaries()
	if err != nil {
		panic(err)
	}

	// only restory while no records exist
	if count == 0 {
		b, err := Load(at.Config)
		if err != nil {
			panic(err)
		} else {
			var words []R.Diary
			err = json.Unmarshal(b, &words)
			if err != nil {
				panic(err)
			} else {
				sort.Slice(words, func(i, j int) bool {
					return words[i].CreatedAt < words[j].CreatedAt
				})
				for _, item := range words {
					item.Id = 0
					_, err = repo.AddDiary(item)
					if err != nil {
						panic(err)
					}
				}
			}
		}
	}

	elapsed := time.Since(start)
	log.Printf("Ran az restore completed, elasped: %s.\n", elapsed)
}
