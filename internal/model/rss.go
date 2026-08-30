package model

import "encoding/xml"

type RSS[T any] struct {
	XMLName xml.Name `xml:"rss"`
	Channel struct {
		XMLName xml.Name `xml:"channel"`
		Items   []T      `xml:"item"`
	} `xml:"channel"`
}
