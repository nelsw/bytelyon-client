package news

import (
	"encoding/xml"
	"time"
)

type RSS[T any] struct {
	XMLName xml.Name `xml:"rss"`
	Channel struct {
		XMLName xml.Name `xml:"channel"`
		Items   []T      `xml:"item"`
	} `xml:"channel"`
}

// PubDate wraps time.Time to add custom unmarshaling logic
type PubDate time.Time

func (t *PubDate) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var s string
	if err := d.DecodeElement(&s, &start); err != nil {
		return err
	}
	tmp, err := time.Parse(time.RFC1123, s)
	if err != nil {
		return err
	}
	*t = PubDate(tmp)
	return nil
}

func (t *PubDate) MarshalJSON() ([]byte, error) {
	return []byte(`"` + time.Time(*t).Format(time.RFC3339) + `"`), nil
}
