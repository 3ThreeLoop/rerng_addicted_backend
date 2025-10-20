package scraping

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	custom_log "rerng_addicted_api/pkg/logs"
	types "rerng_addicted_api/pkg/model"
	"rerng_addicted_api/pkg/responses"
	"rerng_addicted_api/pkg/utils"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/jmoiron/sqlx"
	"github.com/schollz/progressbar/v3"
)

type ScrapingRepo interface {
	Search(keyword string) (*SeriesResponse, *responses.ErrorResponse)
	GetDetail(key string) (*SeriesDetailsResponse, *responses.ErrorResponse)
	GetDeepDetail(key string) (*SeriesDeepDetailsResponse, *responses.ErrorResponse)
}

type ScrapingRepoImpl struct {
	DBPool      *sqlx.DB
	UserContext *types.UserContext
}

func NewScrapingRepoImpl(db_pool *sqlx.DB, user_context *types.UserContext) *ScrapingRepoImpl {
	return &ScrapingRepoImpl{
		DBPool:      db_pool,
		UserContext: user_context,
	}
}

func (sc *ScrapingRepoImpl) Search(keyword string) (*SeriesResponse, *responses.ErrorResponse) {
	// visit main page to get cookies
	main_url := "https://kisskh.co/"
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req_main, _ := http.NewRequest("GET", main_url, nil)
	req_main.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/115.0 Safari/537.36")

	resp_main, err := client.Do(req_main)
	if err != nil {
		custom_log.NewCustomLog("scraping_failed", err.Error(), "error")
		err_msg := &responses.ErrorResponse{}
		return nil, err_msg.NewErrorResponse("scraping_failed", fmt.Errorf("original_source_error"))
	}
	defer resp_main.Body.Close()

	// collect cookies
	cookies := resp_main.Cookies()
	cookie_str := ""
	for _, ck := range cookies {
		cookie_str += ck.Name + "=" + ck.Value + "; "
	}

	// call search API with cookies
	api_url := fmt.Sprintf("https://kisskh.co/api/DramaList/Search?q=%s&type=0", url.QueryEscape(keyword))
	req_api, _ := http.NewRequest("GET", api_url, nil)
	req_api.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/115.0 Safari/537.36")
	req_api.Header.Set("Referer", main_url)
	req_api.Header.Set("Cookie", cookie_str)

	resp_api, err := client.Do(req_api)
	if err != nil {
		custom_log.NewCustomLog("scraping_failed", err.Error(), "error")
		err_msg := &responses.ErrorResponse{}
		return nil, err_msg.NewErrorResponse("scraping_failed", fmt.Errorf("api_fetch_failed"))
	}
	defer resp_api.Body.Close()

	body, _ := io.ReadAll(resp_api.Body)

	// parse JSON response
	var series []Serie
	// fmt.Println("body : ", string(body))
	if err := json.Unmarshal(body, &series); err != nil {
		custom_log.NewCustomLog("scraping_failed", err.Error(), "error")
		err_msg := &responses.ErrorResponse{}
		return nil, err_msg.NewErrorResponse("scraping_failed", fmt.Errorf("parse_data_failed"))
	}

	return &SeriesResponse{
		Series: series,
	}, nil
}

func (sc *ScrapingRepoImpl) GetDetail(key string) (*SeriesDetailsResponse, *responses.ErrorResponse) {
	// visit main page to get cookies
	main_url := "https://kisskh.co/"
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req_main, _ := http.NewRequest("GET", main_url, nil)
	req_main.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/115.0 Safari/537.36")

	resp_main, err := client.Do(req_main)
	if err != nil {
		custom_log.NewCustomLog("scraping_failed", err.Error(), "error")
		err_msg := &responses.ErrorResponse{}
		return nil, err_msg.NewErrorResponse("scraping_failed", fmt.Errorf("original_source_error"))
	}
	defer resp_main.Body.Close()

	// collect cookies from response
	cookies := resp_main.Cookies()
	cookie_str := ""
	for _, ck := range cookies {
		cookie_str += ck.Name + "=" + ck.Value + "; "
	}

	fmt.Println("heloooooo")

	// call search API with cookies
	api_url := fmt.Sprintf("https://kisskh.co/api/DramaList/Drama/%s", key)
	req_api, _ := http.NewRequest("GET", api_url, nil)
	req_api.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/115.0 Safari/537.36")
	req_api.Header.Set("Referer", main_url)
	req_api.Header.Set("Cookie", cookie_str)

	resp_api, err := client.Do(req_api)
	if err != nil {
		custom_log.NewCustomLog("scraping_failed", err.Error(), "error")
		err_msg := &responses.ErrorResponse{}
		return nil, err_msg.NewErrorResponse("scraping_failed", fmt.Errorf("api_fetch_failed"))
	}
	defer resp_api.Body.Close()

	body, _ := io.ReadAll(resp_api.Body)

	// parse JSON response
	var serie_detail SerieDetail
	fmt.Println("body : ", string(body))
	if err := json.Unmarshal(body, &serie_detail); err != nil {
		custom_log.NewCustomLog("scraping_failed", err.Error(), "error")
		err_msg := &responses.ErrorResponse{}
		return nil, err_msg.NewErrorResponse("scraping_failed", fmt.Errorf("parse_data_failed"))
	}

	return &SeriesDetailsResponse{
		SeriesDetails: []SerieDetail{
			serie_detail,
		},
	}, nil
}

func (sc *ScrapingRepoImpl) GetDeepDetail(key string) (*SeriesDeepDetailsResponse, *responses.ErrorResponse) {
	// get main page cookies
	// go to original website to get cookies
	main_url := "https://kisskh.co/"
	client := &http.Client{}
	req_main, _ := http.NewRequest("GET", main_url, nil)
	req_main.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/115.0 Safari/537.36")

	resp_main, err := client.Do(req_main)
	if err != nil {
		custom_log.NewCustomLog("scraping_failed", err.Error(), "error")
		err_msg := &responses.ErrorResponse{}
		return nil, err_msg.NewErrorResponse("scraping_failed", fmt.Errorf("original_source_error"))
	}
	defer resp_main.Body.Close()

	// get cookies from response
	cookies := resp_main.Cookies()
	cookie_str := ""
	for _, ck := range cookies {
		cookie_str += ck.Name + "=" + ck.Value + "; "
	}

	// get drama details via API
	api_url := fmt.Sprintf("https://kisskh.co/api/DramaList/Drama/%s", key)
	req_api, _ := http.NewRequest("GET", api_url, nil)
	req_api.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/115.0 Safari/537.36")
	req_api.Header.Set("Referer", main_url)
	req_api.Header.Set("Cookie", cookie_str)

	resp_api, err := client.Do(req_api)
	if err != nil {
		custom_log.NewCustomLog("scraping_failed", err.Error(), "error")
		err_msg := &responses.ErrorResponse{}
		return nil, err_msg.NewErrorResponse("scraping_failed", fmt.Errorf("fetch_api_failed"))
	}
	defer resp_api.Body.Close()

	// read response body to unmarshal response body to the struct
	body, _ := io.ReadAll(resp_api.Body)
	var serie_deep_detail SerieDeepDetail
	if err := json.Unmarshal(body, &serie_deep_detail); err != nil {
		custom_log.NewCustomLog("scraping_failed", err.Error(), "error")
		err_msg := &responses.ErrorResponse{}
		return nil, err_msg.NewErrorResponse("scraping_failed", fmt.Errorf("parse_data_failed"))
	}

	// launch Rod browser
	// use browser to inspect network work from episode link
	l := launcher.New().
		Headless(true).
		NoSandbox(true).
		MustLaunch()

	browser := rod.New().ControlURL(l).MustConnect()
	defer browser.MustClose()

	// parallel scraping with worker pool with limit concurrency
	var wg sync.WaitGroup
	sem := make(chan struct{}, 5)

	for i := range serie_deep_detail.Episodes {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			ep := serie_deep_detail.Episodes[i]
			fmt.Println("===============================================================")
			fmt.Println("Processing Episode:", ep.Number)

			page := browser.MustPage()
			_ = proto.NetworkSetCacheDisabled{CacheDisabled: true}.Call(page)
			_ = proto.NetworkEnable{}.Call(page)

			epURL := fmt.Sprintf(
				"https://kisskh.co/Drama/%s/Episode-%d?id=%d&ep=%d&page=0&pageSize=100",
				slugify(serie_deep_detail.Title), int(ep.Number), serie_deep_detail.ID, ep.ID,
			)

			fmt.Println("Navigating to:", epURL)
			page.MustNavigate(epURL).MustWaitLoad()

			deadline := time.After(25 * time.Second)

		loop: // 👈 label to break outer loop
			for {
				evt := &proto.NetworkResponseReceived{}
				wait := page.WaitEvent(evt)

				done := make(chan struct{})
				go func() {
					wait()
					done <- struct{}{}
				}()

				select {
				case <-done:
					url_str := evt.Response.URL
					mime := evt.Response.MIMEType

					// prepare proxy base
					host := os.Getenv("API_HOST")
					port := utils.GetenvInt("API_PORT", 8585)
					proxy_base := fmt.Sprintf("http://%s:%d", host, port)

					// capture video
					if strings.Contains(url_str, ".m3u8") || mime == "application/vnd.apple.mpegurl" ||
						strings.Contains(url_str, ".mp4") || mime == "video/mp4" {

						if strings.Contains(url_str, ".m3u8") || mime == "application/vnd.apple.mpegurl" {
							trimmed := strings.TrimPrefix(url_str, "https://")
							trimmed = strings.TrimPrefix(trimmed, "http://")
							serie_deep_detail.Episodes[i].Source = fmt.Sprintf("%s/m3u8/%s", proxy_base, trimmed)
						} else if strings.Contains(url_str, ".mp4") || mime == "video/mp4" {
							encoded := url.QueryEscape(url_str)
							serie_deep_detail.Episodes[i].Source = fmt.Sprintf("%s/mp4?url=%s", proxy_base, encoded)
						}

						fmt.Printf("✅ Found video for ep %.0f: %s\n", ep.Number, serie_deep_detail.Episodes[i].Source)
					}

					// capture subtitles
					sub_url_contains := fmt.Sprintf("/api/Sub/%d", ep.ID)
					if strings.Contains(url_str, sub_url_contains) {
						req_sub, _ := http.NewRequest("GET", url_str, nil)
						req_sub.Header.Set("User-Agent", "Mozilla/5.0")
						req_sub.Header.Set("Referer", epURL)
						req_sub.Header.Set("Cookie", cookie_str)

						if resp_sub, err := client.Do(req_sub); err == nil {
							defer resp_sub.Body.Close()
							subBody, _ := io.ReadAll(resp_sub.Body)

							var subs []Subtitle
							if err := json.Unmarshal(subBody, &subs); err == nil {
								for j, sub := range subs {
									trimmed := strings.TrimPrefix(sub.Src, "https://")
									trimmed = strings.TrimPrefix(trimmed, "http://")
									subs[j].Src = fmt.Sprintf("%s/subtitle/%s", proxy_base, trimmed)
								}

								serie_deep_detail.Episodes[i].Subtitles = subs
								fmt.Printf("✅ Parsed %d subtitles for ep %.0f\n", len(subs), ep.Number)
							}
						}
					}

					// exit when both found
					if serie_deep_detail.Episodes[i].Source != "" && len(serie_deep_detail.Episodes[i].Subtitles) > 0 {
						break loop
					}

				case <-deadline:
					fmt.Println("❌ Timeout for episode", ep.Number)
					break loop
				}
			}

			page.MustClose()
		}(i)
	}

	wg.Wait()

	return &SeriesDeepDetailsResponse{
		SeriesDeepDetails: []SerieDeepDetail{
			serie_deep_detail,
		},
	}, nil
}

func (sc *ScrapingRepoImpl) Seed() {
	series_key := []int{
		// korean drama
		975, 10124, 3749, 7555, 4596, //41, 5043, 5941, 124, 126,

		// chinese drama
		//10826, 8316, 6158, 9148, 773, 9579, 7949, 7628, 10958, 7532,

		// hollywood
		//11698, 11764, 11694, 11739, 11521, 11526, 11706, 11705, 11674, 11652,
	}

	var series_deep_detail []SerieDeepDetail

	// 🌱 Initialize progress bar
	bar := progressbar.NewOptions(len(series_key),
		progressbar.OptionSetDescription("Scraping series..."),
		progressbar.OptionShowCount(),
		progressbar.OptionSetWidth(30),
		progressbar.OptionShowElapsedTimeOnFinish(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "█",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)

	start := time.Now()
	success := 0
	fail := 0

	for _, key := range series_key {
		resp, err := sc.GetDeepDetail(fmt.Sprintf("%d", key))
		if err != nil {
			color.Red("❌ Failed scraping series %d: %v", key, err)
			fail++
			bar.Add(1)
			continue
		}
		success++
		bar.Add(1)
		color.Green("✅ Scraped series ID: %d (%d episodes)", key, len(resp.SeriesDeepDetails))
		series_deep_detail = append(series_deep_detail, resp.SeriesDeepDetails...)
	}

	// Begin DB transaction
	tx, err := sc.DBPool.Beginx()
	if err != nil {
		color.Red("❌ Failed to start transaction: %v", err)
		return
	}

	color.Yellow("\n💾 Inserting data into database...")

	insertBar := progressbar.NewOptions(len(series_deep_detail),
		progressbar.OptionSetDescription("Inserting series..."),
		progressbar.OptionSetWidth(30),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "█",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)

	for _, serie := range series_deep_detail {
		_, err := tx.NamedExec(`
			INSERT INTO series 
				(id, title, description, release_date, trailer, country, status, type, next_ep_date_id,
				episodes_count, label, favorite_id, thumbnail)
			VALUES 
				(:id, :title, :description, :release_date, :trailer, :country, :status, :type, :next_ep_date_id,
				:episodes_count, :label, :favorite_id, :thumbnail)
			ON CONFLICT (id) DO NOTHING
		`, map[string]interface{}{
			"id":              serie.ID,
			"title":           serie.Title,
			"description":     serie.Description,
			"release_date":    serie.ReleaseDate,
			"trailer":         serie.Trailer,
			"country":         serie.Country,
			"status":          serie.Status,
			"type":            serie.Type,
			"next_ep_date_id": serie.NextEpDateID,
			"episodes_count":  serie.EpisodesCount,
			"label":           serie.Label,
			"favorite_id":     serie.FavoriteID,
			"thumbnail":       serie.Thumbnail,
		})

		if err != nil {
			color.Red("❌ [line 430] Failed to insert series %d: %v", serie.ID, err)
			tx.Rollback()
			return
		}

		for _, ep := range serie.Episodes {
			_, err := tx.NamedExec(`
				INSERT INTO episodes (id, series_id, number, sub, src)
					VALUES (:id, :series_id, :number, :sub, :src)
				ON CONFLICT (id) DO NOTHING
			`, map[string]interface{}{
				"id":        ep.ID,
				"series_id": serie.ID,
				"number":    ep.Number,
				"sub":       ep.Sub,
				"src":       ep.Source,
			})
			if err != nil {
				color.Red("❌[line 448] Failed inserting episode %d of series %d: %v", ep.ID, serie.ID, err)
				tx.Rollback()
				return
			}

			for _, sub := range ep.Subtitles {
				_, err := tx.NamedExec(`
					INSERT INTO subtitles (episode_id, src, label, lang, is_default)
					VALUES (:episode_id, :src, :label, :lang, :is_default)
				`, map[string]interface{}{
					"episode_id": ep.ID,
					"src":        sub.Src,
					"label":      sub.Label,
					"lang":       sub.Lang,
					"is_default": sub.Default,
				})
				if err != nil {
					color.Red("❌ [line 465] Failed inserting subtitle for episode %d: %v", ep.ID, err)
					tx.Rollback()
					return
				}
			}
		}
		insertBar.Add(1)
	}

	if err := tx.Commit(); err != nil {
		color.Red("❌ Transaction commit failed: %v", err)
		return
	}

	duration := time.Since(start)
	color.Cyan("\n🎬 Seeding complete in %v", duration)
	color.Green("✅ Scraped: %d | Failed: %d", success, fail)
	color.Yellow("💾 Series inserted: %d", len(series_deep_detail))
}

func slugify(title string) string {
	slug := strings.TrimSpace(title)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "(", "-")
	slug = strings.ReplaceAll(slug, ")", "-")

	// Collapse multiple dashes
	re := regexp.MustCompile(`-+`)
	slug = re.ReplaceAllString(slug, "-")

	return slug
}

// for khfullhd
// func (au *AuthHandler) ScrapingDetail(c *fiber.Ctx) error {
// 	targetURL := c.Query("url")
// 	if targetURL == "" {
// 		return c.Status(http.StatusBadRequest).SendString("missing url")
// 	}

// 	ctx, cancel := context.WithTimeout(c.Context(), 90*time.Second)
// 	defer cancel()

// 	result := make(map[string]interface{})

// 	mainCollector := colly.NewCollector(
// 		colly.Async(true),
// 		colly.MaxDepth(2),
// 	)
// 	mainCollector.SetRequestTimeout(10 * time.Second)

// 	// Debug
// 	mainCollector.OnRequest(func(r *colly.Request) {
// 		fmt.Println("Visiting:", r.URL.String())
// 	})

// 	// Cover
// 	mainCollector.OnHTML("#content-cover", func(e *colly.HTMLElement) {
// 		bg := e.Attr("style")
// 		re := regexp.MustCompile(`url\((.*?)\)`)
// 		if match := re.FindStringSubmatch(bg); len(match) > 1 {
// 			result["cover"] = match[1]
// 		}
// 	})

// 	// Poster
// 	mainCollector.OnHTML(".thumb.mvic-thumb img", func(e *colly.HTMLElement) {
// 		result["poster"] = e.Attr("src")
// 	})

// 	// Description
// 	mainCollector.OnHTML(".mvic-desc .desc", func(e *colly.HTMLElement) {
// 		result["description"] = strings.TrimSpace(e.Text)
// 	})

// 	// Genres + Actors
// 	mainCollector.OnHTML(".mvici-left p", func(e *colly.HTMLElement) {
// 		label := strings.TrimSpace(e.ChildText("strong"))

// 		if strings.Contains(label, "Genre") {
// 			var genres []string
// 			e.ForEach("a", func(_ int, el *colly.HTMLElement) {
// 				genres = append(genres, strings.TrimSpace(el.Text))
// 			})
// 			result["genres"] = genres
// 		}

// 		if strings.Contains(label, "Actors") {
// 			var actors []string
// 			e.ForEach("a", func(_ int, el *colly.HTMLElement) {
// 				actors = append(actors, strings.TrimSpace(el.Text))
// 			})
// 			result["actors"] = actors
// 		}
// 	})

// 	// Studio
// 	mainCollector.OnHTML(".mvici-left p:contains('Studio') a", func(e *colly.HTMLElement) {
// 		result["studio"] = e.Text
// 	})

// 	// Duration
// 	mainCollector.OnHTML(".mvici-right p:contains('Duration') span", func(e *colly.HTMLElement) {
// 		result["duration"] = e.Text
// 	})

// 	// Release
// 	mainCollector.OnHTML(".mvici-right p:contains('Release') a", func(e *colly.HTMLElement) {
// 		result["release"] = e.Text
// 	})

// 	// Seasons + Episodes
// 	var seasons []map[string]interface{}
// 	mainCollector.OnHTML(".tvseason", func(e *colly.HTMLElement) {
// 		season := map[string]interface{}{
// 			"title":    e.ChildText(".les-title strong"),
// 			"episodes": []map[string]interface{}{},
// 		}

// 		e.ForEach(".les-content a", func(_ int, ep *colly.HTMLElement) {
// 			epData := map[string]interface{}{
// 				"title":    strings.TrimSpace(ep.Text),
// 				"link":     ep.Attr("href"),
// 				"iframe":   "",
// 				"playlist": []map[string]interface{}{},
// 			}

// 			// Get iframe link from episode page
// 			epCollector := colly.NewCollector()
// 			epCollector.OnHTML(".movieplay iframe", func(iframe *colly.HTMLElement) {
// 				epData["iframe"] = iframe.Attr("src")
// 			})
// 			epCollector.Visit(ep.Attr("href"))

// 			// Inside your ScrapingDetail function, when fetching the iframe:
// 			if iframeURL, ok := epData["iframe"].(string); ok && iframeURL != "" {
// 				// Pass the episode page URL as a query parameter to your proxy
// 				proxyURL := fmt.Sprintf("%s/proxy/%s?referer=%s",
// 					c.BaseURL(),
// 					strings.TrimPrefix(iframeURL, "https://stream.khfullhd.co/"),
// 					url.QueryEscape(ep.Attr("href")), // correct usage
// 				)

// 				resp, err := http.Get(proxyURL)
// 				if err != nil {
// 					fmt.Println("proxy request error:", err)
// 					return
// 				}
// 				defer resp.Body.Close()

// 				bodyBytes, _ := io.ReadAll(resp.Body)
// 				body := string(bodyBytes)
// 				fmt.Println("response body : ", body)

// 				// Parse JWPlayer playlist
// 				re := regexp.MustCompile(`var\s+playlist\s*=\s*(\[.*?\]);`)
// 				match := re.FindStringSubmatch(body)
// 				if len(match) > 1 {
// 					var playlist []map[string]interface{}
// 					if err := json.Unmarshal([]byte(match[1]), &playlist); err == nil && len(playlist) > 0 {
// 						var enriched []map[string]interface{}
// 						for _, p := range playlist {
// 							epBlock := map[string]interface{}{
// 								"thumbnail": p["image"],
// 							}

// 							// Rewrite HLS and subtitle URLs to go through your proxy
// 							if sources, ok := p["sources"].([]interface{}); ok {
// 								var videoSources []map[string]interface{}
// 								for _, s := range sources {
// 									if src, ok := s.(map[string]interface{}); ok {
// 										file := src["file"].(string)
// 										if strings.HasPrefix(file, "https://stream.khanime.co/") {
// 											// Fetch the master m3u8 playlist
// 											resp, err := http.Get(file)
// 											if err != nil {
// 												continue
// 											}
// 											body, _ := io.ReadAll(resp.Body)
// 											resp.Body.Close()

// 											// Parse resolutions from m3u8
// 											reRes := regexp.MustCompile(`#EXT-X-STREAM-INF:.*RESOLUTION=(\d+x\d+)\s+([^\s]+)`)
// 											matches := reRes.FindAllStringSubmatch(string(body), -1)

// 											if len(matches) > 0 {
// 												// Build sources for each resolution
// 												for _, m := range matches {
// 													resLabel := m[1] // "1280x720"
// 													resURL := m[2]   // "/w/hlsplaylist/10583/.../720"
// 													if strings.HasPrefix(resURL, "/") {
// 														resURL = c.BaseURL() + "/video-proxy" + resURL + "?referer=" + url.QueryEscape(ep.Attr("href"))
// 													}
// 													videoSources = append(videoSources, map[string]interface{}{
// 														"file":    resURL,
// 														"label":   resLabel,
// 														"default": resLabel == "1280x720", // optional default
// 														"type":    "hls",
// 													})
// 												}
// 											} else {
// 												// fallback if no sub-playlists, just use original file
// 												videoSources = append(videoSources, map[string]interface{}{
// 													"file":    c.BaseURL() + "/video-proxy/" + strings.TrimPrefix(file, "https://stream.khanime.co/") + "?referer=" + url.QueryEscape(ep.Attr("href")),
// 													"label":   "auto",
// 													"default": true,
// 													"type":    "hls",
// 												})
// 											}
// 										}
// 									}
// 								}
// 								epBlock["sources"] = videoSources
// 							}

// 							if tracks, ok := p["tracks"].([]interface{}); ok {
// 								var subtitles []map[string]interface{}
// 								for _, t := range tracks {
// 									if tr, ok := t.(map[string]interface{}); ok {
// 										file := tr["file"].(string)
// 										if strings.HasPrefix(file, "https://stream.khanime.co/") {
// 											file = strings.Replace(file, "https://stream.khanime.co/", c.BaseURL()+"/video-proxy/", 1)
// 										}
// 										subtitles = append(subtitles, map[string]interface{}{
// 											"file":  file,
// 											"label": tr["label"],
// 											"kind":  tr["kind"],
// 										})
// 									}
// 								}
// 								epBlock["subtitles"] = subtitles
// 							}

// 							enriched = append(enriched, epBlock)
// 						}
// 						epData["playlist"] = enriched
// 					}
// 				}
// 			}

// 			season["episodes"] = append(season["episodes"].([]map[string]interface{}), epData)
// 		})

// 		seasons = append(seasons, season)
// 	})

// 	mainCollector.Visit(targetURL)
// 	mainCollector.Wait()
// 	result["seasons"] = seasons

// 	select {
// 	case <-ctx.Done():
// 		return c.Status(http.StatusRequestTimeout).SendString("scraping timed out")
// 	default:
// 		return c.JSON(result)
// 	}
// }

// for khfullhd
// func (au *AuthHandler) ScrapingSearch(c *fiber.Ctx) error {
// 	keyword := c.Query("q")
// 	if keyword == "" {
// 		return c.Status(http.StatusBadRequest).SendString("missing keyword")
// 	}

// 	var results []map[string]string
// 	scraper := colly.NewCollector()

// 	// Each movie/series item
// 	scraper.OnHTML(".movies-list-full .ml-item", func(e *colly.HTMLElement) {
// 		link := e.ChildAttr("a.ml-mask", "href")
// 		title := e.ChildText("h2")
// 		img := e.ChildAttr("img", "data-original")
// 		if img == "" {
// 			img = e.ChildAttr("img", "src")
// 		}

// 		results = append(results, map[string]string{
// 			"link":  link,
// 			"title": title,
// 			"image": img,
// 		})
// 	})

// 	searchURL := "https://khfullhd.co/?" + url.Values{
// 		"s": {keyword},
// 	}.Encode()

// 	if err := scraper.Visit(searchURL); err != nil {
// 		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
// 	}

// 	return c.JSON(results)
// }

// func (au *AuthHandler) Scraping(c *fiber.Ctx) error {
// 	var baseURL = c.Query("url")
// 	if baseURL == "" {
// 		return c.Status(http.StatusBadRequest).SendString("missing url")
// 	}

// 	var results []map[string]string

// 	episodeListScraper := colly.NewCollector()

// 	episodePageScraper := episodeListScraper.Clone()

// 	episodePageScraper.OnHTML("div.watch_video.watch-iframe iframe", func(e *colly.HTMLElement) {
// 		videoSrc := e.Attr("src")

// 		if strings.HasPrefix(videoSrc, "//") {
// 			videoSrc = "https:" + videoSrc
// 		}

// 		results = append(results, map[string]string{
// 			"title":     e.Request.Ctx.Get("title"),
// 			"link":      e.Request.URL.String(),
// 			"video_src": videoSrc,
// 		})
// 	})

// 	episodeListScraper.OnHTML("ul.list-episode-item-2.all-episode li a", func(e *colly.HTMLElement) {
// 		epLink := e.Request.AbsoluteURL(e.Attr("href"))
// 		epTitle := strings.TrimSpace(e.ChildText("h3.title"))

// 		ctx := colly.NewContext()
// 		ctx.Put("title", epTitle)
// 		episodePageScraper.Request("GET", epLink, nil, ctx, nil)
// 	})

// 	if err := episodeListScraper.Visit(baseURL); err != nil {
// 		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
// 	}

// 	return c.Status(http.StatusOK).JSON(results)
// }
