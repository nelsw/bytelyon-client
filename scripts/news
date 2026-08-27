#!/usr/bin/env -S uv run --script
#
# /// script
# requires-python = ">=3.14"
# dependencies = [
#   "boto3",
#   "aiofiles",
#   "beautifulsoup4",
#   "httpx",
#   "playwright",
#   "python-dotenv",
#   "pytz",
#   "seleniumbase",
# ]
# ///
import argparse
import asyncio
import uuid
from abc import ABC, abstractmethod
from asyncio import Semaphore
from dataclasses import InitVar, dataclass, field
from datetime import datetime as time
from typing import cast
from urllib.parse import quote
from xml.etree.ElementTree import Element, fromstring

import boto3
import aiohttp
from botocore.exceptions import ClientError
from bs4 import BeautifulSoup, Tag
from bs4.element import NavigableString
from playwright.async_api import (
    BrowserContext,
    Error,
    Page,
    async_playwright,
)
from pytz import timezone
from seleniumbase import cdp_driver

RFC_1123 = "%a, %d %b %Y %H:%M:%S %Z"
TZ = timezone("UTC")


@dataclass
class Config:
    api_key: str
    app_env: str = field(init=False)
    api_url: str = field(init=False)
    headers: dict[str, str] = field(init=False)
    s3_bucket: str = field(init=False)
    s3_client = boto3.client('s3')

    def __post_init__(self) -> None:
        self.headers = {
            "Content-Type": "application/json",
            "x-api-key": self.api_key,
        }
        # the key is the length of a UUID ...
        if len(self.api_key) == 36:
            self.app_env = 'production'
            self.api_url = "https://bytelyon.com/api"
        else:
            self.app_env = 'local'
            self.api_url = "http://localhost:80/api"

    async def api_put(self, route: str, json: dict[str, object]) -> None:
        url = f"{self.api_url}/{route}"
        async with aiohttp.ClientSession() as session:
            print(f"ℹ️  API Request: {url}")
            async with session.put(url=url, json=json, headers=self.headers) as response:
                print("✅ API Request:", await response.text())


@dataclass
class Headline:
    url: str
    at: str
    title: str

    def published_after(self, t: time | None) -> bool:
        return t is None or time.strptime(self.at, RFC_1123).astimezone(tz=TZ) > t


@dataclass
class Doc:
    html: InitVar[str]
    soup: BeautifulSoup = field(init=False)
    meta: dict[str, list[str]] = field(default_factory=dict)

    def __post_init__(self, html: str) -> None:
        self.meta = dict[str, list[str]]()
        self.soup = BeautifulSoup(html, "html.parser")
        for tag in self.soup.find_all("meta"):
            k = tag.get("name") or tag.get("property")
            if not isinstance(k, str):
                continue

            vv = tag.get_attribute_list("content")
            if not k or not vv:
                continue

            k = k.lower()
            vals = set[str]()
            if k in self.meta:
                vals = set(self.meta[k])

            for v in set(vv):
                for val in set(v.split(",")):
                    vals.add(val.strip())

            self.meta[k] = list(vals)
        print("made doc")

    def value(self, *keys: str) -> str:
        for k in keys:
            vals = self.meta.get(k.lower())
            if vals:
                for v in vals:
                    v = v.strip()
                    if v:
                        return v
        return ""

    def title(self) -> str:
        v = self.value("twitter:title", "og:title", "title")
        if v:
            return v
        title_tag = self.soup.find("title")
        return title_tag.get_text().strip() if title_tag else ""

    def img_url(self) -> str:
        return self.value(
            "twitter:image:src",
            "twitter:image",
            "og:image:secure_url",
            "og:image:url",
            "og:image",
            "image",
        )

    def img_alt(self) -> str:
        return self.value("twitter:image:alt", "og:image:alt")

    def source(self) -> str:
        return self.value("twitter:site", "og:site_name", "og:site")

    def description(self) -> str:
        return self.value(
            "twitter:description", "og:description", "description", "abstract"
        )

    def keywords(self) -> list[str]:
        kw = set[str]()
        for opt in ["keywords", "news_keywords", "article:tag"]:
            vals = self.meta.get(opt)
            if vals is None:
                continue
            for val in vals:
                for v in val.split(","):
                    kw.add(v)

        return sorted(kw)

    def body(self) -> str:
        # Clone soup to avoid modifying the original if needed,
        # though in Go version it seems to modify the document if it falls back to 'body' or 'html'

        sel = self.soup.find("article")
        if not sel:
            sel = self.soup.find("main")
        if not sel:
            sel = self.soup.find("body")
        if not sel:
            sel = self.soup

        unique_text: dict[str, int] = {}
        # The Go code does: sel.Find("*").Contents().Each(...)
        # and checks if node.Type == html.TextNode
        # In BeautifulSoup, we can iterate over all descendants and check if they are NavigableString

        # We need to maintain order based on appearance
        all_elements = sel.find_all(True)  # True finds all tags

        # We also need to check the selection itself if it's a tag and has direct text
        elements_to_check = [sel] + all_elements

        index = 0
        for element in elements_to_check:
            for content in element.contents:
                if isinstance(content, NavigableString) and not isinstance(
                    content, Tag
                ):
                    text = str(content).strip()
                    if text and text not in unique_text:
                        unique_text[text] = index
                    index += 1

        # Sort by index to preserve order
        ordered_text = sorted(unique_text.items(), key=lambda item: item[1])
        return "\n\n".join([item[0] for item in ordered_text])


@dataclass
class Bot:
    id: int
    query: str
    headless: bool
    after: time | None = None

    def object_key(self, url: str, ext: str = "png") -> str:
        return f"output/{self.id}/{uuid.uuid5(uuid.NAMESPACE_URL, url)}.{ext}"


@dataclass
class AsyncJob[T](ABC):
    bot: Bot
    cfg: Config
    max_concurrency: int = 5
    max_retries: int = 2
    retry_backoff: float | int = 2.5

    bounder: Semaphore = field(init=False)
    queue: asyncio.Queue[T] = field(init=False)
    lock: asyncio.Lock = field(init=False)

    def __post_init__(self):
        self.bounder = asyncio.Semaphore(self.max_concurrency)
        self.queue = asyncio.Queue()
        self.lock = asyncio.Lock()


@dataclass
class News(AsyncJob[Headline]):
    visited_urls: set[str] = field(default_factory=set)

    async def run(self) -> None:
        urls = [
            f"https://www.bing.com/news/search?format=rss&q={quote(self.bot.query)}",
            f"https://news.google.com/rss/search?q={quote(self.bot.query)}&hl=en-US&gl=US&ceid=US:en",
        ]
        async with aiohttp.ClientSession() as session:
            _ = await asyncio.gather(*[self.add_headlines(session, url) for url in urls])

        driver = await cdp_driver.start_async(headless=self.bot.headless)
        endpoint_url = driver.get_endpoint_url()
        async with async_playwright() as p:
            browser = await p.chromium.connect_over_cdp(endpoint_url)
            try:
                workers = [
                    asyncio.create_task(coro=self.task(context=browser.contexts[0]))
                    for _ in range(self.max_concurrency)
                ]
                await self.queue.join()
                for worker_task in workers:
                    _ = worker_task.cancel()
                _ = await asyncio.gather(*workers, return_exceptions=True)
            finally:
                await browser.close()
        driver.stop()
        await self.cfg.api_put(f"bots/{self.bot.id}", {"result": "ok"})

    async def task(self, context: BrowserContext):
        while True:
            a: Headline = await self.queue.get()
            if a.url == "chrome-error://chromewebdata/":
                continue
            try:
                async with self.lock:
                    if a.url in self.visited_urls:
                        continue
                    self.visited_urls.add(a.url)
                await self.scrape(context, a)
            finally:
                self.queue.task_done()

    async def scrape(self, ctx: BrowserContext, headline: Headline) -> None:
        for attempt in range(1, self.max_retries + 2):
            print(f"scrape title={headline.title} attempt={attempt}")
            async with self.bounder:
                page = await ctx.new_page()
                try:
                    await page.goto(url=headline.url, wait_until="domcontentloaded", timeout=5_000)
                    await page.wait_for_timeout(1600)

                    doc = Doc(await page.content())
                    await self.cfg.api_put(f"bots/{self.bot.id}/articles", {
                        "url": page.url,
                        "published_at": headline.at,
                        "title": headline.title,
                        "source": doc.source(),
                        "publisher": doc.source(),
                        "img_url": doc.img_url(),
                        "img_alt": doc.img_alt(),
                        "body": doc.body(),
                        "keywords": doc.keywords(),
                        "description": doc.description(),
                    })
                    return
                except Error as e:
                    print(f"[!] upsert error: {e!s}")
                    if attempt <= self.max_retries:
                        delay = self.retry_backoff * attempt
                        await asyncio.sleep(delay)
                finally:
                    await page.close()

    @staticmethod
    async def get_xml(session: aiohttp.ClientSession, url: str) -> Element[str] | None:
        print(f"[ ] fetch_xml {url}")
        async with session.get(url) as response:
            if response.status >= 300:
                print(f"[!] fetch_xml {url} - {response.status}")
                return None

            print(f"[+] fetch_xml {url}")
            return fromstring(text=await response.text(encoding="utf-8"))

    async def add_headlines(self, session: aiohttp.ClientSession, url: str) -> None:
        xml = await self.get_xml(session, url)
        if xml is None:
            return

        for e in xml.findall(".//item"):
            h = Headline(
                url=e.findtext("link", default=""),
                at=e.findtext("pubDate", default=""),
                title=e.findtext("title", default=""),
            )
            if h.published_after(self.bot.after) and h.url != "chrome-error://chromewebdata/":
                await self.queue.put(h)
        return


if __name__ == "__main__":

    parser = argparse.ArgumentParser(description="Run a 🤖")
    _ = parser.add_argument("-i", "--id", type=int, help="ID of the bot")
    _ = parser.add_argument("-q", "--query", type=str, help="Query for the bot")
    _ = parser.add_argument(
        "-a", "--after", type=str, help="Result date start", default=None
    )
    _ = parser.add_argument("-k", "--key", type=str, help="API Auth Key", default="my-random-32-character-x-api-key")
    _ = parser.add_argument(
        "--headless", action="store_true", help="Run in headless mode"
    )
    args = parser.parse_args()

    after_arg: str | None = cast(str | None, args.after)
    if after_arg is not None:
        after = time.fromisoformat(after_arg)
    else:
        after = None

    bot = Bot(
        id=cast(int, args.id),
        query=cast(str, args.query),
        after=after,
        headless=cast(bool, args.headless),
    )

    cfg = Config(args.key)

    asyncio.run(News(bot, cfg).run())
