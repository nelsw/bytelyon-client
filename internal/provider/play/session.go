package play

import (
	"encoding/json"
	"os"

	"github.com/mxschmitt/playwright-go"
)

// It is a helper method for quickly, safely, and effectively allowing us to "play.It" using device fingerprinting,
// without a playwright.Browser or potentially leaking resources by failing to close critical Playwright interfaces.
func It(headless bool, ƒ func(playwright.BrowserContext)) {
	Init()

	bro, err := NewBrowser(headless)
	if err != nil {
		return
	}
	// close the browser last
	defer func() { _ = bro.Close() }()

	var ctx playwright.BrowserContext
	if ctx, err = NewBrowserContext(bro); err != nil {
		return
	}
	// close the context before the browser to avoid a "force close" equivalent
	defer func() { _ = ctx.Close() }()

	// before closing the browser or context, save the storage state locally
	// state can grow large, so we should write it to a file and eventually
	// implement a cleanup mechanism to truncate the file after a certain period
	defer func() {
		var state *playwright.StorageState
		if state, err = ctx.StorageState(); err != nil {
			return
		}
		var out []byte
		if out, err = json.Marshal(state); err != nil {
			return
		}
		_ = os.WriteFile("state.json", out, os.ModePerm)
	}()

	// execute the function provided by the caller;
	// we use the Dutch florin ascii character
	// because this code is absolutely money.
	ƒ(ctx)
}
