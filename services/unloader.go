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

// создаем структуру жанров, при этом в структуре АНИМЕ будем использовать
// jikan.MAlItem из за конвертации.
type Genre struct {
	MalId int    `json:"mal_id"`
	Type  string `json:"type"`
	Name  string `json:"name"`
	Url   string `json:"url"`
}

// Структура для необходимых данных
type Anime struct {
	MalId        int           `json:"mal_id"`
	Url          string        `json:"url"`
	Images       jikan.Images3 `json:"images"`
	Title        string        `json:"title"`
	TitleEnglish string        `json:"title_english"`
	Type         string        `json:"type"`
	Episodes     int           `json:"episodes"`
	Status       string        `json:"status"`
	Rating       string        `json:"rating"`
	Score        float64       `json:"score"`
	Synopsis     string        `json:"synopsis"`
	Year         int           `json:"year"`
	Genres       []int         `json:"genres"`
}

// TODO: разбить объект Anime на несколько подъобъектов для удобной записи в супабейз

// получили поля Аниме. Получили поля жанров.
// жанры будут записываться в БД и проверяться на новые поля.
// аниме сущность будет хранить массив айдишников на жанры, что создает связь 1 ко мн
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

	// список и для жанров
	var genreList []Genre

	// запускаем все в цикле
	for i := 0; i < len(anime.Data); i++ {
		// TODO: так может не делать общую структуру а разбить по мелким объектам сразу ?
		// А нужно ли вообще разбивать все на разные таблицы ?

		// пришел к выводу о том, что ТИПЫ аниме выделять в отдельную таблицу не стоит
		// поскольку типы не нужны в аналитике и в данном проекте. они являются просто текстовым полем - не более
		// и влияют только на описание, в то время как ЖАНРЫ, которые следует выделить в отдельную таблицу,
		// помогут в реализации основной задумки проекта
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
		}

		for j := 0; j < len(anime.Data[i].Genres); j++ {
			gnr := Genre{
				MalId: anime.Data[i].Genres[j].MalId,
				Type:  anime.Data[i].Genres[j].Type,
				Name:  anime.Data[i].Genres[j].Name,
				Url:   anime.Data[i].Genres[j].Url,
			}

			anm.Genres = append(anm.Genres, gnr.MalId)

			genreList = append(genreList, gnr)
		}

		animeList = append(animeList, anm)
	}

	// выводим полученные объекты
	for i := 0; i < len(animeList); i++ {
		pp.Println(animeList[i])
	}

	uniqueGenres := removeDup(genreList)
	count1 := 0
	for i := 0; i < len(uniqueGenres); i++ {
		pp.Println(uniqueGenres[i])
		count1++
	}

	
}

// для удаления повторяющихся жанров
func removeDup(inputSlice []Genre) []Genre {
	isUnique := map[Genre]bool{}

	resultSlice := []Genre{}

	for _, item := range inputSlice {
		if !isUnique[item] {
			isUnique[item] = true
			resultSlice = append(resultSlice, item)
		}
	}

	return resultSlice
}
