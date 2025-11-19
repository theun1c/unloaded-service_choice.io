package services

import (
	"fmt"
	"time"

	"github.com/darenliang/jikan-go"
	"github.com/k0kubun/pp"
)

type Unloader struct {
}

func NewUnloader() *Unloader {
	return &Unloader{}
}

// take all 50 animes in 1ne json from https://myanimelist.net/topanime.php?type=bypopularity page
func (u *Unloader) Start() {
	for i := 1; i <= 1; i++ {
		anime, err := jikan.GetTopAnime(jikan.TopAnimeTypeTv, "bypopularity", 1) // https://myanimelist.net/topanime.php?type=bypopularity
		if err != nil {
			fmt.Println(err)
		} else {
			pp.Println(anime.Data)
		}

		time.Sleep(10 * time.Second)
	}
}
