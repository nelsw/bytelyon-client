package model

import "bytelyon-client/internal/provider/api"

type Screenshot struct {
	// ScreenshotKey is the full s3 key for the screenshot.
	ScreenshotKey string `json:"screenshot_key"`

	// ScreenshotData is compressed bytes of a full-page screenshot.
	ScreenshotData []byte `json:"-"`
}

func (s *Screenshot) Save(path string, id int) {
	api.Post(s.ScreenshotData, s.ScreenshotKey, "screenshot_data", path, id)
}
