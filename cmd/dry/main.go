package main

import "github.com/nelsw/bytelyon-client/internal/sitemap"

func main() {
	//meta := `{
	//"X-UA-Compatible": "IE=edge",
	//"article:modified_time": "2026-02-10T15:31:32+00:00",
	//"description": "prado aims to reimagine invisible home technology by reducing visual noise and creating calmer, more refined spaces.",
	//"generator": [
	//  "WordPress 7.1.2",
	//  "WooCommerce 9.9.7",
	//  "WPML ver:4.9.2 stt:8,67;"

	sitemap.SavePage(32, "prado.com", "https://prado.com", "Country Selector - prado | invisible home technology", ".storage/sitemap/32/298a5280-92b2-5d78-9aa4-8e39028d0bcb.png", map[string]any{
		"X-UA-Compatible":       "IE=edge",
		"article:modified_time": "2026-02-10T15:31:32+00:00",
		"description":           "prado aims to reimagine invisible home technology by reducing visual noise and creating calmer, more refined spaces.",
		"generator": []string{
			"WordPress 7.1.2",
			"WooCommerce 9.9.7",
			"WPML ver:4.9.2 stt:8,67;",
		},
		"msapplication-TileImage": "https://prado.com/wp-content/uploads/2024/02/cropped-cropped-brandmark-270x270.png",
		"og:description":          "prado aims to reimagine invisible home technology by reducing visual noise and creating calmer, more refined spaces.",
		"og:locale":               "en_US",
		"og:site_name":            "prado | invisible home technology",
		"og:title":                "Country Selector - prado | invisible home technology",
		"og:type":                 "article",
		"og:url":                  "https://prado.com/",
		"origin-trial":            "A7JYkbIvWKmS8mWYjXO12SIIsfPdI7twY91Y3LWOV/YbZmN1ZhYv8O+Zs6/IPCfBE99aV9tIC8sWZSCN09vf7gkAAACWeyJvcmlnaW4iOiJodHRwczovL2N0LnBpbnRlcmVzdC5jb206NDQzIiwiZmVhdHVyZSI6IkRpc2FibGVUaGlyZFBhcnR5U3RvcmFnZVBhcnRpdGlvbmluZzIiLCJleHBpcnkiOjE3NDIzNDIzOTksImlzU3ViZG9tYWluIjp0cnVlLCJpc1RoaXJkUGFydHkiOnRydWV9",
		"robots":                  "index, follow, max-image-preview:large, max-snippet:-1, max-video-preview:-1",
		"twitter:card":            "summary_large_image",
		"viewport":                "width=device-width, initial-scale=1.0",
	})
}
