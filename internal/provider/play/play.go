package play

import (
	"bytelyon-client/internal/provider/logs"
	_ "embed"
	"encoding/json/v2"
	"fmt"
	"math/rand"
	"regexp"
	"strings"

	"github.com/mxschmitt/playwright-go"
	"github.com/rs/zerolog/log"
)

var (
	//go:embed state.json
	stateData []byte

	pwc *playwright.Playwright

	blockedRegex = regexp.MustCompile(`(google.com/sorry|captcha|unusual traffic)`)

	searchSelectors = []string{
		"input[name='q']",
		"input[title='Search']",
		"input[aria-label='Search']",
		"textarea[title='Search']",
		"textarea[name='q']",
		"textarea[aria-label='Search']",
		"textarea",
	}
)

func delay(min, max int) *float64 {
	return new(float64(rand.Intn(max-min) + min))
}

func Init() {
	if pwc != nil {
		return
	}
	opts := &playwright.RunOptions{Logger: logs.NewSlog()}
	if err := playwright.Install(opts); err != nil {
		panic(err)
	} else if pwc, err = playwright.Run(opts); err != nil {
		panic(err)
	}
}

// NewBrowser creates a new Browser instance
func NewBrowser(headless bool) (playwright.Browser, error) {
	Init()
	return pwc.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: &headless,
		Timeout:  new(2 * 60_000.0),
		Args: []string{
			"--disable-accelerated-2d-canvas",
			"--disable-background-networking",
			"--disable-background-timer-throttling",
			"--disable-backgrounding-occluded-windows",
			"--disable-blink-features=AutomationControlled",
			"--disable-breakpad",
			"--disable-component-extensions-with-background-page",
			"--disable-dev-shm-usage",
			"--disable-extensions",
			"--disable-features=IsolateOrigins,site-per-process",
			"--disable-features=TranslateUI",
			"--disable-gpu",
			"--disable-ipc-flooding-protection",
			"--disable-renderer-backgrounding",
			"--disable-setuid-sandbox",
			"--disable-site-isolation-trials",
			"--disable-document-security",
			"--enable-features=NetworkService,NetworkServiceInProcess",
			"--force-color-profile=srgb",
			"--hide-scrollbars",
			"--metrics-recording-only",
			"--mute-audio",
			"--no-first-run",
			"--no-sandbox",
			"--no-zygote",
		},
		IgnoreDefaultArgs: []string{
			"--enable-automation",
		},
	})
}

// NewBrowserContext creates a new BrowserContext instance
func NewBrowserContext(bro playwright.Browser) (playwright.BrowserContext, error) {

	userAgent := func() *string {

		agents := []string{
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/104.0.0.0 Safari/537.36",
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/104.0.5112.79 Safari/537.36",
			"Mozilla/5.0 (Windows 7 Enterprise; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.6099.71 Safari/537.36",
			"Mozilla/5.0 (Windows NT 10.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/104.0.0.0 Safari/537.36",
			"Mozilla/5.0 (Windows NT 10.0; WOW64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.5756.197 Safari/537.36",
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/106.0.3713.147 Safari/537.36",
			"Mozilla/5.0 (Windows NT 11.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.6998.166 Safari/537.36",
			"Mozilla/5.0 (Windows Server 2012 R2 Standard; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.5975.80 Safari/537.36",
			"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/103.0.5060.53 Safari/537.36",
			"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/113.0.5672 Safari/537.36",
			"Mozilla/5.0 (X11; Linux x86_64; CentOS Ubuntu 19.04) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.5957.0 Safari/537.36",
		}

		return new(agents[rand.Intn(len(agents))])
	}

	var storageState playwright.OptionalStorageState
	_ = json.Unmarshal(stateData, &storageState)

	ctx, err := bro.NewContext(playwright.BrowserNewContextOptions{
		AcceptDownloads:   new(true),
		ColorScheme:       playwright.ColorSchemeDark,
		ForcedColors:      playwright.ForcedColorsNone,
		HasTouch:          new(false),
		IsMobile:          new(false),
		JavaScriptEnabled: new(true),
		Locale:            new("en-US"),
		Permissions:       []string{"geolocation", "notifications"},
		ReducedMotion:     playwright.ReducedMotionNoPreference,
		TimezoneId:        new("America/New_York"),
		UserAgent:         userAgent(),
		StorageState:      &storageState,
	})

	if err != nil {
		return nil, err
	}

	err = ctx.AddInitScript(playwright.Script{Content: new(`() => {
  // navigator
  Object.defineProperty(navigator, "webdriver", { get: () => false });
  Object.defineProperty(navigator, "plugins", {
	get: () => [1, 2, 3, 4, 5],
  });
  Object.defineProperty(navigator, "languages", {
	get: () => ["en-US", "en", "zh-CN"],
  });

  // window
  window.chrome = {
	runtime: {},
	loadTimes: function () {},
	csi: function () {},
	app: {},
  };

  // WebGL
  if (typeof WebGLRenderingContext !== "undefined") {
	const getParameter = WebGLRenderingContext.prototype.getParameter;
	WebGLRenderingContext.prototype.getParameter = function (
	  parameter: number
	) {
	  // UNMASKED_VENDOR_WEBGL / UNMASKED_RENDERER_WEBGL
	  if (parameter === 37445) {
		return "Intel Inc.";
	  }
	  if (parameter === 37446) {
		return "Intel Iris OpenGL Engine";
	  }
	  return getParameter.call(this, parameter);
	};
  }
}`)})

	if err != nil {
		return nil, err
	}

	ctx.SetDefaultTimeout(60_000)

	return ctx, nil
}

// IsPageBlocked determines if we have been blocked from visiting a URL.
func IsPageBlocked(page playwright.Page) bool {
	return blockedRegex.MatchString(page.URL())
}

// IsRequestBlocked determines if we have been blocked from requesting a URL.
func IsRequestBlocked(res playwright.Response) bool {
	return !res.Ok()
}

// Type fills text to type into a focused element.
func Type(page playwright.Page, s string) error {
	return page.Keyboard().Type(s, playwright.KeyboardTypeOptions{
		Delay: delay(200, 350),
	})
}

// Press executes a keyboard event on a document.
func Press(page playwright.Page, s string) error {
	return page.Keyboard().Press(s, playwright.KeyboardPressOptions{
		Delay: delay(200, 500),
	})
}

// NewPage creates a new document in the browser context.
func NewPage(ctx playwright.BrowserContext) (page playwright.Page, err error) {
	if page, err = ctx.NewPage(); err == nil {
		err = page.AddInitScript(playwright.Script{Content: new(`() => {
  Object.defineProperty(window.screen, "width", { get: () => 1920 });
  Object.defineProperty(window.screen, "height", { get: () => 1080 });
  Object.defineProperty(window.screen, "colorDepth", { get: () => 24 });
  Object.defineProperty(window.screen, "pixelDepth", { get: () => 24 });
}`)})
	}
	return
}

func NewTab(ctx playwright.BrowserContext, l playwright.Locator) (page playwright.Page, err error) {
	var cb = func() error {
		return l.Click(playwright.LocatorClickOptions{
			Force:     new(true),
			Modifiers: []playwright.KeyboardModifier{"Meta"},
			Timeout:   new(0.0),
		})
	}

	var opt = playwright.BrowserContextExpectPageOptions{
		Predicate: func(p playwright.Page) bool { return true },
	}

	if page, err = ctx.ExpectPage(cb, opt); err != nil {
		log.Warn().Err(err).Msg("Client - Failed to ExpectPage")
	} else if err = page.BringToFront(); err != nil {
		log.Warn().Err(err).Msg("Client - Failed to BringToFront")
	} else if err = ScrollToBottomThenTop(page); err != nil {
		log.Warn().Err(err).Msg("Client - Failed to ScrollToBottomThenTop")
	}

	return
}

// GoTo returns the main resource response.
func GoTo(page playwright.Page, url string) (playwright.Response, error) {
	return page.Goto(url, playwright.PageGotoOptions{
		Timeout:   new(10_000.0),
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
}

// Visit navigates to the given URL, waits for the document to load,
// and scrolls to the bottom of the page to ensure all content is visible.
func Visit(page playwright.Page, url string, scroll ...bool) error {
	if res, err := GoTo(page, url); err != nil {
		return err
	} else if !res.Ok() {
		return fmt.Errorf("failed to visit %s: [%d] %s", url, res.Status(), res.StatusText())
	}
	if len(scroll) > 0 && scroll[0] {
		return ScrollToBottomThenTop(page)
	}
	return nil
}

func ScrollToBottomThenTop(page playwright.Page) error {
	_, err := page.Evaluate(`async () => {
  await new Promise((resolve) => {
    let totalHeight = 0;
    let distance = 100;
    let timer = setInterval(() => {
      let scrollHeight = document.body.scrollHeight;
      window.scrollBy(0, distance);
      totalHeight += distance;
      if (totalHeight >= scrollHeight || totalHeight >= 10_000) {
		window.scrollTo(0, 0);
        clearInterval(timer);
        resolve();
      }
    }, 100);
  });
}`)
	return err
}

// Click the first document element located by the given selectors.
func Click(page playwright.Page, selectors ...string) (err error) {

	var count int
	var selector string
	var locator playwright.Locator
	for _, selector = range selectors {

		if locator = page.Locator(selector); locator == nil {
			continue
		}

		if count, err = locator.Count(); err != nil || count == 0 {
			continue
		}

		if err = locator.Click(playwright.LocatorClickOptions{Delay: delay(200, 500)}); err != nil {
			continue
		}

		break
	}
	return
}

// WaitForLoadState returns nil when the required load state has been reached, or error if an exception occurred.
func WaitForLoadState(page playwright.Page, ls ...playwright.LoadState) error {
	s := playwright.LoadStateDomcontentloaded
	if len(ls) > 0 {
		s = &ls[0]
	}
	return page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: s,
	})
}

// Content returns the document content or an empty string if the document has failed to load.
func Content(page playwright.Page) string {
	s, err := page.Content()
	if err != nil {
		log.Err(err).Msg("failed to get document content")
		return ""
	}
	return s
}

// Screenshot returns the screenshot of the document as a byte array or an empty byte array if the document has failed to load.
func Screenshot(page playwright.Page) []byte {
	b, err := page.Screenshot(playwright.PageScreenshotOptions{FullPage: new(true)})
	if err != nil {
		log.Err(err).Msg("failed to get document screenshot")
		return nil
	}
	return b
}

func SearchGoogle(q string, ctx playwright.BrowserContext) (page playwright.Page, err error) {
	if page, err = NewPage(ctx); err != nil {
		return
	}

	var resp playwright.Response
	if resp, err = GoTo(page, "https://www.google.com"); err != nil {
		return
	}

	if IsRequestBlocked(resp) || IsPageBlocked(page) {
		if err = WaitForLoadState(page); err != nil || IsRequestBlocked(resp) || IsPageBlocked(page) {
			return
		}
	}

	if err = Click(page, searchSelectors...); err != nil {
		return
	} else if err = Type(page, q); err != nil {
		return
	} else if err = Press(page, "Enter"); err != nil {
		return
	} else if err = WaitForLoadState(page); err != nil {
		return
	} else if IsPageBlocked(page) {
		return
	}

	log.Info().Msgf("Reached Google SERP for query: %s", q)

	return
}

func Locators(page playwright.Page, s string) []playwright.Locator {
	arr, err := page.Locator(s).All()
	if err != nil {
		log.Warn().
			Err(err).
			Str("selector", s).
			Msg("failed to get locators")
		return []playwright.Locator{}
	}
	log.Trace().Int("count", len(arr)).Msgf("found %d locators for selector: %s", len(arr), s)
	return arr
}

func Attribute(l playwright.Locator, a string) string {
	s, err := l.GetAttribute(a)
	if err != nil {
		log.Warn().Err(err).Msg("failed to get attribute")
		return ""
	}
	return strings.TrimSpace(s)
}

func Scrape(url string, ctx playwright.BrowserContext) (content string, screenshot []byte) {
	l := log.With().
		Str("ƒ", "scrape").
		Str("url", url).
		Logger()

	l.Trace().Send()

	page, err := NewPage(ctx)
	if err != nil {
		l.Warn().Msgf("scrape failed: %s", err.Error())
		return
	}

	defer func() {
		_ = page.Close()
	}()

	if err = Visit(page, url); err != nil {
		l.Warn().Msgf("Visit failed: %s", err.Error())
		return
	}

	l.Debug().Send()

	return Content(page), Screenshot(page)
}

func ScrapeContent(ctx playwright.BrowserContext, url string) (content string, err error) {

	l := log.With().
		Str("ƒ", "scrapeContent").
		Str("url", url).
		Logger()

	l.Trace().Send()

	var page playwright.Page

	if page, err = NewPage(ctx); err != nil {
		l.Warn().Msgf("scrape failed: %s", err.Error())
		return
	}

	defer func() {
		_ = page.Close()
	}()

	if err = Visit(page, url); err != nil {
		l.Warn().Msgf("Visit failed: %s", err.Error())
		return
	}

	if content, err = page.Content(); err != nil {
		l.Warn().Err(err).Msg("failed to get content")
		return
	}

	l.Debug().
		Int("size", len(content)).
		Send()

	return
}
