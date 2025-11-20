package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/darenliang/jikan-go"
	"github.com/joho/godotenv"
)

type Unloader struct {
}

func NewUnloader() *Unloader {
	return &Unloader{}
}

type Genre struct {
	ID    int64  `json:"id,omitempty"`
	MalId int    `json:"mal_id"`
	Type  string `json:"type"`
	Name  string `json:"name"`
	Url   string `json:"url"`
}

type Anime struct {
	ID           int64         `json:"id,omitempty"`
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
}

type AnimeGenre struct {
	AnimeID int64 `json:"anime_id"`
	GenreID int64 `json:"genre_id"`
}

// Start - выгрузка аниме по страницам
func (u *Unloader) Start() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Warning: Could not load .env file")
	}

	var page int
	fmt.Print("Введите номер страницы для выгрузки: ")
	fmt.Scan(&page)

	fmt.Printf("Загружаем страницу %d...\n", page)

	anime, err := jikan.GetTopAnime("tv", "airing", page)
	if err != nil {
		fmt.Printf("Ошибка загрузки страницы %d: %v\n", page, err)
		return
	}

	if len(anime.Data) == 0 {
		fmt.Println("Страница пустая")
		return
	}

	fmt.Printf("Получено %d аниме для обработки\n", len(anime.Data))

	totalAnime := 0
	totalGenres := 0
	totalRelations := 0

	for i := 0; i < len(anime.Data); i++ {
		animeItem := Anime{
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

		fmt.Printf("[%d] %s\n", i+1, animeItem.Title)

		insertedAnime, err := insertAnime(animeItem)
		if err != nil {
			fmt.Printf("Ошибка аниме: %v\n", err)
			continue
		}

		if insertedAnime.ID == 0 {
			fmt.Printf("Аниме уже существует\n")
		} else {
			totalAnime++
			fmt.Printf("Аниме добавлено (ID: %d)\n", insertedAnime.ID)
		}

		fmt.Printf("Жанров: %d\n", len(anime.Data[i].Genres))
		genreCount := 0

		for j := 0; j < len(anime.Data[i].Genres); j++ {
			genreItem := Genre{
				MalId: anime.Data[i].Genres[j].MalId,
				Type:  anime.Data[i].Genres[j].Type,
				Name:  anime.Data[i].Genres[j].Name,
				Url:   anime.Data[i].Genres[j].Url,
			}

			insertedGenre, err := insertGenre(genreItem)
			if err != nil {
				fmt.Printf("Ошибка жанра %s: %v\n", genreItem.Name, err)
				continue
			}

			if insertedGenre.ID == 0 {
				fmt.Printf("Жанр уже существует: %s\n", insertedGenre.Name)
			} else {
				totalGenres++
				fmt.Printf("Жанр добавлен: %s (ID: %d)\n", insertedGenre.Name, insertedGenre.ID)
			}

			err = insertAnimeGenre(insertedAnime.ID, insertedGenre.ID)
			if err != nil {
				fmt.Printf("Ошибка связи: %v\n", err)
				continue
			}

			totalRelations++
			genreCount++
			fmt.Printf("Связь создана\n")
		}

		fmt.Printf("Создано связей: %d/%d\n", genreCount, len(anime.Data[i].Genres))
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Printf("Страница %d завершена\n", page)
	fmt.Printf("Итого за страницу:\n")
	fmt.Printf("Новых аниме: %d\n", totalAnime)
	fmt.Printf("Новых жанров: %d\n", totalGenres)
	fmt.Printf("Создано связей: %d\n", totalRelations)
}

// insertAnime - вставка аниме с проверкой существования
func insertAnime(animeItem Anime) (*Anime, error) {
	exists, animeID, err := animeExists(animeItem.MalId)
	if err != nil {
		return nil, fmt.Errorf("failed to check anime existence: %w", err)
	}

	if exists {
		fmt.Printf("Аниме уже существует: %s (ID: %d)\n", animeItem.Title, animeID)
		return &Anime{
			ID:           animeID,
			MalId:        animeItem.MalId,
			Url:          animeItem.Url,
			Images:       animeItem.Images,
			Title:        animeItem.Title,
			TitleEnglish: animeItem.TitleEnglish,
			Type:         animeItem.Type,
			Episodes:     animeItem.Episodes,
			Status:       animeItem.Status,
			Rating:       animeItem.Rating,
			Score:        animeItem.Score,
			Synopsis:     animeItem.Synopsis,
			Year:         animeItem.Year,
		}, nil
	}

	err = godotenv.Load()
	if err != nil {
		fmt.Println("Warning: Could not load .env file")
	}

	supabaseKey := os.Getenv("API_KEY")
	supabaseURL := os.Getenv("API_URL")

	if supabaseURL == "" {
		return nil, fmt.Errorf("SUPABASE_URL is empty")
	}
	if supabaseKey == "" {
		return nil, fmt.Errorf("SUPABASE_KEY is empty")
	}

	jsonData, err := json.Marshal(animeItem)
	if err != nil {
		return nil, fmt.Errorf("marshal error: %w", err)
	}

	url := fmt.Sprintf("%s/rest/v1/anime", supabaseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("request creation error: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", supabaseKey)
	req.Header.Set("Authorization", "Bearer "+supabaseKey)
	req.Header.Set("Prefer", "return=representation")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 201 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	var result []Anime
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("no anime returned after insert")
	}

	return &result[0], nil
}

// animeExists - проверка существования аниме
func animeExists(malId int) (bool, int64, error) {
	supabaseKey := os.Getenv("API_KEY")
	supabaseURL := os.Getenv("API_URL")

	url := fmt.Sprintf("%s/rest/v1/anime?mal_id=eq.%d&select=id", supabaseURL, malId)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, 0, err
	}

	req.Header.Set("apikey", supabaseKey)
	req.Header.Set("Authorization", "Bearer "+supabaseKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return false, 0, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	var result []struct {
		ID int64 `json:"id"`
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return false, 0, err
	}

	if len(result) > 0 {
		return true, result[0].ID, nil
	}

	return false, 0, nil
}

// insertGenre - вставка жанра с проверкой существования
func insertGenre(genreItem Genre) (*Genre, error) {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Warning: Could not load .env file")
	}

	exists, genreID, err := genreExists(genreItem.MalId)
	if err != nil {
		return nil, err
	}

	if exists {
		return &Genre{ID: genreID, MalId: genreItem.MalId, Name: genreItem.Name, Type: genreItem.Type, Url: genreItem.Url}, nil
	}

	supabaseKey := os.Getenv("API_KEY")
	supabaseURL := os.Getenv("API_URL")

	if supabaseURL == "" {
		return nil, fmt.Errorf("SUPABASE_URL is empty")
	}
	if supabaseKey == "" {
		return nil, fmt.Errorf("SUPABASE_KEY is empty")
	}

	jsonData, err := json.Marshal(genreItem)
	if err != nil {
		return nil, fmt.Errorf("marshal error: %w", err)
	}

	url := fmt.Sprintf("%s/rest/v1/genres", supabaseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("request creation error: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", supabaseKey)
	req.Header.Set("Authorization", "Bearer "+supabaseKey)
	req.Header.Set("Prefer", "return=representation")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 201 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	var result []Genre
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("no genre returned after insert")
	}

	return &result[0], nil
}

// genreExists - проверка существования жанра
func genreExists(malId int) (bool, int64, error) {
	supabaseKey := os.Getenv("API_KEY")
	supabaseURL := os.Getenv("API_URL")

	url := fmt.Sprintf("%s/rest/v1/genres?mal_id=eq.%d&select=id", supabaseURL, malId)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, 0, err
	}

	req.Header.Set("apikey", supabaseKey)
	req.Header.Set("Authorization", "Bearer "+supabaseKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return false, 0, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	var result []struct {
		ID int64 `json:"id"`
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return false, 0, err
	}

	if len(result) > 0 {
		return true, result[0].ID, nil
	}

	return false, 0, nil
}

// insertAnimeGenre - создание связи аниме-жанр
func insertAnimeGenre(animeID, genreID int64) error {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Warning: Could not load .env file")
	}

	supabaseKey := os.Getenv("API_KEY")
	supabaseURL := os.Getenv("API_URL")

	if supabaseURL == "" {
		return fmt.Errorf("SUPABASE_URL is empty")
	}
	if supabaseKey == "" {
		return fmt.Errorf("SUPABASE_KEY is empty")
	}

	animeGenreItem := AnimeGenre{
		AnimeID: animeID,
		GenreID: genreID,
	}

	jsonData, err := json.Marshal(animeGenreItem)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}

	url := fmt.Sprintf("%s/rest/v1/anime_genres", supabaseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("request creation error: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", supabaseKey)
	req.Header.Set("Authorization", "Bearer "+supabaseKey)
	req.Header.Set("Prefer", "return=representation")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 201 {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	var result []AnimeGenre
	err = json.Unmarshal(body, &result)
	if err != nil {
		fmt.Printf("Note: Could not decode response: %v\n", err)
	}

	return nil
}
