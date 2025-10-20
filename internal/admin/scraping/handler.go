package scraping

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	response "rerng_addicted_api/pkg/http/response"
	custom_log "rerng_addicted_api/pkg/logs"
	types "rerng_addicted_api/pkg/model"
	"rerng_addicted_api/pkg/utils"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

type ScrapingHandler struct {
	DBPool          *sqlx.DB
	ScrapingService func(c *fiber.Ctx) *ScrapingService
}

func NewScrapingHandler(db_pool *sqlx.DB) *ScrapingHandler {
	return &ScrapingHandler{
		DBPool: db_pool,
		ScrapingService: func(c *fiber.Ctx) *ScrapingService {
			var uCtx types.UserContext
			// convert map to UserContext struct
			uCtx, ok := c.Locals("UserContext").(types.UserContext)
			if !ok {
				custom_log.NewCustomLog("user_context_failed", "UserContext missing or invalid", "warn")
				uCtx = types.UserContext{}
			}

			return NewScrapingService(db_pool, &uCtx)
		},
	}
}

func (sc *ScrapingHandler) Search(c *fiber.Ctx) error {
	keyword := c.Query("keyword")

	resp, err := sc.ScrapingService(c).Search(keyword)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(
			response.NewResponseError(
				utils.Translate(err.MessageID, nil, c),
				-2000,
				fmt.Errorf("%s", utils.Translate(err.Err.Error(), nil, c)),
			),
		)
	}

	return c.Status(http.StatusOK).JSON(
		response.NewResponse(
			utils.Translate("scraping_success", nil, c),
			2000,
			resp,
		),
	)
}

func (sc *ScrapingHandler) GetDetail(c *fiber.Ctx) error {
	key := c.Params("key")
	if strings.TrimSpace(key) == "" {
		return c.Status(http.StatusBadRequest).JSON(
			response.NewResponseError(
				utils.Translate("scraping_failed", nil, c),
				-2001,
				fmt.Errorf("%s", utils.Translate("need_key", nil, c)),
			),
		)
	}

	resp, err := sc.ScrapingService(c).GetDetail(key)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(
			response.NewResponseError(
				utils.Translate(err.MessageID, nil, c),
				-2001,
				fmt.Errorf("%s", utils.Translate(err.Err.Error(), nil, c)),
			),
		)
	}

	return c.Status(http.StatusOK).JSON(
		response.NewResponse(
			utils.Translate("scraping_success", nil, c),
			2001,
			resp,
		),
	)
}

func (sc *ScrapingHandler) GetDeepDetail(c *fiber.Ctx) error {
	key := c.Params("key")
	if strings.TrimSpace(key) == "" {
		return c.Status(http.StatusBadRequest).JSON(
			response.NewResponseError(
				utils.Translate("scraping_failed", nil, c),
				-2001,
				fmt.Errorf("%s", utils.Translate("need_key", nil, c)),
			),
		)
	}

	locale := c.AcceptsLanguages("en", "km", "zh")
	if locale == "" {
		locale = "en"
	}

	fmt.Println("Accept-languages :", locale)

	translate := func(msg string, params map[string]interface{}) string {
		return utils.TranslateSafe(msg, params, locale)
	}

	scrapingService := sc.ScrapingService(c)

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		totalSteps := 20
		maxFakePercent := 80 // only go up to 80% before service finishes

		// Fake progress up to 80%
		for i := 0; i <= totalSteps*maxFakePercent/100; i++ {
			filled := strings.Repeat("█", i)
			empty := strings.Repeat("░", totalSteps-i)
			percent := (i * 100) / totalSteps
			fmt.Fprintf(w, "data: [%s%s] %d%%\n\n", filled, empty, percent)
			w.Flush()
			time.Sleep(200 * time.Millisecond)
		}

		// Do actual scraping (real work)
		resp, err := scrapingService.GetDeepDetail(key)
		if err != nil {
			msg := translate(err.Err.Error(), nil)
			fmt.Fprintf(w, "data: %s\n\n", msg)
			fmt.Fprintf(w, "event: done\ndata: error\n\n")
			w.Flush()
			return
		}

		// Jump progress to 100%
		fmt.Fprintf(w, "data: [%s] 100%%\n\n", strings.Repeat("█", totalSteps))
		w.Flush()

		// Send final data
		result, _ := json.Marshal(response.NewResponse(
			translate("scraping_success", nil),
			2001,
			resp,
		))
		fmt.Fprintf(w, "data: %s\n\n", string(result))
		fmt.Fprintf(w, "event: done\ndata: complete\n\n")
		w.Flush()
	})

	return nil
}
