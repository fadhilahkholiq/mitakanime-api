package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
)

type APIResponse struct {
	Status        string      `json:"status"`
	Creator       string      `json:"creator"`
	StatusCode    int         `json:"statusCode"`
	StatusMessage string      `json:"statusMessage"`
	Message       string      `json:"message"`
	Ok            bool        `json:"ok"`
	Data          interface{} `json:"data"`
	Pagination    interface{} `json:"pagination"`
}
type Pagination struct {
	CurrentPage int  `json:"currentPage"`
	HasPrevPage bool `json:"hasPrevPage"`
	PrevPage    *int `json:"prevPage"`
	HasNextPage bool `json:"hasNextPage"`
	NextPage    *int `json:"nextPage"`
	TotalPages  int  `json:"totalPages"`
}
type AnimeOngoing struct {
	Title             string `json:"title"`
	Poster            string `json:"poster"`
	Episodes          int    `json:"episodes"`
	ReleaseDay        string `json:"releaseDay"`
	LatestReleaseDate string `json:"latestReleaseDate"`
	AnimeId           string `json:"animeId"`
	Href              string `json:"href"`
	OtakudesuUrl      string `json:"otakudesuUrl"`
}
type AnimeCompleted struct {
	Title           string `json:"title"`
	Poster          string `json:"poster"`
	Episodes        int    `json:"episodes"`
	Score           string `json:"score"`
	LastReleaseDate string `json:"lastReleaseDate"`
	AnimeId         string `json:"animeId"`
	Href            string `json:"href"`
	OtakudesuUrl    string `json:"otakudesuUrl"`
}
type HomeData struct {
	Ongoing struct {
		Href         string         `json:"href"`
		OtakudesuUrl string         `json:"otakudesuUrl"`
		AnimeList    []AnimeOngoing `json:"animeList"`
	} `json:"ongoing"`
	Completed struct {
		Href         string           `json:"href"`
		OtakudesuUrl string           `json:"otakudesuUrl"`
		AnimeList    []AnimeCompleted `json:"animeList"`
	} `json:"completed"`
}
type AnimeSchedule struct {
	Title  string `json:"title"`
	Slug   string `json:"slug"`
	Url    string `json:"url"`
	Poster string `json:"poster"`
}
type DaySchedule struct {
	Day       string          `json:"day"`
	AnimeList []AnimeSchedule `json:"anime_list"`
}
type Genre struct {
	Title        string `json:"title"`
	GenreId      string `json:"genreId"`
	Href         string `json:"href"`
	OtakudesuUrl string `json:"otakudesuUrl"`
}
type AnimeBaseData struct {
	Title        string  `json:"title"`
	Poster       string  `json:"poster"`
	Episodes     *int    `json:"episodes"`
	Score        string  `json:"score,omitempty"`
	Season       string  `json:"season,omitempty"`
	AnimeId      string  `json:"animeId"`
	Href         string  `json:"href"`
	OtakudesuUrl string  `json:"otakudesuUrl"`
	Studios      string  `json:"studios,omitempty"`
	Status       string  `json:"status,omitempty"`
	GenreList    []Genre `json:"genreList,omitempty"`
}
type ListAnimeData struct {
	AnimeList []AnimeBaseData `json:"animeList"`
}
type Episode struct {
	Title        string `json:"title"`
	Eps          int    `json:"eps"`
	Date         string `json:"date"`
	EpisodeId    string `json:"episodeId"`
	Href         string `json:"href"`
	OtakudesuUrl string `json:"otakudesuUrl"`
}
type DetailAnimeData struct {
	Title       string    `json:"title"`
	Poster      string    `json:"poster"`
	Japanese    string    `json:"japanese"`
	Score       string    `json:"score"`
	Producers   string    `json:"producers"`
	Type        string    `json:"type"`
	Status      string    `json:"status"`
	Episodes    int       `json:"episodes"`
	Duration    string    `json:"duration"`
	Aired       string    `json:"aired"`
	Studios     string    `json:"studios"`
	Synopsis    Synopsis  `json:"synopsis"`
	GenreList   []Genre   `json:"genreList"`
	EpisodeList []Episode `json:"episodeList"`
}
type Synopsis struct {
	Paragraphs []string `json:"paragraphs"`
}
type NavEpisode struct {
	Title        string `json:"title"`
	EpisodeId    string `json:"episodeId"`
	Href         string `json:"href"`
	OtakudesuUrl string `json:"otakudesuUrl"`
}
type ServerLink struct {
	Title    string `json:"title"`
	ServerId string `json:"serverId"`
	Href     string `json:"href"`
}
type ServerQuality struct {
	Title      string       `json:"title"`
	ServerList []ServerLink `json:"serverList"`
}
type ServerInfo struct {
	Qualities []ServerQuality `json:"qualities"`
}
type DownloadLink struct {
	Title string `json:"title"`
	Url   string `json:"url"`
}
type DownloadQuality struct {
	Title string         `json:"title"`
	Size  string         `json:"size"`
	Urls  []DownloadLink `json:"urls"`
}
type DownloadInfo struct {
	Qualities []DownloadQuality `json:"qualities"`
}
type MetaInfo struct {
	Credit      string    `json:"credit"`
	Encoder     string    `json:"encoder"`
	Duration    string    `json:"duration"`
	Type        string    `json:"type"`
	GenreList   []Genre   `json:"genreList"`
	EpisodeList []Episode `json:"episodeList"`
}
type EpisodeData struct {
	Title               string       `json:"title"`
	AnimeId             string       `json:"animeId"`
	ReleaseTime         string       `json:"releaseTime"`
	DefaultStreamingUrl string       `json:"defaultStreamingUrl"`
	HasPrevEpisode      bool         `json:"hasPrevEpisode"`
	PrevEpisode         *NavEpisode  `json:"prevEpisode,omitempty"`
	HasNextEpisode      bool         `json:"hasNextEpisode"`
	NextEpisode         *NavEpisode  `json:"nextEpisode,omitempty"`
	Server              ServerInfo   `json:"server"`
	DownloadUrl         DownloadInfo `json:"downloadUrl"`
	Info                MetaInfo     `json:"info"`
}

func extractNumber(text string) int {
	re := regexp.MustCompile(`\d+`)
	match := re.FindString(text)
	if match != "" {
		val, _ := strconv.Atoi(match)
		return val
	}
	return 0
}
func extractEpisodeNumber(text string) int {
	re := regexp.MustCompile(`(?i)episode\s+(\d+)`)
	match := re.FindStringSubmatch(text)
	if len(match) > 1 {
		val, _ := strconv.Atoi(match[1])
		return val
	}
	return extractNumber(text)
}
func extractAnimeId(urlStr string) string {
	parts := strings.Split(strings.TrimSuffix(urlStr, "/"), "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}
func getHTML(url string) (*goquery.Document, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	return goquery.NewDocumentFromReader(res.Body)
}
func getPoster(animeId string) string {
	doc, err := getHTML("https://otakudesu.best/anime/" + animeId)
	if err != nil {
		return ""
	}
	posterUrl, _ := doc.Find(".fotoanime img").Attr("src")
	return posterUrl
}
func parsePagination(doc *goquery.Document, currentPageStr string) *Pagination {
	page, _ := strconv.Atoi(currentPageStr)
	if page < 1 {
		page = 1
	}
	paginationDiv := doc.Find(".pagenavix")
	if paginationDiv.Length() == 0 {
		return nil
	}
	totalPages := page
	paginationDiv.Find(".page-numbers").Each(func(i int, s *goquery.Selection) {
		num, err := strconv.Atoi(s.Text())
		if err == nil && num > totalPages {
			totalPages = num
		}
	})
	var prev, next *int
	if page > 1 {
		p := page - 1
		prev = &p
	}
	if page < totalPages {
		n := page + 1
		next = &n
	}
	return &Pagination{CurrentPage: page, HasPrevPage: prev != nil, PrevPage: prev, HasNextPage: next != nil, NextPage: next, TotalPages: totalPages}
}
func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	anime := r.Group("/anime")
	{
		anime.GET("/home", func(c *gin.Context) {
			doc, err := getHTML("https://otakudesu.best/")
			if err != nil {
				c.JSON(500, gin.H{"error": "Gagal scraping"})
				return
			}
			var ongoingList []AnimeOngoing = []AnimeOngoing{}
			var completedList []AnimeCompleted = []AnimeCompleted{}
			doc.Find(".venz").Eq(0).Find("ul > li").Each(func(i int, s *goquery.Selection) {
				url, _ := s.Find(".thumb a").Attr("href")
				id := extractAnimeId(url)
				epText := s.Find(".epz").Text()
				img, _ := s.Find(".thumbz img").Attr("src")
				ongoingList = append(ongoingList, AnimeOngoing{Title: s.Find(".jdlflm").Text(), Poster: img, Episodes: extractNumber(epText), ReleaseDay: strings.TrimSpace(s.Find(".epztipe").Text()), LatestReleaseDate: strings.TrimSpace(s.Find(".newnime").Text()), AnimeId: id, Href: "/anime/anime/" + id, OtakudesuUrl: url})
			})
			doc.Find(".venz").Eq(1).Find("ul > li").Each(func(i int, s *goquery.Selection) {
				url, _ := s.Find(".thumb a").Attr("href")
				id := extractAnimeId(url)
				epText := s.Find(".epz").Text()
				img, _ := s.Find(".thumbz img").Attr("src")
				completedList = append(completedList, AnimeCompleted{Title: s.Find(".jdlflm").Text(), Poster: img, Episodes: extractNumber(epText), Score: strings.TrimSpace(s.Find(".epztipe").Text()), LastReleaseDate: strings.TrimSpace(s.Find(".newnime").Text()), AnimeId: id, Href: "/anime/anime/" + id, OtakudesuUrl: url})
			})
			response := APIResponse{Status: "success", Creator: "Asa Mitaka", StatusCode: 200, StatusMessage: "OK", Ok: true, Data: HomeData{Ongoing: struct {
				Href         string         `json:"href"`
				OtakudesuUrl string         `json:"otakudesuUrl"`
				AnimeList    []AnimeOngoing `json:"animeList"`
			}{Href: "/anime/ongoing-anime", OtakudesuUrl: "https://otakudesu.best/ongoing-anime/", AnimeList: ongoingList}, Completed: struct {
				Href         string           `json:"href"`
				OtakudesuUrl string           `json:"otakudesuUrl"`
				AnimeList    []AnimeCompleted `json:"animeList"`
			}{Href: "/anime/complete-anime", OtakudesuUrl: "https://otakudesu.best/complete-anime/", AnimeList: completedList}}, Pagination: nil}
			c.IndentedJSON(200, response)
		})
		anime.GET("/schedule", func(c *gin.Context) {
			doc, err := getHTML("https://otakudesu.best/jadwal-rilis/")
			if err != nil {
				c.JSON(500, gin.H{"error": "Gagal scraping"})
				return
			}
			var scheduleData []DaySchedule = []DaySchedule{}
			doc.Find(".kglist321").Each(func(i int, s *goquery.Selection) {
				dayName := s.Find("h2").Text()
				var animeList []AnimeSchedule = []AnimeSchedule{}
				s.Find("ul > li").Each(func(j int, li *goquery.Selection) {
					title := li.Find("a").Text()
					url, _ := li.Find("a").Attr("href")
					slug := extractAnimeId(url)
					animeList = append(animeList, AnimeSchedule{Title: title, Slug: slug, Url: "/anime/anime/" + slug, Poster: getPoster(slug)})
				})
				scheduleData = append(scheduleData, DaySchedule{Day: dayName, AnimeList: animeList})
			})
			c.IndentedJSON(200, APIResponse{Status: "success", Creator: "Asa Mitaka", StatusCode: 200, StatusMessage: "OK", Ok: true, Data: scheduleData, Pagination: nil})
		})
		anime.GET("/genre", func(c *gin.Context) {
			doc, err := getHTML("https://otakudesu.best/genre-list/")
			if err != nil {
				c.JSON(500, gin.H{"error": "Gagal scraping"})
				return
			}
			var genres []Genre = []Genre{}
			doc.Find(".genres li a").Each(func(i int, s *goquery.Selection) {
				url, _ := s.Attr("href")
				id := extractAnimeId(url)
				genres = append(genres, Genre{Title: s.Text(), GenreId: id, Href: "/anime/genre/" + id, OtakudesuUrl: url})
			})
			c.IndentedJSON(200, APIResponse{Status: "success", Creator: "Asa Mitaka", StatusCode: 200, StatusMessage: "OK", Ok: true, Data: map[string][]Genre{"genreList": genres}, Pagination: nil})
		})
		listHandler := func(c *gin.Context, baseUrl string) {
			pageId := c.DefaultQuery("page", "1")
			doc, err := getHTML(fmt.Sprintf("%s/page/%s", baseUrl, pageId))
			if err != nil {
				c.JSON(500, gin.H{"error": "Gagal scraping"})
				return
			}
			var animeList []AnimeBaseData = []AnimeBaseData{}
			doc.Find(".venz ul > li").Each(func(i int, s *goquery.Selection) {
				url, _ := s.Find(".thumb a").Attr("href")
				id := extractAnimeId(url)
				epInt := extractNumber(s.Find(".epz").Text())
				img, _ := s.Find(".thumbz img").Attr("src")
				animeList = append(animeList, AnimeBaseData{Title: s.Find(".jdlflm").Text(), Poster: img, Episodes: &epInt, Score: strings.TrimSpace(s.Find(".epztipe").Text()), AnimeId: id, Href: "/anime/anime/" + id, OtakudesuUrl: url})
			})
			c.IndentedJSON(200, APIResponse{Status: "success", Creator: "Asa Mitaka", StatusCode: 200, StatusMessage: "OK", Ok: true, Data: ListAnimeData{AnimeList: animeList}, Pagination: parsePagination(doc, pageId)})
		}
		anime.GET("/complete-anime", func(c *gin.Context) { listHandler(c, "https://otakudesu.best/complete-anime") })
		anime.GET("/ongoing-anime", func(c *gin.Context) { listHandler(c, "https://otakudesu.best/ongoing-anime") })
		anime.GET("/genre/:genreId", func(c *gin.Context) {
			genreId := c.Param("genreId")
			pageId := c.DefaultQuery("page", "1")
			doc, err := getHTML(fmt.Sprintf("https://otakudesu.best/genres/%s/page/%s/", genreId, pageId))
			if err != nil {
				c.JSON(500, gin.H{"error": "Gagal scraping"})
				return
			}
			var animeList []AnimeBaseData = []AnimeBaseData{}
			doc.Find(".col-anime").Each(func(i int, s *goquery.Selection) {
				url, _ := s.Find(".col-anime-title a").Attr("href")
				id := extractAnimeId(url)
				img, _ := s.Find(".col-anime-cover img").Attr("src")
				var genres []Genre = []Genre{}
				s.Find(".col-anime-genre a").Each(func(j int, ga *goquery.Selection) {
					gUrl, _ := ga.Attr("href")
					gId := extractAnimeId(gUrl)
					genres = append(genres, Genre{Title: ga.Text(), GenreId: gId, Href: "/anime/genre/" + gId, OtakudesuUrl: gUrl})
				})
				epInt := extractNumber(s.Find(".col-anime-eps").Text())
				animeList = append(animeList, AnimeBaseData{Title: s.Find(".col-anime-title a").Text(), Poster: img, Episodes: &epInt, Studios: s.Find(".col-anime-studio").Text(), Score: s.Find(".col-anime-rating").Text(), AnimeId: id, Href: "/anime/anime/" + id, OtakudesuUrl: url, GenreList: genres})
			})
			c.IndentedJSON(200, APIResponse{Status: "success", Creator: "Asa Mitaka", StatusCode: 200, StatusMessage: "OK", Ok: true, Data: ListAnimeData{AnimeList: animeList}, Pagination: parsePagination(doc, pageId)})
		})
		anime.GET("/search/:keyword", func(c *gin.Context) {
			keyword := c.Param("keyword")
			doc, err := getHTML("https://otakudesu.best/?s=" + keyword + "&post_type=anime")
			if err != nil {
				c.JSON(500, gin.H{"error": "Gagal scraping"})
				return
			}
			var animeList []AnimeBaseData = []AnimeBaseData{}
			doc.Find(".chivsrc li").Each(func(i int, s *goquery.Selection) {
				url, _ := s.Find("h2 a").Attr("href")
				id := extractAnimeId(url)
				img, _ := s.Find("img").Attr("src")
				var genres []Genre = []Genre{}
				s.Find(".set a").Each(func(j int, ga *goquery.Selection) {
					gUrl, _ := ga.Attr("href")
					gId := extractAnimeId(gUrl)
					genres = append(genres, Genre{Title: ga.Text(), GenreId: gId, Href: "/anime/genre/" + gId, OtakudesuUrl: gUrl})
				})
				animeList = append(animeList, AnimeBaseData{Title: s.Find("h2 a").Text(), Poster: img, AnimeId: id, Href: "/anime/anime/" + id, OtakudesuUrl: url, GenreList: genres})
			})
			c.IndentedJSON(200, APIResponse{Status: "success", Creator: "Asa Mitaka", StatusCode: 200, StatusMessage: "OK", Ok: true, Data: ListAnimeData{AnimeList: animeList}, Pagination: nil})
		})
		anime.GET("/anime/:animeId", func(c *gin.Context) {
			animeId := c.Param("animeId")
			doc, err := getHTML("https://otakudesu.best/anime/" + animeId)
			if err != nil {
				c.JSON(500, gin.H{"error": "Gagal scraping"})
				return
			}
			var detail DetailAnimeData
			detail.Synopsis.Paragraphs = []string{}
			detail.GenreList = []Genre{}
			detail.EpisodeList = []Episode{}
			doc.Find(".infozingle p").Each(func(i int, s *goquery.Selection) {
				parts := strings.SplitN(s.Text(), ":", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(strings.ToLower(parts[0]))
					val := strings.TrimSpace(parts[1])
					switch key {
					case "judul":
						detail.Title = val
					case "japanese":
						detail.Japanese = val
					case "skor":
						detail.Score = val
					case "produser":
						detail.Producers = val
					case "tipe":
						detail.Type = val
					case "status":
						detail.Status = val
					case "total episode":
						detail.Episodes = extractNumber(val)
					case "durasi":
						detail.Duration = val
					case "tanggal rilis":
						detail.Aired = val
					case "studio":
						detail.Studios = val
					}
				}
			})
			detail.Poster, _ = doc.Find(".fotoanime img").Attr("src")
			doc.Find(".sinopc p").Each(func(i int, s *goquery.Selection) {
				if text := strings.TrimSpace(s.Text()); text != "" {
					detail.Synopsis.Paragraphs = append(detail.Synopsis.Paragraphs, text)
				}
			})
			doc.Find(".infozingle p:contains('Genre') a").Each(func(i int, s *goquery.Selection) {
				url, _ := s.Attr("href")
				id := extractAnimeId(url)
				detail.GenreList = append(detail.GenreList, Genre{Title: s.Text(), GenreId: id, Href: "/anime/genre/" + id, OtakudesuUrl: url})
			})
			doc.Find(".episodelist ul li").Each(func(i int, s *goquery.Selection) {
				url, _ := s.Find("a").Attr("href")
				id := extractAnimeId(url)
				title := s.Find("a").Text()
				detail.EpisodeList = append(detail.EpisodeList, Episode{Title: title, Eps: extractEpisodeNumber(title), Date: s.Find(".zeebr").Text(), EpisodeId: id, Href: "/anime/episode/" + id, OtakudesuUrl: url})
			})
			c.IndentedJSON(200, APIResponse{Status: "success", Creator: "Asa Mitaka", StatusCode: 200, StatusMessage: "OK", Ok: true, Data: detail, Pagination: nil})
		})
		anime.GET("/episode/:episodeId", func(c *gin.Context) {
			episodeId := c.Param("episodeId")
			doc, err := getHTML("https://otakudesu.best/episode/" + episodeId)
			if err != nil {
				c.JSON(500, gin.H{"error": "Gagal scraping"})
				return
			}
			var epsData EpisodeData
			epsData.Title = doc.Find(".posttl").Text()
			epsData.AnimeId = episodeId
			epsData.ReleaseTime = strings.TrimSpace(strings.ReplaceAll(doc.Find(".kategoz span:contains('Release on')").Text(), "Release on ", ""))
			epsData.DefaultStreamingUrl, _ = doc.Find("#lightsVideo iframe").Attr("src")
			if epsData.DefaultStreamingUrl == "" {
				epsData.DefaultStreamingUrl, _ = doc.Find(".responsive-embed iframe").Attr("src")
			}
			doc.Find(".flir a").Each(func(i int, s *goquery.Selection) {
				text := strings.ToLower(s.Text())
				url, _ := s.Attr("href")
				id := extractAnimeId(url)
				if strings.Contains(text, "previous") {
					epsData.HasPrevEpisode = true
					epsData.PrevEpisode = &NavEpisode{Title: "Prev", EpisodeId: id, Href: "/anime/episode/" + id, OtakudesuUrl: url}
				} else if strings.Contains(text, "next") {
					epsData.HasNextEpisode = true
					epsData.NextEpisode = &NavEpisode{Title: "Next", EpisodeId: id, Href: "/anime/episode/" + id, OtakudesuUrl: url}
				}
			})
			epsData.Server.Qualities = []ServerQuality{}
			doc.Find(".mirrorstream ul").Each(func(i int, ul *goquery.Selection) {
				if qualClass, _ := ul.Attr("class"); qualClass != "" {
					qualTitle := strings.ReplaceAll(qualClass, "m", "")
					var serverList []ServerLink = []ServerLink{}
					ul.Find("li a").Each(func(j int, a *goquery.Selection) {
						serverId, _ := a.Attr("data-content")
						if serverId == "" {
							serverId, _ = a.Attr("data-id")
						}
						if serverId == "" {
							hrefVal, _ := a.Attr("href")
							if hrefVal != "#" && !strings.Contains(strings.ToLower(hrefVal), "javascript") {
								serverId = hrefVal
							}
						}
						if serverId != "" {
							serverList = append(serverList, ServerLink{Title: strings.TrimSpace(a.Text()), ServerId: serverId, Href: "/anime/server/" + serverId})
						}
					})
					if len(serverList) > 0 {
						epsData.Server.Qualities = append(epsData.Server.Qualities, ServerQuality{Title: qualTitle, ServerList: serverList})
					}
				}
			})
			epsData.DownloadUrl.Qualities = []DownloadQuality{}
			doc.Find(".download ul li").Each(func(i int, s *goquery.Selection) {
				qual := DownloadQuality{Title: s.Find("strong").Text(), Size: s.Find("i").Text(), Urls: []DownloadLink{}}
				s.Find("a").Each(func(j int, a *goquery.Selection) {
					url, _ := a.Attr("href")
					qual.Urls = append(qual.Urls, DownloadLink{Title: strings.TrimSpace(a.Text()), Url: url})
				})
				if len(qual.Urls) > 0 {
					epsData.DownloadUrl.Qualities = append(epsData.DownloadUrl.Qualities, qual)
				}
			})
			epsData.Info.GenreList = []Genre{}
			epsData.Info.EpisodeList = []Episode{}
			doc.Find(".infozingle p").Each(func(i int, s *goquery.Selection) {
				parts := strings.SplitN(s.Text(), ":", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(strings.ToLower(parts[0]))
					val := strings.TrimSpace(parts[1])
					if key == "credit" {
						epsData.Info.Credit = val
					} else if key == "encoder" {
						epsData.Info.Encoder = val
					} else if key == "duration" {
						epsData.Info.Duration = val
					} else if key == "tipe" {
						epsData.Info.Type = val
					}
				}
			})
			doc.Find(".infozingle p:contains('Genres') a").Each(func(i int, s *goquery.Selection) {
				url, _ := s.Attr("href")
				id := extractAnimeId(url)
				epsData.Info.GenreList = append(epsData.Info.GenreList, Genre{Title: s.Text(), GenreId: id, Href: "/anime/genre/" + id, OtakudesuUrl: url})
			})
			doc.Find(".keyingpost li").Each(func(i int, s *goquery.Selection) {
				a := s.Find("a")
				url, _ := a.Attr("href")
				id := extractAnimeId(url)
				title := a.Text()
				epsData.Info.EpisodeList = append(epsData.Info.EpisodeList, Episode{Title: title, Eps: extractEpisodeNumber(title), EpisodeId: id, Href: "/anime/episode/" + id, OtakudesuUrl: url})
			})
			c.IndentedJSON(200, APIResponse{Status: "success", Creator: "Asa Mitaka", StatusCode: 200, StatusMessage: "OK", Ok: true, Data: epsData, Pagination: nil})
		})
		anime.GET("/server/:serverId", func(c *gin.Context) {
			serverId := c.Param("serverId")
			padding := len(serverId) % 4
			if padding > 0 {
				serverId += strings.Repeat("=", 4-padding)
			}
			decodedPayload, err := base64.StdEncoding.DecodeString(serverId)
			if err != nil {
				c.JSON(400, gin.H{"error": "Format serverId tidak valid"})
				return
			}
			var payloadData map[string]interface{}
			if err := json.Unmarshal(decodedPayload, &payloadData); err != nil {
				c.JSON(400, gin.H{"error": "Gagal membaca data JSON serverId"})
				return
			}
			ajaxUrl := "https://otakudesu.best/wp-admin/admin-ajax.php"
			nonceResp, err := http.PostForm(ajaxUrl, url.Values{"action": {"aa1208d27f29ca340c92c66d1926f13f"}})
			if err != nil {
				c.JSON(500, gin.H{"error": "Gagal terhubung ke AJAX"})
				return
			}
			defer nonceResp.Body.Close()
			var nonceResult struct {
				Data string `json:"data"`
			}
			json.NewDecoder(nonceResp.Body).Decode(&nonceResult)
			if nonceResult.Data == "" {
				c.JSON(500, gin.H{"error": "Gagal mendapatkan nonce"})
				return
			}
			iframeForm := url.Values{"nonce": {nonceResult.Data}, "action": {"2a3505c93b0035d3f455df82bf976b84"}}
			for k, v := range payloadData {
				iframeForm.Add(k, fmt.Sprintf("%v", v))
			}
			iframeResp, err := http.PostForm(ajaxUrl, iframeForm)
			if err != nil {
				c.JSON(500, gin.H{"error": "Gagal mengambil data video"})
				return
			}
			defer iframeResp.Body.Close()
			var iframeResult struct {
				Data string `json:"data"`
			}
			json.NewDecoder(iframeResp.Body).Decode(&iframeResult)
			decodedHtml, err := base64.StdEncoding.DecodeString(iframeResult.Data)
			if err != nil {
				c.JSON(500, gin.H{"error": "Gagal decode HTML Iframe"})
				return
			}
			docHTML, _ := goquery.NewDocumentFromReader(strings.NewReader(string(decodedHtml)))
			iframeSrc, _ := docHTML.Find("iframe").Attr("src")
			c.IndentedJSON(200, APIResponse{Status: "success", Creator: "Asa Mitaka", StatusCode: 200, StatusMessage: "OK", Ok: true, Data: map[string]string{"url": iframeSrc}, Pagination: nil})
		})
	}
	log.Println("Server berjalan di port 80 (http://localhost)")
	r.Run(":8080")
}
