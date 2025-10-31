package scraping

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"rerng_addicted_api/internal/admin/serie"
	custom_log "rerng_addicted_api/pkg/logs"
	types "rerng_addicted_api/pkg/model"
	"rerng_addicted_api/pkg/responses"
	"rerng_addicted_api/pkg/utils"
	"strings"
	"sync"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/jmoiron/sqlx"
)

type ScrapingRepo interface {
	Search(keyword string) (*SeriesResponse, *responses.ErrorResponse)
	ViewDetail(key string) (*SeriesDetailsResponse, *responses.ErrorResponse)
	GetDeepDetail(key string) (*SeriesDeepDetailsResponse, *responses.ErrorResponse)
	GetEpisodes(key int, ep_num int) (*EpisodesResponse, *responses.ErrorResponse)
	Seed() []serie.SerieDeepDetail
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
	var series_json []serie.SerieJSON
	// fmt.Println("body : ", string(body))
	if err := json.Unmarshal(body, &series_json); err != nil {
		custom_log.NewCustomLog("scraping_failed", err.Error(), "error")
		err_msg := &responses.ErrorResponse{}
		return nil, err_msg.NewErrorResponse("scraping_failed", fmt.Errorf("parse_data_failed"))
	}

	series := make([]serie.Serie, len(series_json))
	for i, s := range series_json {
		series[i] = serie.Serie{
			ID:            s.ID,
			Title:         s.Title,
			EpisodesCount: s.EpisodesCount,
			Label:         s.Label,
			FavoriteID:    s.FavoriteID,
			Thumbnail:     s.Thumbnail,
		}
	}

	return &SeriesResponse{
		Series: series,
	}, nil
}

func (sc *ScrapingRepoImpl) ViewDetail(key string) (*SeriesDetailsResponse, *responses.ErrorResponse) {
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
	var serie_detail_json serie.SerieDetailJSON
	// fmt.Println("body : ", string(body))
	if err := json.Unmarshal(body, &serie_detail_json); err != nil {
		custom_log.NewCustomLog("scraping_failed", err.Error(), "error")
		err_msg := &responses.ErrorResponse{}
		return nil, err_msg.NewErrorResponse("scraping_failed", fmt.Errorf("parse_data_failed"))
	}

	episodes := make([]serie.Episode, len(serie_detail_json.Episodes))
	for i, ep := range serie_detail_json.Episodes {
		episodes[i] = serie.Episode{
			ID:     ep.ID,
			Number: ep.Number,
			Sub:    ep.Sub,
		}
	}

	serie_detail := serie.SerieDetail{
		ID:            serie_detail_json.ID,
		Title:         serie_detail_json.Title,
		Description:   serie_detail_json.Description,
		ReleaseDate:   serie_detail_json.ReleaseDate,
		Trailer:       serie_detail_json.Trailer,
		Country:       serie_detail_json.Country,
		Status:        serie_detail_json.Status,
		Type:          serie_detail_json.Type,
		NextEpDateID:  serie_detail_json.NextEpDateID,
		Episodes:      episodes,
		EpisodesCount: serie_detail_json.EpisodesCount,
		Label:         serie_detail_json.Label,
		FavoriteID:    serie_detail_json.FavoriteID,
		Thumbnail:     serie_detail_json.Thumbnail,
	}

	return &SeriesDetailsResponse{
		SeriesDetails: []serie.SerieDetail{
			serie_detail,
		},
	}, nil
}

func (sc *ScrapingRepoImpl) GetDetail(key string) (*SeriesDeepDetailsResponse, *responses.ErrorResponse) {
	main_url := "https://kisskh.co/"
	client := &http.Client{Timeout: 15 * time.Second}

	// get cookies
	req_main, _ := http.NewRequest("GET", main_url, nil)
	req_main.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/115.0 Safari/537.36")

	resp_main, err := client.Do(req_main)
	if err != nil {
		custom_log.NewCustomLog("scraping_failed", err.Error(), "error")
		return nil, (&responses.ErrorResponse{}).NewErrorResponse("scraping_failed", fmt.Errorf("original_source_error"))
	}
	defer resp_main.Body.Close()

	var cookie_str strings.Builder
	for _, ck := range resp_main.Cookies() {
		cookie_str.WriteString(fmt.Sprintf("%s=%s; ", ck.Name, ck.Value))
	}

	// fetch series info
	apiURL := fmt.Sprintf("https://kisskh.co/api/DramaList/Drama/%s", key)
	req_api, _ := http.NewRequest("GET", apiURL, nil)
	req_api.Header.Set("User-Agent", req_main.Header.Get("User-Agent"))
	req_api.Header.Set("Referer", main_url)
	req_api.Header.Set("Cookie", cookie_str.String())

	resp_api, err := client.Do(req_api)
	if err != nil {
		custom_log.NewCustomLog("scraping_failed", err.Error(), "error")
		return nil, (&responses.ErrorResponse{}).NewErrorResponse("scraping_failed", fmt.Errorf("fetch_api_failed"))
	}
	defer resp_api.Body.Close()

	body, _ := io.ReadAll(resp_api.Body)
	var serie_detail_json serie.SerieDeepDetailJSON
	if err := json.Unmarshal(body, &serie_detail_json); err != nil {
		custom_log.NewCustomLog("scraping_failed", err.Error(), "error")
		return nil, (&responses.ErrorResponse{}).NewErrorResponse("scraping_failed", fmt.Errorf("parse_data_failed"))
	}

	episodes := make([]serie.EpisodeDeep, len(serie_detail_json.Episodes))
	for i, ep := range serie_detail_json.Episodes {
		subtitles := make([]serie.Subtitle, len(ep.Subtitles))
		for j, sub := range ep.Subtitles {
			subtitles[j] = serie.Subtitle{
				Src:     sub.Src,
				Label:   sub.Label,
				Lang:    sub.Lang,
				Default: sub.Default,
			}
		}

		episodes[i] = serie.EpisodeDeep{
			ID:        ep.ID,
			SeriesID:  ep.SeriesID,
			Number:    ep.Number,
			Sub:       ep.Sub,
			Source:    ep.Source,
			Subtitles: subtitles,
		}
	}

	fmt.Println("release date:", serie_detail_json.ReleaseDate)

	var release_date *time.Time
	if serie_detail_json.ReleaseDate != "" {
		parsedTime, err := time.Parse("2006-01-02T15:04:05", serie_detail_json.ReleaseDate)
		if err == nil {
			release_date = &parsedTime
		} else {
			custom_log.NewCustomLog("date_parse_failed", err.Error(), "warning")
		}
	}

	serie_detail := serie.SerieDeepDetail{
		ID:            serie_detail_json.ID,
		Title:         serie_detail_json.Title,
		Description:   serie_detail_json.Description,
		ReleaseDate:   release_date,
		Trailer:       serie_detail_json.Trailer,
		Country:       serie_detail_json.Country,
		Status:        serie_detail_json.Status,
		Type:          serie_detail_json.Type,
		NextEpDateID:  serie_detail_json.NextEpDateID,
		Episodes:      episodes,
		EpisodesCount: serie_detail_json.EpisodesCount,
		Label:         serie_detail_json.Label,
		FavoriteID:    serie_detail_json.FavoriteID,
		Thumbnail:     serie_detail_json.Thumbnail,
	}

	// launch optimized Rod browser
	path := "/usr/bin/google-chrome-stable"
	launcher_instance := launcher.New().
		Bin(path).
		Headless(true).
		NoSandbox(true).
		Set("disable-gpu").
		Set("disable-sync").
		Set("disable-background-networking").
		Set("disable-default-apps").
		MustLaunch()

	browser := rod.New().ControlURL(launcher_instance).MustConnect()
	defer browser.MustClose()

	incognito := browser.MustIncognito()

	// create page pool (reuse pages)
	concurrency := 6
	page_pool := make(chan *rod.Page, concurrency)
	for i := 0; i < concurrency; i++ {
		page_pool <- incognito.MustPage()
	}

	host := os.Getenv("API_HOST")
	port := utils.GetenvInt("API_PORT", 8585)
	proxy_base := fmt.Sprintf("http://%s:%d", host, port)

	// inject sniffer once per page
	const sniff_js = `() => {
			if (window.__scrape_sniffer_ready) return;
			window.__scrape_sniffer_ready = true;
			window.__scrape_sniffer_results = [];

			const push_result = (t, info) => {
				try {
					window.__scrape_sniffer_results.push({ t, info, ts: Date.now() });
					if (t === 'xhr_video' || t === 'video_element') window.__video_found = info;
					if (t === 'xhr_sub') window.__sub_found = info;
				} catch (e) {}
			};

			// Hook XMLHttpRequest
			const orig_open = XMLHttpRequest.prototype.open;
			const orig_send = XMLHttpRequest.prototype.send;
			XMLHttpRequest.prototype.open = function(m, u) { this.__url = u; return orig_open.apply(this, arguments); };
			XMLHttpRequest.prototype.send = function() {
				const url = this.__url || "";
				this.addEventListener('load', function() {
					const type = this.getResponseHeader('content-type') || "";
					if (url.includes('.m3u8') || url.includes('/hls') || type.includes('application/vnd.apple.mpegurl'))
						push_result('xhr_video', { url });
					else if (url.includes('/api/Sub/'))
						push_result('xhr_sub', { url });
				});
				return orig_send.apply(this, arguments);
			};

			// Hook fetch
			const orig_fetch = window.fetch;
			window.fetch = async (i, init) => {
				const req_url = typeof i === 'string' ? i : (i && i.url) || "";
				try {
					if (req_url && (req_url.includes('.m3u8') || req_url.includes('/hls')))
						push_result('xhr_video', { url: req_url });
				} catch (e) {}
				const resp = await orig_fetch(i, init);
				try {
					const type = resp && resp.headers && resp.headers.get ? (resp.headers.get('content-type') || "") : "";
					if (type.includes('application/vnd.apple.mpegurl'))
						push_result('xhr_video', { url: req_url });
				} catch (e) {}
				return resp;
			};

			// Watch <video>
			const watch_video = v => {
				if (!v || v.__watched) return;
				v.__watched = true;
				const report = () => {
					const s = v.currentSrc || v.src || "";
					if (s.includes('.mp4') || s.includes('.m3u8')) push_result('video_element', { url: s });
					else if (s.startsWith('blob:')) push_result('video_blob', { url: s });
				};
				report();
				v.addEventListener('loadedmetadata', report);
				new MutationObserver(report).observe(v, { attributes: true, attributeFilter: ['src'] });
			};
			document.querySelectorAll('video').forEach(watch_video);

			// Watch <iframe> for src OR data-src attributes (countdown or lazyload)
			const isPlayerOrCountdown = src => {
				if (!src || typeof src !== 'string') return false;
				const s = src.toLowerCase();
				return s.includes('countdown') || s.includes('tickcounter') || s.includes('/player/') || s.includes('.m3u8') || s.includes('/hls');
			};

			const normalizeURL = u => {
				if (!u) return '';
				if (u.startsWith('//')) return 'https:' + u;
				return u;
			};

			const watch_iframe = ifr => {
				if (!ifr || ifr.__iframe_watched) return;
				ifr.__iframe_watched = true;
				const report = () => {
					try {
						let src = ifr.getAttribute('src') || "";
						let dataSrc = ifr.getAttribute('data-src') || "";
						let finalSrc = src || dataSrc;
						finalSrc = normalizeURL(finalSrc);
						if (isPlayerOrCountdown(finalSrc)) {
							push_result('video_element', { url: finalSrc });
						}
					} catch (e) {}
				};
				report();

				// Watch for src or data-src changes
				new MutationObserver(() => report()).observe(ifr, { attributes: true, attributeFilter: ['src', 'data-src'] });
			};

			// Existing iframes
			document.querySelectorAll('iframe').forEach(watch_iframe);

			// Watch DOM changes
			new MutationObserver(muts => {
				muts.forEach(m => {
					m.addedNodes.forEach(n => {
						try {
							const tag = (n && n.tagName) ? n.tagName.toUpperCase() : '';
							if (tag === 'VIDEO') watch_video(n);
							if (tag === 'IFRAME') watch_iframe(n);
							if (n.querySelectorAll) {
								n.querySelectorAll('video').forEach(watch_video);
								n.querySelectorAll('iframe').forEach(watch_iframe);
							}
						} catch (e) {}
					});
				});
			}).observe(document.body, { childList: true, subtree: true });

			window.waitForVideo = new Promise(r => {
				const check = () => {
					if (window.__video_found) return r(window.__video_found);
					try {
						for (const it of window.__scrape_sniffer_results) {
							if (it && (it.t === 'video_element' || it.t === 'xhr_video')) {
								window.__video_found = it.info;
								return r(it.info);
							}
						}
					} catch (e) {}
					setTimeout(check, 500);
				};
				check();
			});
		}`

	var wg sync.WaitGroup
	for i := range serie_detail.Episodes {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			page := <-page_pool
			defer func() { page_pool <- page }()

			ep := &serie_detail.Episodes[i]
			fmt.Println("🎬 Processing Episode:", ep.Number)

			// define scraping logic for a single attempt
			try_scrape := func() (string, bool) {
				ep_url := fmt.Sprintf(
					"https://kisskh.co/Drama/%s/Episode-%d?id=%d&ep=%d&page=0&pageSize=100",
					slugify(serie_detail.Title),
					int(ep.Number),
					serie_detail.ID,
					ep.ID,
				)

				page.MustNavigate(ep_url).MustWaitLoad()
				page.Eval(sniff_js)
				page.Eval(`() => { const v = document.querySelector('video'); if (v) { v.muted = true; v.play && v.play().catch(()=>{}); } }`)

				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				ch := make(chan string, 1)
				go func() {
					val, err := page.Eval(`() => window.waitForVideo`)
					if err != nil {
						fmt.Println("Eval error:", err)
						return
					}

					fmt.Println("value val: ", val)

					// convert gson.JSON to map
					obj := val.Value.Map()
					if url_json, ok := obj["url"]; ok {
						url_str := url_json.Str()
						if url_str != "" {
							fmt.Println("✅ Video URL:", url_str)
							ch <- url_str
						}
					}
				}()

				select {
				case video_url := <-ch:
					fmt.Printf("✅ Found video for ep %.0f: %s\n", ep.Number, video_url)
					return video_url, true

				case <-ctx.Done():
					fmt.Println("⏱ Timeout on episode", ep.Number)
					return "", false
				}
			}

			// try to scrap up to 3 times
			var video_url string
			max_retries := 2
			for attempt := 0; attempt <= max_retries; attempt++ {
				url_found, ok := try_scrape()
				if ok {
					video_url = url_found
					break
				}

				if attempt < max_retries {
					backoff := time.Duration(2+attempt*3) * time.Second
					fmt.Printf("🔁 Retrying episode %.0f (attempt %d/%d) after %v...\n", ep.Number, attempt+1, max_retries, backoff)

					// cleanup and reopen fresh page
					page.MustClose()
					page = incognito.MustPage()

					time.Sleep(backoff)
				}
			}

			if video_url == "" {
				fmt.Printf("❌ Failed to find video for ep %.0f after retries\n", ep.Number)
				return
			}

			//  handle video url type
			mime := getMimeFromURL(video_url)
			if strings.Contains(video_url, ".m3u8") || mime == "application/vnd.apple.mpegurl" {
				trimmed := strings.TrimPrefix(video_url, "https://")
				trimmed = strings.TrimPrefix(trimmed, "http://")
				ep.Source = fmt.Sprintf("%s/m3u8/%s", proxy_base, trimmed)
			} else if strings.Contains(video_url, ".mp4") || mime == "video/mp4" {
				encoded := url.QueryEscape(video_url)
				ep.Source = fmt.Sprintf("%s/mp4?url=%s", proxy_base, encoded)
			} else {
				ep.Source = video_url
			}

			// handle subtitle fetching
			val, err := page.Eval(`() => window.__sub_found ? window.__sub_found.url : null`)
			if err == nil {
				sub_url := val.Value.String()
				if sub_url != "" {
					full_sub_url := "https://kisskh.co" + sub_url
					req_sub, _ := http.NewRequest("GET", full_sub_url, nil)
					req_sub.Header.Set("User-Agent", "Mozilla/5.0")
					req_sub.Header.Set("Referer", fmt.Sprintf("https://kisskh.co/Drama/%s", slugify(serie_detail.Title)))
					req_sub.Header.Set("Cookie", cookie_str.String())

					if resp_sub, err := client.Do(req_sub); err == nil {
						defer resp_sub.Body.Close()
						sub_body, _ := io.ReadAll(resp_sub.Body)
						var subs_json []serie.SubtitleJSON
						if json.Unmarshal(sub_body, &subs_json) == nil {
							for j := range subs_json {
								trimmed := strings.TrimPrefix(subs_json[j].Src, "https://")
								trimmed = strings.TrimPrefix(trimmed, "http://")
								subs_json[j].Src = fmt.Sprintf("%s/subtitle/%s", proxy_base, trimmed)
							}
							subs := make([]serie.Subtitle, len(subs_json))
							for j, sub := range subs_json {
								trimmed := strings.TrimPrefix(sub.Src, "https://")
								trimmed = strings.TrimPrefix(trimmed, "http://")
								subs[j] = serie.Subtitle{
									Src:     sub.Src,
									Label:   sub.Label,
									Lang:    sub.Lang,
									Default: sub.Default,
								}
							}

							ep.Subtitles = subs
							fmt.Printf("✅ Parsed %d subtitles for ep %.0f\n", len(subs), ep.Number)
						}
					}
				}
			}
		}(i)
	}

	wg.Wait()

	return &SeriesDeepDetailsResponse{
		SeriesDeepDetails: []serie.SerieDeepDetail{serie_detail},
	}, nil
}

func (sc *ScrapingRepoImpl) GetEpisodes(key int, ep_num int) (*EpisodesResponse, *responses.ErrorResponse) {
	main_url := "https://kisskh.co/"
	client := &http.Client{Timeout: 15 * time.Second}

	// get cookies
	req_main, _ := http.NewRequest("GET", main_url, nil)
	req_main.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/115.0 Safari/537.36")
	resp_main, err := client.Do(req_main)
	if err != nil {
		custom_log.NewCustomLog("scraping_failed", err.Error(), "error")
		return nil, (&responses.ErrorResponse{}).NewErrorResponse("scraping_failed", fmt.Errorf("original_source_error"))
	}
	defer resp_main.Body.Close()

	var cookie_str strings.Builder
	for _, ck := range resp_main.Cookies() {
		cookie_str.WriteString(fmt.Sprintf("%s=%s; ", ck.Name, ck.Value))
	}

	// fetch series detail
	api_url := fmt.Sprintf("https://kisskh.co/api/DramaList/Drama/%d", key)
	req_api, _ := http.NewRequest("GET", api_url, nil)
	req_api.Header.Set("User-Agent", req_main.Header.Get("User-Agent"))
	req_api.Header.Set("Referer", main_url)
	req_api.Header.Set("Cookie", cookie_str.String())

	resp_api, err := client.Do(req_api)
	if err != nil {
		custom_log.NewCustomLog("scraping_failed", err.Error(), "error")
		return nil, (&responses.ErrorResponse{}).NewErrorResponse("scraping_failed", fmt.Errorf("fetch_api_failed"))
	}
	defer resp_api.Body.Close()

	body, _ := io.ReadAll(resp_api.Body)
	var serie_detail_json serie.SerieDeepDetailJSON
	if err := json.Unmarshal(body, &serie_detail_json); err != nil {
		custom_log.NewCustomLog("scraping_failed", err.Error(), "error")
		return nil, (&responses.ErrorResponse{}).NewErrorResponse("scraping_failed", fmt.Errorf("parse_data_failed"))
	}

	// find the requested episode
	var target_ep *serie.EpisodeDeepJSON
	for _, ep := range serie_detail_json.Episodes {
		if fmt.Sprintf("%.0f", ep.Number) == fmt.Sprintf("%d", ep_num) {
			target_ep = &ep
			break
		}
	}
	if target_ep == nil {
		return nil, (&responses.ErrorResponse{}).NewErrorResponse("scraping_failed", fmt.Errorf("episode %d not found in series %d", ep_num, key))
	}

	// setup Rod browser
	path := "/usr/bin/google-chrome-stable"
	launcher_instance := launcher.New().
		Bin(path).
		Headless(true).
		NoSandbox(true).
		Set("disable-gpu").
		Set("disable-sync").
		Set("disable-background-networking").
		Set("disable-default-apps").
		MustLaunch()
	browser := rod.New().ControlURL(launcher_instance).MustConnect()
	defer browser.MustClose()

	incognito := browser.MustIncognito()
	page := incognito.MustPage()

	// sniffer js
	const sniff_js = `() => {
			if (window.__scrape_sniffer_ready) return;
			window.__scrape_sniffer_ready = true;
			window.__scrape_sniffer_results = [];

			const push_result = (t, info) => {
				try {
					window.__scrape_sniffer_results.push({ t, info, ts: Date.now() });
					if (t === 'xhr_video' || t === 'video_element') window.__video_found = info;
					if (t === 'xhr_sub') window.__sub_found = info;
				} catch (e) {}
			};

			// Hook XMLHttpRequest
			const orig_open = XMLHttpRequest.prototype.open;
			const orig_send = XMLHttpRequest.prototype.send;
			XMLHttpRequest.prototype.open = function(m, u) { this.__url = u; return orig_open.apply(this, arguments); };
			XMLHttpRequest.prototype.send = function() {
				const url = this.__url || "";
				this.addEventListener('load', function() {
					const type = this.getResponseHeader('content-type') || "";
					if (url.includes('.m3u8') || url.includes('/hls') || type.includes('application/vnd.apple.mpegurl'))
						push_result('xhr_video', { url });
					else if (url.includes('/api/Sub/'))
						push_result('xhr_sub', { url });
				});
				return orig_send.apply(this, arguments);
			};

			// Hook fetch
			const orig_fetch = window.fetch;
			window.fetch = async (i, init) => {
				const req_url = typeof i === 'string' ? i : (i && i.url) || "";
				try {
					if (req_url && (req_url.includes('.m3u8') || req_url.includes('/hls')))
						push_result('xhr_video', { url: req_url });
				} catch (e) {}
				const resp = await orig_fetch(i, init);
				try {
					const type = resp && resp.headers && resp.headers.get ? (resp.headers.get('content-type') || "") : "";
					if (type.includes('application/vnd.apple.mpegurl'))
						push_result('xhr_video', { url: req_url });
				} catch (e) {}
				return resp;
			};

			// Watch <video>
			const watch_video = v => {
				if (!v || v.__watched) return;
				v.__watched = true;
				const report = () => {
					const s = v.currentSrc || v.src || "";
					if (s.includes('.mp4') || s.includes('.m3u8')) push_result('video_element', { url: s });
					else if (s.startsWith('blob:')) push_result('video_blob', { url: s });
				};
				report();
				v.addEventListener('loadedmetadata', report);
				new MutationObserver(report).observe(v, { attributes: true, attributeFilter: ['src'] });
			};
			document.querySelectorAll('video').forEach(watch_video);

			// Watch <iframe> for src OR data-src attributes (countdown or lazyload)
			const isPlayerOrCountdown = src => {
				if (!src || typeof src !== 'string') return false;
				const s = src.toLowerCase();
				return s.includes('countdown') || s.includes('tickcounter') || s.includes('/player/') || s.includes('.m3u8') || s.includes('/hls');
			};

			const normalizeURL = u => {
				if (!u) return '';
				if (u.startsWith('//')) return 'https:' + u;
				return u;
			};

			const watch_iframe = ifr => {
				if (!ifr || ifr.__iframe_watched) return;
				ifr.__iframe_watched = true;
				const report = () => {
					try {
						let src = ifr.getAttribute('src') || "";
						let dataSrc = ifr.getAttribute('data-src') || "";
						let finalSrc = src || dataSrc;
						finalSrc = normalizeURL(finalSrc);
						if (isPlayerOrCountdown(finalSrc)) {
							push_result('video_element', { url: finalSrc });
						}
					} catch (e) {}
				};
				report();

				// Watch for src or data-src changes
				new MutationObserver(() => report()).observe(ifr, { attributes: true, attributeFilter: ['src', 'data-src'] });
			};

			// Existing iframes
			document.querySelectorAll('iframe').forEach(watch_iframe);

			// Watch DOM changes
			new MutationObserver(muts => {
				muts.forEach(m => {
					m.addedNodes.forEach(n => {
						try {
							const tag = (n && n.tagName) ? n.tagName.toUpperCase() : '';
							if (tag === 'VIDEO') watch_video(n);
							if (tag === 'IFRAME') watch_iframe(n);
							if (n.querySelectorAll) {
								n.querySelectorAll('video').forEach(watch_video);
								n.querySelectorAll('iframe').forEach(watch_iframe);
							}
						} catch (e) {}
					});
				});
			}).observe(document.body, { childList: true, subtree: true });

			window.waitForVideo = new Promise(r => {
				const check = () => {
					if (window.__video_found) return r(window.__video_found);
					try {
						for (const it of window.__scrape_sniffer_results) {
							if (it && (it.t === 'video_element' || it.t === 'xhr_video')) {
								window.__video_found = it.info;
								return r(it.info);
							}
						}
					} catch (e) {}
					setTimeout(check, 500);
				};
				check();
			});
		}`

	// navigate to episode
	ep_url := fmt.Sprintf(
		"https://kisskh.co/Drama/%s/Episode-%d?id=%d&ep=%d&page=0&pageSize=100",
		slugify(serie_detail_json.Title),
		int(target_ep.Number),
		serie_detail_json.ID,
		target_ep.ID,
	)

	// fmt.Println(ep_url)
	page.MustNavigate(ep_url).MustWaitLoad()
	page.Eval(sniff_js)
	page.Eval(`() => { const v = document.querySelector('video'); if(v){v.muted=true;v.play&&v.play().catch(()=>{});} }`)

	// wait for video with retry
	video_url := ""
	retries := 3

	for i := 0; i < retries; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		ch := make(chan string, 1)

		go func() {
			defer close(ch)

			val, err := page.Eval(`() => window.waitForVideo`)
			if err != nil {
				fmt.Println("Eval error:", err)
				return
			}

			obj := val.Value.Map()
			url_json, ok := obj["url"]
			if !ok {
				return
			}

			url_str := url_json.Str()
			if url_str != "" {
				select {
				case ch <- url_str:
				default:
				}
			}
		}()

		select {
		case <-ctx.Done():
			fmt.Println("⏱ Timeout scraping video for episode", ep_num, "retry", i+1)
		case url := <-ch:
			video_url = url
			fmt.Println("✅ Found video:", video_url)
		}

		cancel()

		if video_url != "" {
			break
		}
	}

	if video_url == "" {
		err_msg := &responses.ErrorResponse{}
		return nil, err_msg.NewErrorResponse(
			"scraping_failed",
			fmt.Errorf("no video found for episode %d", ep_num),
		)
	}

	// prepare proxy URL
	host := os.Getenv("API_HOST")
	port := utils.GetenvInt("API_PORT", 8585)
	proxy_base := fmt.Sprintf("http://%s:%d", host, port)

	mime := getMimeFromURL(video_url)
	var proxy_video string
	if strings.Contains(video_url, ".m3u8") || mime == "application/vnd.apple.mpegurl" {
		trimmed := strings.TrimPrefix(video_url, "https://")
		trimmed = strings.TrimPrefix(trimmed, "http://")
		proxy_video = fmt.Sprintf("%s/m3u8/%s", proxy_base, trimmed)
	} else if strings.Contains(video_url, ".mp4") || mime == "video/mp4" {
		proxy_video = fmt.Sprintf("%s/mp4?url=%s", proxy_base, url.QueryEscape(ep_url))
	} else {
		proxy_video = video_url
	}

	// fetch subtitles
	subtitles := []serie.Subtitle{}
	val, err := page.Eval(`() => window.__sub_found ? window.__sub_found.url : null`)
	if err == nil && val.Value.Str() != "" {
		sub_url := "https://kisskh.co" + val.Value.Str()
		req_sub, _ := http.NewRequest("GET", sub_url, nil)
		req_sub.Header.Set("User-Agent", "Mozilla/5.0")
		req_sub.Header.Set("Referer", fmt.Sprintf("https://kisskh.co/Drama/%s", slugify(serie_detail_json.Title)))
		req_sub.Header.Set("Cookie", cookie_str.String())

		if resp_sub, err := client.Do(req_sub); err == nil {
			defer resp_sub.Body.Close()
			sub_body, _ := io.ReadAll(resp_sub.Body)
			var subs_json []serie.SubtitleJSON
			if json.Unmarshal(sub_body, &subs_json) == nil {
				subs := make([]serie.Subtitle, len(subs_json))
				for j, sub := range subs_json {
					trimmed := strings.TrimPrefix(sub.Src, "https://")
					trimmed = strings.TrimPrefix(trimmed, "http://")
					subs[j] = serie.Subtitle{
						Src:     fmt.Sprintf("%s/subtitle/%s", proxy_base, trimmed),
						Label:   sub.Label,
						Lang:    sub.Lang,
						Default: sub.Default,
					}
				}
				subtitles = subs
			}
		}
	}

	return &EpisodesResponse{
		Episodes: []serie.EpisodeDeep{
			{
				ID:        target_ep.ID,
				SeriesID:  key,
				Number:    target_ep.Number,
				Sub:       target_ep.Sub,
				Source:    proxy_video,
				Subtitles: subtitles,
			},
		},
	}, nil
}

func getMimeFromURL(u string) string {
	if strings.Contains(u, ".m3u8") {
		return "application/vnd.apple.mpegurl"
	}
	if strings.Contains(u, ".mp4") {
		return "video/mp4"
	}
	return ""
}

func (sc *ScrapingRepoImpl) Seed() []serie.SerieDeepDetail {
	series_key := []int{
		// korean drama
		//975, 10124, 3749, 7555, 4596, 41, 5043, 5941, 124, 126,

		// chinese drama
		//10826, 8316, 6158, 9148, 773, 9579, 7949, 7628, 10958, 7532,

		// hollywood
		11698, 11764, 11694, 11739, 11521, 11526, 11706, 11705, 11674, 11652,
	}

	var series_deep_detail []serie.SerieDeepDetail

	for _, key := range series_key {
		resp, err := sc.GetDetail(fmt.Sprintf("%d", key))
		if err != nil {
			continue
		}
		series_deep_detail = append(series_deep_detail, resp.SeriesDeepDetails...)
	}

	return series_deep_detail
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
