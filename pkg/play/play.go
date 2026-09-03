package play

import (
	_ "embed"
	"fmt"
	"math/rand"
	"time"

	"github.com/nelsw/bytelyon-client/pkg/logs"
	"github.com/nelsw/bytelyon-client/pkg/store"

	"github.com/mxschmitt/playwright-go"
	"github.com/rs/zerolog/log"
)

const (
	scrollScript = `async () => {
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
}`
)

var (
	pwc *playwright.Playwright

	pageScript = new(`() => {
  Object.defineProperty(window.screen, "width", { get: () => 1920 });
  Object.defineProperty(window.screen, "height", { get: () => 1080 });
  Object.defineProperty(window.screen, "colorDepth", { get: () => 24 });
  Object.defineProperty(window.screen, "pixelDepth", { get: () => 24 });
}`)
	contextScript = new(`() => {
  // navigator
  Object.defineProperty(navigator, "webdriver", { get: () => false });
  Object.defineProperty(navigator, "plugins", {
	get: () => [1, 2, 3, 4, 5],
  });
  Object.defineProperty(navigator, "languages", {
	get: () => ["en-US", "en"],
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
}`)
	browserArgs = []string{
		"--disable-accelerated-2d-canvas",
		"--disable-background-networking",
		"--disable-background-timer-throttling",
		"--disable-backgrounding-occluded-windows",
		"--disable-blink-features=AutomationControlled",
		"--disable-breakpad",
		"--disable-component-extensions-with-background-pages",
		"--disable-dev-shm-usage",
		"--disable-extensions",
		"--disable-features=IsolateOrigins,site-per-process",
		"--disable-features=TranslateUI",
		"--disable-gpu",
		"--disable-ipc-flooding-protection",
		"--disable-renderer-backgrounding",
		"--disable-setuid-sandbox",
		"--disable-site-isolation-trials",
		"--disable-web-security",
		"--enable-features=NetworkService,NetworkServiceInProcess",
		"--force-color-profile=srgb",
		"--hide-scrollbars",
		"--metrics-recording-only",
		"--mute-audio",
		"--no-first-run",
		"--no-sandbox",
		"--no-zygote",
	}
)

func init() {
	opts := &playwright.RunOptions{
		Logger:              logs.NewSlog(),
		SkipInstallBrowsers: true,
	}
	if err := playwright.Install(opts); err != nil {
		panic(err)
	} else if pwc, err = playwright.Run(opts); err != nil {
		panic(err)
	}
}

func delay(min, max int) *float64 { return new(float64(rand.Intn(max-min) + min)) }

func Close(context playwright.BrowserContext) {

	if state, err := context.StorageState(); err != nil {
		log.Warn().Err(err).Msg("failed to get browser context storage state")
	} else if err = store.Save(state, "state.json"); err != nil {
		log.Warn().Err(err).Msg("failed to save browser context storage state")
	} else {
		log.Debug().Msg("saved browser context storage state")
	}

	bro := context.Browser()
	_ = context.Close()
	_ = bro.Close()
}

func New(b ...bool) (ctx playwright.BrowserContext, err error) {
	headless := len(b) > 0 && b[0]
	var bro playwright.Browser
	if bro, err = NewBrowser(headless); err == nil {
		ctx, err = NewBrowserContext(bro)
	}
	return
}

// NewBrowser creates a new Browser instance
func NewBrowser(headless ...bool) (playwright.Browser, error) {
	return pwc.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless:          new(len(headless) > 0 && headless[0]),
		Channel:           new("chrome"),
		Timeout:           new(2 * 60_000.),
		Args:              browserArgs,
		IgnoreDefaultArgs: []string{"--enable-automation"},
	})
}

// NewBrowserContext creates a new BrowserContext instance
func NewBrowserContext(bro playwright.Browser) (playwright.BrowserContext, error) {

	state, _ := store.Find[playwright.OptionalStorageState]("state.json")

	var colorScheme *playwright.ColorScheme
	if hour := time.Now().Hour(); hour >= 19 || hour < 7 {
		colorScheme = playwright.ColorSchemeDark
	} else {
		colorScheme = playwright.ColorSchemeLight
	}

	ctx, err := bro.NewContext(playwright.BrowserNewContextOptions{
		AcceptDownloads:   new(true),
		ColorScheme:       colorScheme,
		ForcedColors:      playwright.ForcedColorsNone,
		HasTouch:          new(false),
		IsMobile:          new(false),
		JavaScriptEnabled: new(true),
		Locale:            new("en-US"),
		Permissions:       []string{"geolocation", "notifications"},
		ReducedMotion:     playwright.ReducedMotionNoPreference,
		TimezoneId:        new("America/New_York"),
		UserAgent:         new("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36"),
		StorageState:      &state,
	})

	if err == nil {
		if err = ctx.AddInitScript(playwright.Script{Content: contextScript}); err == nil {
			ctx.SetDefaultTimeout(60_000)
		}
	}
	return ctx, err
}

// NewPage creates a new document in the browser context.
func NewPage(ctx playwright.BrowserContext) (page playwright.Page, err error) {
	if page, err = ctx.NewPage(); err == nil {
		err = page.AddInitScript(playwright.Script{Content: pageScript})
	}
	return
}

func NewTab(x playwright.BrowserContext, l playwright.Locator) (p playwright.Page, err error) {
	if p, err = x.ExpectPage(func() error {
		return l.Click(playwright.LocatorClickOptions{
			Force:     new(true),
			Modifiers: []playwright.KeyboardModifier{"Meta"},
		})
	}); err != nil {
		log.Warn().Err(err).Msg("Client - Failed to ExpectPage")
	} else if err = p.BringToFront(); err != nil {
		log.Warn().Err(err).Msg("Client - Failed to BringToFront")
	} else if err = Scroll(p); err != nil {
		log.Warn().Err(err).Msg("Client - Failed to ScrollToBottomThenTop")
	} else {
		log.Info().Str("url", p.URL()).Msg("Client - NewTab")
	}
	return
}

func Click(p playwright.Page, s string, min, max int) (err error) {
	if err = p.Locator(s).Click(playwright.LocatorClickOptions{Delay: delay(min, max)}); err != nil {
		log.Warn().Err(err).Str("url", p.URL()).Str("selector", s).Msg("click failed")
	} else {
		log.Debug().Str("url", p.URL()).Str("selector", s).Msg("clicked!")
	}
	return
}

func Type(p playwright.Page, s string, min, max int) (err error) {
	if err = p.Keyboard().Type(s, playwright.KeyboardTypeOptions{Delay: delay(min, max)}); err != nil {
		log.Warn().Err(err).Str("url", p.URL()).Str("text", s).Msg("typed failed")
	} else {
		log.Debug().Str("url", p.URL()).Str("text", s).Msg("typed!")
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
		return Scroll(page)
	}
	return nil
}

func Sleep(p playwright.Page, min, max int) { p.WaitForTimeout(*delay(min, max)) }

func Wait(p playwright.Page, s *playwright.LoadState) error {
	return p.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State:   s,
		Timeout: new(60_000.0),
	})
}

func Scroll(p playwright.Page) error {
	_, err := p.Evaluate(scrollScript)
	return err
}

func Title(page playwright.Page) string {
	s, err := page.Title()
	if err != nil {
		log.Err(err).Str("url", page.URL()).Msg("failed to get title")
	} else {
		log.Trace().Str("url", page.URL()).Msg("got title")
	}
	return s
}

func HTML(page playwright.Page) string {
	s, err := page.Content()
	if err != nil {
		log.Err(err).Str("url", page.URL()).Msg("failed to get content")
	} else {
		log.Trace().Str("url", page.URL()).Msg("got content")
	}
	return s
}

func IMG(p playwright.Page, s ...string) []byte {

	opts := playwright.PageScreenshotOptions{FullPage: new(true)}
	if len(s) > 0 {
		opts.Path = new(".storage/" + s[0])
	}

	b, err := p.Screenshot(opts)
	if err != nil {
		log.Err(err).Str("url", p.URL()).Msg("failed to get screenshot")
		return nil
	}

	log.Trace().Str("url", p.URL()).Msg("got screenshot")
	return b
}

func Scrape(url string, ctx playwright.BrowserContext, s ...string) (content string, screenshot []byte) {
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

	Sleep(page, 100, 450)

	l.Debug().Send()

	if len(s) > 0 {
		return HTML(page), IMG(page, s[0])
	}
	return HTML(page), IMG(page)

}
