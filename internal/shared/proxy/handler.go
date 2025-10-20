package proxy

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

type ProxyHandler struct {
	DBPool *sqlx.DB
}

func NewProxyHandler(db_pool *sqlx.DB) *ProxyHandler {
	return &ProxyHandler{
		DBPool: db_pool,
	}
}

var (
	mediaCache sync.Map // cache for discovered media URLs
	speedCache sync.Map // cache for speed test results per IP
)

func (pr *ProxyHandler) M3u8(c *fiber.Ctx) error {
	pathParam := c.Params("*")
	log.Println("Incoming request for:", pathParam)

	// Handle preflight OPTIONS request for CORS
	if c.Method() == "OPTIONS" {
		c.Set("Access-Control-Allow-Origin", "*")
		c.Set("Access-Control-Allow-Headers", "*")
		c.Set("Access-Control-Allow-Methods", "*")
		return c.SendStatus(204)
	}

	// Determine target URL
	parts := strings.SplitN(pathParam, "/", 2)
	var target string
	if len(parts) > 1 && strings.Contains(parts[0], ".") {
		target = "https://" + parts[0] + "/" + parts[1]
	} else {
		target = "https://" + pathParam
	}
	// Preserve query params
	if q := c.Context().QueryArgs().String(); q != "" {
		target += "?" + q
	}
	log.Println("Fetching upstream URL:", target)

	// Build request
	req, err := http.NewRequest("GET", target, nil)
	if err != nil {
		return c.Status(500).SendString("failed to build request: " + err.Error())
	}

	// Forward headers
	req.Header.Set("User-Agent", c.Get("User-Agent", "Mozilla/5.0"))
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", c.Get("Accept-Language", "en-US,en;q=0.9"))
	if rng := c.Get("Range"); rng != "" {
		req.Header.Set("Range", rng)
	}
	if len(parts) > 1 && strings.Contains(parts[0], ".") {
		req.Header.Set("Referer", "https://"+parts[0])
		req.Header.Set("Origin", "https://"+parts[0])
	} else {
		req.Header.Set("Referer", "https://khfullhd.co")
		req.Header.Set("Origin", "https://khfullhd.co")
	}
	req.Header.Set("X-Forwarded-For", c.IP())

	client := &http.Client{
		Timeout: 0, // stream until finished
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil
		},
	}

	resp, err := client.Do(req)

	fmt.Println("[RESPONSE] : ", resp)
	if err != nil {
		log.Println("Error fetching upstream:", err)
		return c.Status(500).SendString(err.Error())
	}
	defer resp.Body.Close()

	// Always set CORS headers
	c.Set("Access-Control-Allow-Origin", "*")
	c.Set("Access-Control-Allow-Headers", "*")
	c.Set("Access-Control-Allow-Methods", "*")

	contentType := resp.Header.Get("Content-Type")
	c.Set("Content-Type", contentType)
	c.Status(resp.StatusCode)

	// ✅ Handle .m3u8 playlists (keep your logic untouched)
	if strings.Contains(contentType, "application/vnd.apple.mpegurl") || strings.HasSuffix(pathParam, ".m3u8") {
		body, _ := io.ReadAll(resp.Body)
		lines := strings.Split(string(body), "\n")
		base := "/video-proxy-2"

		for i, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			// Handle EXT-X-KEY separately
			if strings.HasPrefix(line, "#EXT-X-KEY") {
				if strings.Contains(line, "URI=") {
					start := strings.Index(line, "URI=\"") + len("URI=\"")
					end := strings.Index(line[start:], "\"")
					if start > -1 && end > -1 {
						uri := line[start : start+end]
						u, err := url.Parse(uri)
						if err == nil {
							proxyURL := base + "/" + u.Host + u.Path
							if u.RawQuery != "" {
								proxyURL += "?" + u.RawQuery
							}
							line = strings.Replace(line, uri, proxyURL, 1)
						}
					}
				}
				lines[i] = line
				continue
			}

			// Skip comments
			if strings.HasPrefix(line, "#") {
				continue
			}

			// Absolute URLs
			if strings.HasPrefix(line, "http://") || strings.HasPrefix(line, "https://") {
				u, err := url.Parse(line)
				if err != nil {
					continue
				}
				proxyURL := base + "/" + u.Host + u.Path
				if u.RawQuery != "" {
					proxyURL += "?" + u.RawQuery
				}
				lines[i] = proxyURL
			} else {
				// Relative path
				if strings.Contains(line, "?") {
					relParts := strings.SplitN(line, "?", 2)
					lines[i] = base + "/" + parts[0] + "/" + relParts[0] + "?" + relParts[1]
				} else {
					lines[i] = base + "/" + parts[0] + "/" + line
				}
			}
		}
		return c.SendString(strings.Join(lines, "\n"))
	}

	// ✅ Passthrough for .ts, .mp4, etc.
	for k, v := range resp.Header {
		for _, vv := range v {
			c.Set(k, vv)
		}
	}

	// Stream without buffering
	_, err = io.Copy(c.Response().BodyWriter(), resp.Body)
	if err != nil {
		log.Println("Error streaming content:", err)
		return c.Status(500).SendString(err.Error())
	}

	return nil

}

// adaptiveChunkSize determines the optimal chunk size based on network speed.
// It caches the result per client IP.
// ✅ Define adaptiveChunkSize OUTSIDE the route handle
func (pr *ProxyHandler) Mp4(c *fiber.Ctx) error {
	pageURL := c.Query("url")
	if pageURL == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Missing url query")
	}

	clientIP := c.IP()
	// fmt.Println("CACHE URL:", mediaCache)
	log.Println("Starting browser-proxy for:", pageURL, "from", clientIP)

	var mediaURL string

tryFetch:
	// --- Step 1: Check media cache ---
	if val, ok := mediaCache.Load(pageURL); ok {
		mediaURL = val.(string)
		log.Println("[CACHE HIT]", mediaURL)
	} else {
		// --- Step 2: Use Rod to discover media URL ---
		l := launcher.New().Headless(true).NoSandbox(true).MustLaunch()
		browser := rod.New().ControlURL(l).MustConnect()
		page := browser.MustPage()
		defer page.MustClose()

		_ = proto.NetworkEnable{}.Call(page)

		done := make(chan string, 1)
		var once sync.Once
		requestMap := sync.Map{}

		go page.EachEvent(func(e *proto.NetworkRequestWillBeSent) {
			requestMap.Store(e.RequestID, e.Request.URL)
		})()

		go page.EachEvent(func(e *proto.NetworkLoadingFinished) {
			v, ok := requestMap.Load(e.RequestID)
			if !ok {
				return
			}
			url := v.(string)
			if strings.Contains(url, ".mp4") || strings.Contains(url, ".m3u8") || strings.Contains(url, ".ts") {
				log.Println(">>>> FOUND MEDIA URL >>>", url)
				once.Do(func() { done <- url })
			}
		})()

		page.MustNavigate(pageURL)
		page.WaitLoad()

		select {
		case mediaURL = <-done:
			mediaCache.Store(pageURL, mediaURL)
		case <-time.After(30 * time.Second):
			return c.Status(fiber.StatusGatewayTimeout).SendString("Timeout: no media found")
		}
	}

	// --- Step 3: Build request to real media server ---
	rangeHeader := c.Get("Range")

	log.Println("")
	log.Println("=========================================")
	log.Println("[R A N G E - H E A D E R] : ", rangeHeader)
	log.Println("=========================================")
	log.Println("")
	// if rangeHeader == "" || strings.HasSuffix(rangeHeader, "0-") {
	// 	chunkSize := adaptiveChunkSize(clientIP, mediaURL)
	// 	rangeHeader = fmt.Sprintf("bytes=0-%d", chunkSize-1)
	// 	log.Printf("[ADAPTIVE RANGE] %s -> %d MB", clientIP, chunkSize/(1024*1024))
	// }

	if rangeHeader == "" {
		// no range => start adaptive from 0
		chunkSize := adaptiveChunkSize(clientIP, mediaURL)
		rangeHeader = fmt.Sprintf("bytes=0-%d", chunkSize-1)
		log.Printf("[ADAPTIVE RANGE] %s -> %d MB", clientIP, chunkSize/(1024*1024))
	} else {
		// preserve original Range if provided
		log.Printf("[CLIENT RANGE PRESERVED] %s", rangeHeader)
	}

	log.Println("")
	log.Println("=========================================")
	log.Println("[C O N F I G - R A N G E - H E A D E R] : ", rangeHeader)
	log.Println("=========================================")
	log.Println("")

	req, _ := http.NewRequest("GET", mediaURL, nil)
	req.Proto = "HTTP/1.1"
	req.ProtoMajor = 1
	req.ProtoMinor = 1
	req.Header.Set("User-Agent", c.Get("User-Agent", "Mozilla/5.0"))
	req.Header.Set("Referer", "https://kisskh.co/")
	req.Header.Set("Accept", "video/*,audio/*,*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "identity")
	req.Header.Set("Range", rangeHeader)

	// at package level (create once, reuse)
	var mediaTransport = &http.Transport{
		MaxIdleConns:       100,
		MaxConnsPerHost:    100,
		IdleConnTimeout:    90 * time.Second,
		DisableCompression: true,
		ForceAttemptHTTP2:  false, // prefer HTTP/1.1; but TLSNextProto override below is more reliable
		// Disable HTTP/2 completely:
		TLSNextProto: make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
	}

	var mediaClient = &http.Client{
		Transport: mediaTransport,
		Timeout:   0, // long-lived streaming
	}

	resp, err := mediaClient.Do(req)
	if err != nil || resp.StatusCode >= 400 {
		log.Println("Failed proxy request:", err, "status:", resp.StatusCode)

		// --- Step 3.1: Invalidate cache and retry once ---
		if _, ok := mediaCache.Load(pageURL); ok {
			log.Println("[CACHE INVALID] Removing old cache and retrying...")
			mediaCache.Delete(pageURL)
			if resp != nil {
				resp.Body.Close()
			}
			goto tryFetch
		}
		return c.Status(fiber.StatusBadGateway).SendString("Failed to fetch media")
	}
	// defer resp.Body.Close()

	log.Println("")
	log.Println("========================")
	log.Println("[[RESPONSE STATUS]]:", resp)
	log.Println("========================")
	log.Println("")
	// --- Step 4: Copy headers ---
	for k, v := range resp.Header {
		for _, vv := range v {
			c.Set(k, vv)
		}
	}
	c.Status(resp.StatusCode)

	// --- Step 5: Stream video directly, closing upstream only after streaming finishes ---
	// we need to hijack the connection manually to stream efficiently
	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		buf := make([]byte, 32*1024)
		for {
			n, err := resp.Body.Read(buf)
			if n > 0 {
				_, werr := w.Write(buf[:n])
				if werr != nil {
					log.Println("[CLIENT CLOSED WRITE]", werr)
					break
				}
				if ferr := w.Flush(); ferr != nil {
					log.Println("[FLUSH ERROR]", ferr)
					break
				}
			}
			if err != nil {
				if err != io.EOF {
					log.Println("[STREAM READ ERROR]", err)
				}
				break
			}
		}

		// only close AFTER we’ve streamed everything
		if cerr := resp.Body.Close(); cerr != nil {
			log.Println("[UPSTREAM CLOSE ERROR]", cerr)
		}
	})

	return nil
}

func (pr *ProxyHandler) Subtitle(c *fiber.Ctx) error {
	log.Println("🎬 Subtitle proxy request for:", c.Params("*"))

	// --- Handle CORS Preflight ---
	if c.Method() == "OPTIONS" {
		c.Set("Access-Control-Allow-Origin", "*")
		c.Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		c.Set("Access-Control-Allow-Headers", "*")
		return c.SendStatus(204)
	}

	// --- Build Target URL ---
	pathParam := c.Params("*")
	parts := strings.SplitN(pathParam, "/", 2)
	var target string
	if len(parts) > 1 && strings.Contains(parts[0], ".") {
		target = "https://" + parts[0] + "/" + parts[1]
	} else {
		target = "https://" + pathParam
	}

	// --- Preserve Query Parameters ---
	if q := c.Context().QueryArgs().String(); q != "" {
		target += "?" + q
	}

	log.Println("📡 Fetching subtitle from:", target)

	// --- Build Request ---
	req, err := http.NewRequest("GET", target, nil)
	if err != nil {
		return c.Status(500).SendString("Failed to build request: " + err.Error())
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Referer", "https://kisskh.co")
	req.Header.Set("Origin", "https://kisskh.co")

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("❌ Error fetching subtitle:", err)
		return c.Status(500).SendString(err.Error())
	}
	defer resp.Body.Close()

	// --- Allow CORS ---
	c.Set("Access-Control-Allow-Origin", "*")
	c.Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.Set("Access-Control-Allow-Headers", "Content-Type, Origin, Accept")

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return c.Status(500).SendString("Failed to read subtitle body: " + err.Error())
	}

	content := string(body)
	isSRT := strings.HasSuffix(strings.ToLower(pathParam), ".srt") ||
		strings.Contains(strings.ToLower(resp.Header.Get("Content-Type")), "srt")

	if isSRT {
		log.Println("🌀 Converting .srt → .vtt")
		content = convertSRTtoVTT(content)
		c.Set("Content-Type", "text/vtt; charset=utf-8")
	} else {
		c.Set("Content-Type", resp.Header.Get("Content-Type"))
	}

	c.Status(200)
	return c.SendString(content)
}

func (pr *ProxyHandler) Download(c *fiber.Ctx) error {
	pageURL := c.Query("url")
	if pageURL == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Missing url query")
	}

	log.Println("Starting download for:", pageURL)

	// --- Step 1: Check cache ---
	val, ok := mediaCache.Load(pageURL)
	var mediaURL string
	if ok {
		mediaURL = val.(string)
		log.Println("[CACHE HIT]", mediaURL)
	} else {
		// --- Step 2: Use Rod to discover media URL ---
		l := launcher.New().Headless(true).NoSandbox(true).MustLaunch()
		browser := rod.New().ControlURL(l).MustConnect()
		page := browser.MustPage()
		defer page.MustClose()

		_ = proto.NetworkEnable{}.Call(page)

		done := make(chan string, 1)
		var once sync.Once
		requestMap := sync.Map{}

		go page.EachEvent(func(e *proto.NetworkRequestWillBeSent) {
			requestMap.Store(e.RequestID, e.Request.URL)
		})()

		go page.EachEvent(func(e *proto.NetworkLoadingFinished) {
			v, ok := requestMap.Load(e.RequestID)
			if !ok {
				return
			}
			url := v.(string)

			if strings.Contains(url, ".mp4") || strings.Contains(url, ".m3u8") || strings.Contains(url, ".ts") {
				log.Println(">>>> FOUND MEDIA URL >>>", url)
				once.Do(func() { done <- url })
			}
		})()

		page.MustNavigate(pageURL)
		page.WaitLoad()

		select {
		case mediaURL = <-done:
			mediaCache.Store(pageURL, mediaURL) // cache
		case <-time.After(30 * time.Second):
			return c.Status(fiber.StatusGatewayTimeout).SendString("Timeout: no media found")
		}
	}

	go func() {
		rangeHeader := c.Get("Range")
		req, _ := http.NewRequest("GET", mediaURL, nil)
		req.Header.Set("User-Agent", "Mozilla/5.0")
		req.Header.Set("Referer", "https://kisskh.co/")
		req.Header.Set("Accept", "video/*,audio/*,*/*")
		req.Header.Set("Accept-Language", "en-US,en;q=0.5")
		req.Header.Set("Accept-Encoding", "identity")
		req.Header.Set("Range", rangeHeader)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Println("Download failed:", err)
			return
		}
		defer resp.Body.Close()

		file, _ := os.Create("video.mp4")
		defer file.Close()

		buf := make([]byte, 64*1024)
		var total int64
		contentLength, _ := strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64)

		for {
			n, err := resp.Body.Read(buf)
			if n > 0 {
				file.Write(buf[:n])
				total += int64(n)
				if contentLength > 0 {
					percent := float64(total) / float64(contentLength) * 100
					log.Printf("Downloading: %.2f%%\n", percent)
				} else {
					log.Printf("Downloaded: %d bytes\n", total)
				}
			}
			if err != nil {
				if err != io.EOF {
					log.Println("Error reading:", err)
				}
				break
			}
		}

		log.Println("Download completed!")
	}()

	return c.SendString("Download started in background.")
}

func adaptiveChunkSize(clientIP, mediaURL string) int64 {
	// --- Step 1: Check speed cache ---
	if val, ok := speedCache.Load(clientIP); ok {
		chunkSize := val.(int64)
		log.Printf("[SPEED CACHE HIT] IP=%s -> %d MB chunk", clientIP, chunkSize/(1024*1024))
		return chunkSize
	}

	// --- Step 2: Test network speed with 1MB ---
	testRange := "bytes=0-1048575"
	req, _ := http.NewRequest("GET", mediaURL, nil)
	req.Header.Set("Range", testRange)
	req.Header.Set("Accept", "video/*,audio/*,*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "identity")
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Referer", "https://kisskh.co/")

	start := time.Now()
	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode >= 400 {
		log.Printf("[SPEED TEST FAIL] Using default 4MB chunk for %s", clientIP)
		return 4 * 1024 * 1024
	}
	defer resp.Body.Close()

	var totalRead int64
	buf := make([]byte, 64*1024)
	for {
		n, err := resp.Body.Read(buf)
		totalRead += int64(n)
		if err != nil {
			break
		}
	}

	duration := time.Since(start).Seconds()
	if duration <= 0 {
		duration = 0.1
	}

	speedMbps := float64(totalRead*8) / (duration * 1_000_000)
	log.Printf("[SPEED TEST] %s -> %.2f Mbps", clientIP, speedMbps)

	var chunkSize int64
	switch {
	case speedMbps < 2:
		chunkSize = 2 * 1024 * 1024
	case speedMbps < 5:
		chunkSize = 4 * 1024 * 1024
	case speedMbps < 10:
		chunkSize = 8 * 1024 * 1024
	default:
		chunkSize = 16 * 1024 * 1024
	}

	speedCache.Store(clientIP, chunkSize)
	log.Printf("[SPEED CACHED] %s -> %d MB chunk", clientIP, chunkSize/(1024*1024))
	return chunkSize
}

// --- Helper: Convert SRT → VTT ---
func convertSRTtoVTT(srt string) string {
	lines := strings.Split(srt, "\n")
	var vtt strings.Builder
	vtt.WriteString("WEBVTT\n\n")

	for _, line := range lines {
		// Convert --> to --> with .
		if strings.Contains(line, "-->") {
			line = strings.ReplaceAll(line, ",", ".")
		}
		vtt.WriteString(line + "\n")
	}

	return vtt.String()
}
