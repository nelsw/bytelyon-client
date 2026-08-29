package play

import (
	"encoding/json"
	"os"

	"github.com/mxschmitt/playwright-go"
)

func It(headless bool, ƒ func(playwright.BrowserContext)) {
	Init()

	bro, err := NewBrowser(headless)
	if err != nil {
		return
	}
	defer func() { _ = bro.Close() }()

	var ctx playwright.BrowserContext
	if ctx, err = NewBrowserContext(bro); err != nil {
		return
	}
	defer func() { _ = ctx.Close() }()

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

	ƒ(ctx)
}
