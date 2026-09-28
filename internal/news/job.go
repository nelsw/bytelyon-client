package news

import (
	"encoding/json"
	"strconv"

	"github.com/nelsw/bytelyon-client/pkg/emo"
	"github.com/nelsw/bytelyon-client/pkg/model"
	"github.com/rs/zerolog/log"
)

type Job struct {
	headless bool
	article  *Article
}

func (j *Job) Args() []string {
	return []string{
		"-m", strconv.FormatBool(j.headless),
		"-u", j.article.URL,
	}
}

func (j *Job) Success(bytes []byte) {
	var page model.Page
	if err := json.Unmarshal(bytes, &page); err != nil {
		j.Failure(err)
		return
	}

	j.article.Keywords = page.Keywords
	j.article.ImgURL = page.ImgSrc
	j.article.ImgAlt = page.ImgAlt
	j.article.Body = page.Body
	if j.article.Desc == "" {
		j.article.Desc = page.Description
	}

	UpsertArticle(j.article)
	log.Info().EmbedObject(j.article).Msg(emo.Truthy)
	decr(j.article.BotID)
}

func (j *Job) Name() string {
	return "./scripts/sync_news"
}

func (j *Job) Failure(err error) {
	log.Err(err).EmbedObject(j.article).Msg(emo.Falsey)
	decr(j.article.BotID)
}

func (j *Job) Validate() bool {
	incr(j.article.BotID)
	return true
}
