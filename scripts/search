#!/usr/bin/env -S uv run --script
#
# /// script
# requires-python = ">=3.14"
# dependencies = [
#   "boto3",
#   "requests",
#   "beautifulsoup4",
#   "playwright",
#   "seleniumbase",
# ]
# ///
import argparse
import asyncio
import uuid
from dataclasses import dataclass, InitVar, field
from urllib.parse import urlparse

import aioboto3
import aiohttp
import boto3
from botocore.exceptions import ClientError
from bs4 import BeautifulSoup, Tag, NavigableString
from playwright.async_api import async_playwright, Error, Page, Locator, BrowserContext
from seleniumbase import cdp_driver

s3_client = boto3.client('s3')


@dataclass
class Config:
    api_key: str
    app_env: str = field(init=False)
    api_url: str = field(init=False)
    headers: dict[str, str] = field(init=False)
    s3_bucket: str = field(init=False)

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

    async def api_put(self, route: str, data: dict[str, object]) -> None:
        url = f"{self.api_url}/{route}"
        async with aiohttp.ClientSession() as session:
            print(f"ℹ️  API Request: {url}")
            async with session.put(url=url, json=data, headers=self.headers) as response:
                print("✅ API Request:", await response.text())

    def s3_put(self, body: bytes, key: str, ct: str) -> None:
        try:
            print(f"🔷 S3 Upload:  {key}")
            response = s3_client.put_object(
                Body=body,
                Bucket="bytelyon-private",
                Key=key,
                ContentType=ct,
            )
            print(f"✅ S3 Upload:  {key} - ETag: {response['ETag']}")
        except ClientError as e:
            print(f"❌ S3 Upload:  {key} - Error: {e}")

    def upload_content_to_s3(self, body: bytes, key: str):
        self.s3_put(body, key, "text/html")

    def upload_screenshot_to_s3(self, body: bytes, key: str):
        self.s3_put(body, key, "image/png")


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
class Search:
    url: str = field(init=False)
    content_key: str = field(default_factory=str)
    screenshot_key: str = field(default_factory=str)
    data: dict[str, str] = field(default_factory=dict)


# @dataclass
# class Node:
#     index: int
#     kind: str
#     page: InitVar[Page]
#     domain: str = field(init=False)
#     title: str = field(init=False)
#     url: str = field(init=False)
#     meta: dict[str, list[str]] = field(init=False)
#     screenshot_key: str = field(init=False)
#
#     def __post_init__(self, page: Page):
#         self.title = page.title()
#         self.url = page.url
#         self.domain = str(urlparse(self.url).netloc).removeprefix("www.")
#         self.meta = Doc(page.content()).meta
#         self.screenshot_key = file_key(page.url)

@dataclass
class Bot:
    id: int
    search_id: int
    query: str
    blacklist: set[str] = field(default_factory=set)
    url: str = field(init=False)

    def __post_init__(self):
        self.url = f"https://www.google.com?q={self.query.replace(' ', '+')}"


@dataclass
class Job:
    bot: Bot
    cfg: Config
    imgs: dict[str, bytes] = field(default_factory=dict)

    async def run(self):
        driver = await cdp_driver.start_async()
        endpoint_url = driver.get_endpoint_url()
        async with async_playwright() as p:
            browser = await p.chromium.connect_over_cdp(endpoint_url)
            context = browser.contexts[0]
            page = context.pages[0]

            await page.goto("https://www.google.com")
            await page.wait_for_timeout(200)
            search_box = page.locator('textarea[name="q"], input[name="q"]').first
            await search_box.press_sequentially(self.bot.query, delay=40)
            await search_box.press("Enter")
            await page.wait_for_load_state("domcontentloaded")

            captcha = page.locator("iframe[src*='recaptcha'], form#captcha-form")
            if await captcha.count() > 0:
                print("[!] Captcha detected!")
                waited = 0
                while await captcha.count() and waited < 5 * 60:
                    waited += 3
                    await asyncio.sleep(waited)
                if waited >= 5 * 60:
                    print("[-] Timed out waiting for the CAPTCHA to be solved.")
                    return

            await page.wait_for_selector("#search", timeout=20000)

            src_key = self.file_key(self.bot.url, 'html')
            img_key = self.file_key(self.bot.url)
            self.cfg.upload_content_to_s3(bytes(await page.content(), 'utf-8'), src_key)
            self.cfg.upload_screenshot_to_s3(await page.screenshot(), img_key)
            await self.cfg.api_put(f"bots/{self.bot.id}/searches", {
                "data": {
                    "similar_queries": await self.handle_similar_queries(page),
                },
                "content_key": src_key,
                "screenshot_key": img_key,
            })
            await self.handle_organic_products(context, page)
            await self.handle_organic_results(
                context, await page.locator("h3[id]").all()
            )
            await self.handle_sponsored_results(
                context, await page.locator("[data-pcu]").all()
            )
            await self.handle_sponsored_products(
                context, await page.locator("[data-dtld]").all()
            )

            session = aioboto3.Session()
            async with session.client("s3") as s3:
                await asyncio.gather(
                    *[
                        s3.put_object( Body=body, Bucket="bytelyon-private", Key=key, ContentType='image/png')
                        for key, body in self.imgs.items()
                    ]
                )

            await self.cfg.api_put(f"bots/{self.bot.id}", {"result": "ok"})

    def file_key(self, url: str, ext: str = 'png') -> str:
        return f"{self.cfg.app_env}/{self.bot.id}/{uuid.uuid5(uuid.NAMESPACE_URL, url)}.{ext}"

    async def handle_similar_queries(self, p: Page) -> list[str]:
        print(f"ℹ️  Similar Queries")
        top = [await e.get_attribute("data-q") for e in await p.locator("div[data-notify-expansion]").all()]
        end = [await e.text_content() for e in await p.locator("div#botstuff").locator("a").all()]
        result: set[str] = set()
        for t in top + end:
            if t is not None and len(t) > 4:
                result.add(t)
        print(f"✅ Similar Queries: {len(result)}")
        return sorted(list(result))

    async def handle_sponsored_products(self,
                                        context: BrowserContext, locators: list[Locator]
                                        ) -> None:
        print(f"[ ] Handling Sponsored Products {len(locators)}")
        for idx, loc in enumerate(locators):
            domain = await loc.get_attribute("data-dtld")
            merchant_id = await loc.get_attribute("data-merchant-id")
            if domain is None or merchant_id is None:
                continue
            href = await loc.locator(
                f'a[data-merchant-id="{merchant_id}"]'
            ).get_attribute("href")
            if href is not None:
                p = await context.new_page()
                await  p.goto(href, wait_until="domcontentloaded", timeout=5000)
                await self.save_page(p, kind="sponsored_products", index=idx)

    async def handle_sponsored_results(self,
                                       context: BrowserContext, locators: list[Locator]
                                       ) -> None:
        print(f"[ ] Handling Sponsored Results {len(locators)}")
        for idx, loc in enumerate(locators):
            domain = await loc.get_attribute("data-pcu")
            href = await loc.get_attribute("href")
            if domain is None or href is None:
                continue

            p = await context.new_page()
            await p.goto(href, wait_until="domcontentloaded", timeout=5000)
            await self.save_page(p, kind="sponsored_results", index=idx)

    async def handle_organic_results(self,
                                     context: BrowserContext, locators: list[Locator]
                                     ) -> None:
        print(f"[ ] Handling Organic Results {len(locators)}")
        for idx, loc in enumerate(locators):
            href = await loc.locator("xpath=ancestor::a[1]").get_attribute("href")
            if href is not None:
                p = await context.new_page()
                await p.goto(href, wait_until="domcontentloaded", timeout=5000)
                await self.save_page(p, kind="organic_results", index=idx)

    async def handle_organic_products(self, context: BrowserContext, p: Page) -> None:
        locators = await p.locator("product-viewer-entrypoint").all()
        for idx, loc in enumerate(locators):
            print(f"[ ] Handling Organic Product {idx}")
            try:
                await loc.locator("img").first.click(timeout=3000, force=True)
            except Error as e:
                print(f"[!] organic products handler failed: {idx}, {e}")
                continue

            u = await p.locator("div[data-redirect-url]").first.get_attribute(
                "data-redirect-url"
            )
            if u is not None:
                np = await context.new_page()
                await np.goto(url=u, wait_until="domcontentloaded", timeout=5000)
                await self.save_page(np, kind="organic_products", index=idx)


    async def save_page(self, p: Page, kind: str, index: int) -> None:
        key = self.file_key(p.url)
        self.imgs[key] = await p.screenshot(full_page=True)
        await self.cfg.api_put(f"searches/{self.bot.search_id}/page", {
            "title": await p.title(),
            "url": p.url,
            "meta": Doc(await p.content()).meta,
            "screenshot_key": key,
            "kind": kind,
            "index": index,
            "domain": str(urlparse(p.url).netloc).removeprefix("www.")
        })
        await p.close()

# todo - login to google
if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Run a Search 🤖")
    _ = parser.add_argument("-i", "--id", type=int, help="ID of the bot", required=True)
    _ = parser.add_argument("-x", "--search_id", type=int, help="ID of the bot", required=True)
    _ = parser.add_argument("-q", "--query", type=str, help="Query for the bot", required=True)
    _ = parser.add_argument("-b", "--blacklist", type=str, help="Domain Blacklist", required=False, default="")
    _ = parser.add_argument("-k", "--key", type=str, help="ID of the bot", required=True)
    args = parser.parse_args()

    bot = Bot(
        id=args.id,
        search_id=args.search_id,
        query=args.query,
        blacklist=set(args.blacklist.split(","))
    )
    cfg = Config(args.key)

    asyncio.run(Job(bot, cfg).run())
