package services

import (
	"fmt"

	"github.com/darenliang/jikan-go"
	"github.com/k0kubun/pp"
)

// структура лоадера 
type Unloader struct {
}

// конструктор лоадера
func NewUnloader() *Unloader {
	return &Unloader{}
}

// Структура для необходимых данных
type Anime struct {
	MalId        int             `json:"mal_id"`
	Url          string          `json:"url"`
	Images       jikan.Images3   `json:"images"`
	Title        string          `json:"title"`
	TitleEnglish string          `json:"title_english"`
	Type         string          `json:"type"`
	Episodes     int             `json:"episodes"`
	Status       string          `json:"status"`
	Rating       string          `json:"rating"`
	Score        float64         `json:"score"`
	Synopsis     string          `json:"synopsis"`
	Year         int             `json:"year"`
	Genres       []jikan.MalItem `json:"genres"`
}

//TODO: разбить объект Anime на несколько подъобъектов для удобной записи в супабейз
func (u *Unloader) Start() {
	// в этом месте получаем только 25 записей за 1 запрос.
	// нужно будет увеличить таймаут и итерировать страницы
	anime, err := jikan.GetTopAnime("tv", "bypopularity", 2)
	if err != nil {
		fmt.Println("error: ", err)
		return
	}
	
	// объявляем слайс, в котором будут лежать объекты аниме с нуными полями
	var animeList []Anime
	// запускаем все в цикле 
	for i := 0; i < len(anime.Data); i++ {
		anm := Anime{
			MalId:        anime.Data[i].MalId,
			Url:          anime.Data[i].Url,
			Images:       anime.Data[i].Images,
			Title:        anime.Data[i].Title,
			TitleEnglish: anime.Data[i].TitleEnglish,
			Type:         anime.Data[i].Type,
			Episodes:     anime.Data[i].Episodes,
			Status:       anime.Data[i].Status,
			Rating:       anime.Data[i].Rating,
			Score:        anime.Data[i].Score,
			Synopsis:     anime.Data[i].Synopsis,
			Year:         anime.Data[i].Year,
			Genres:       anime.Data[i].Genres,
		}
		animeList = append(animeList, anm)
	}

	// выводим полученные объекты 
	for i := 0; i < len(animeList) ; i++ {
		pp.Println(animeList[i])
	}
}
