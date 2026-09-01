package play

import (
	"github.com/mxschmitt/playwright-go"
	"github.com/nelsw/bytelyon-client/pkg/store"
)

// It is a helper method for quickly, safely, and effectively allowing us to "play.It" using device fingerprinting,
// without a playwright.Browser or potentially leaking resources by failing to close critical Playwright interfaces.
func It(headless bool, ƒ func(playwright.BrowserContext)) {

	bro, err := NewBrowser(headless)
	if err != nil {
		return
	}
	// close the browser last
	defer func() { _ = bro.Close() }()

	var context playwright.BrowserContext
	if context, err = NewBrowserContext(bro); err != nil {
		return
	}
	// close the context before the browser to avoid a "force close" equivalent
	defer func() { _ = context.Close() }()

	// before closing the browser or context, save the storage state locally
	// state can grow large, so we should write it to a file and eventually
	// implement a cleanup mechanism to truncate the file after a certain period
	defer func() {
		var state *playwright.StorageState
		if state, err = context.StorageState(); err == nil {
			_ = store.Save(state)
		}
	}()

	// execute the function provided by the caller;
	// we use the Dutch florin ascii character
	// because this code is absolutely money.
	ƒ(context)
}

func New(headless bool) (ctx playwright.BrowserContext, err error) {
	var bro playwright.Browser
	if bro, err = NewBrowser(headless); err == nil {
		ctx, err = NewBrowserContext(bro)
	}
	return
}

func Close(context playwright.BrowserContext) {
	if state, err := context.StorageState(); err == nil {
		_ = store.Save(state)
	}
	bro := context.Browser()
	_ = context.Close()
	_ = bro.Close()
}
