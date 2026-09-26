<div align="center">
<a href="https://bytelyon.com" target="_blank">
<img src="https://bytelyon.com/bytelyon-banner.png" alt="ByteLyon Logo">
</a>

![Golang](https://img.shields.io/badge/1.27-00ADD8?logo=go&logoColor=white&labelColor=474748)
![Python](https://img.shields.io/badge/3.14-476E99?logo=python&logoColor=white&labelColor=474748)
![Postgres](https://img.shields.io/badge/18-316192?logo=postgresql&logoColor=white&labelColor=474748)
![Redis](https://img.shields.io/badge/9-DD0031?logo=redis&logoColor=white&labelColor=474748)
</div>

## Model

A **bot** is a scheduled job of one of three types (`news`, `search`, `sitemap`). A bot is _due_ when it is enabled and
`last_run_at + frequency <= NOW()` (or it has never run). Search and sitemap bots write into a child record (`child_id`),
resolved at query time in `internal/bot/repo.go`.

```mermaid
erDiagram
    bots ||--o{ articles : "news"
    bots ||--o{ serps : "search (latest = child_id)"
    bots ||--o| sitemaps : "sitemap (domain = query)"
    sitemaps ||--o{ pages : "pageable"

    bots {
        int id PK
        string type "news | search | sitemap"
        string query "keywords, search term, or domain"
        string frequency "hourly | daily | weekly | monthly"
        bool enabled
        bool headless
        string blacklist "newline-delimited words"
        timestamp last_run_at
    }
    articles {
        int bot_id FK
        string url "unique with bot_id"
        string title
        string publisher
        string source "Bing News | Google News"
        string description
        string body
        string img_url
        string img_alt
        array keywords
        timestamp published_at
    }
    serps {
        int id PK
        int bot_id FK
        json data
        string screenshot_key "S3"
        string content_key "S3"
        timestamp deleted_at
    }
    sitemaps {
        int id PK
        int bot_id FK
        string domain
        array urls
        timestamp deleted_at
    }
    pages {
        int pageable_id FK "sitemaps.id"
        string pageable_type "App\\Models\\Sitemap"
        string url "unique with pageable"
        string domain
        string title
        json meta
        string screenshot_key "S3"
    }
```

## Workflows

There are three entry points, all of which funnel into `app.HandleBot`:

| Command    | Source         | Behavior                                                         |
|------------|----------------|------------------------------------------------------------------|
| `make all` | `cmd/all`      | Works every due bot now, then again every 15 minutes             |
| `make one` | `cmd/one`      | Works the single most overdue bot and exits                      |
| `make sub` | `cmd/sub`      | Subscribes to the Redis `bots` channel and works each bot pushed |

```mermaid
flowchart TD
    subgraph entry [Entry points]
        ALL["cmd/all<br/>ticker: 15m"] --> FA["bot.FindAll()"]
        ONE["cmd/one"] --> FO["bot.FindOne()"]
        SUB["cmd/sub"] --> RS[("Redis pub/sub<br/>channel: bots")]
    end

    FA -- "due bots" --> HB
    FO -- "most overdue bot" --> HB
    RS -- "JSON bot payload" --> HB

    HB{"app.HandleBot<br/>switch b.Type"}
    HB -- "id == 0" --> SKIP["log: bot not found"]
    HB -- news --> NEWS["news.Fetch"]
    HB -- search --> SEARCH["search.Fetch"]
    HB -- sitemap --> SITEMAP["sitemap.Fetch"]

    NEWS & SEARCH & SITEMAP --> UPD["bot.Update<br/>last_run_at = NOW()"]
    UPD --> PG[(Postgres)]
```

Each fetcher shells out to a Python/Playwright script in `scripts/`, which writes `<name>.html`, `<name>.png`, and
`<name>.json` under `.storage/`. `play.HandleFiles` then uploads the HTML and screenshot to S3 and returns the parsed
JSON for persistence.

### News

```mermaid
sequenceDiagram
    autonumber
    participant N as news.Fetch
    participant F as Bing / Google News RSS
    participant R as Redis
    participant P as scripts/news
    participant S as S3
    participant DB as Postgres

    par Bing and Google in parallel
        N->>F: GET RSS feed for query
        F-->>N: articles
    end
    loop each article
        N->>N: drop if published before last_run_at or blacklisted
        alt Google News link
            N->>R: cached decoded URL?
            opt cache miss
                N->>F: resolve via batchexecute
                N->>R: cache URL (6h)
            end
        else Bing link
            N->>N: extract ?url= param
        end
    end
    N->>P: -u <unique article URLs>
    P-->>N: .html / .png / .json per URL
    loop each article
        N->>S: upload html + png
        N->>DB: UPSERT articles (url, bot_id)
    end
```

### Search

```mermaid
sequenceDiagram
    autonumber
    participant Q as search.Fetch
    participant P as scripts/sync_search
    participant S as S3
    participant DB as Postgres

    Q->>P: -i serp_id -q query
    P-->>Q: .html / .png / .json
    Q->>S: upload html + png
    Q->>DB: UPDATE serps SET data, screenshot_key, content_key
```

### Sitemap

A breadth-first crawl of same-domain links starting at `https://<query>`, bounded by `SITEMAP_DEPTH`.

```mermaid
flowchart TD
    START["sitemap.Fetch(domain)"] --> CHK{"child_id == 0?"}
    CHK -- yes --> WARN["log: no sitemap for bot"]
    CHK -- no --> BATCH["urls = [https://domain]"]
    BATCH --> PLAY["scripts/pages -u urls"]
    PLAY --> EACH["for each url:<br/>upload html + png to S3<br/>UPSERT pages<br/>mark done"]
    EACH --> LINKS["collect unseen same-domain links → todo"]
    LINKS --> DEPTH{"depth remaining?"}
    DEPTH -- yes --> BATCH2["urls = todo"] --> PLAY
    DEPTH -- no --> MERGE["done += todo"]
    MERGE --> SAVE["UPDATE sitemaps SET urls = done"]
```
