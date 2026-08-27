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
import json
import uuid
from dataclasses import dataclass, InitVar, field

import boto3
from botocore.exceptions import ClientError
from bs4 import BeautifulSoup, Tag, NavigableString
from playwright.async_api import async_playwright, Error, Page, BrowserContext

s3_client = boto3.client('s3')

api_url = "http://localhost:80/api"
api_headers = {
    "Content-Type": "application/json",
    "x-api-key": "",
}
app_env = "local"

async def put(
    session: ClientSession, route: str, json_data: dict[str, object | list[object]]
) -> None:
    try:
        body = json.dumps(data, indent=4).encode("utf-8")
        print(f"ℹ️  API Request: {url}", body)
        response = await session.put(
            url=f"{os.getenv('APP_URL', default='http://localhost:80')}/api/{route}",
            json=json_data,
            headers={
                "Content-Type": "application/json",
                "x-api-key": os.getenv("API_KEY", default=""),
            },
        )
        print("✅ API Request:", response.json())
    except requests.exceptions.RequestException as e:
        print(f"❌ API Request: {e}")


def scroll_page(page: Page):
    page.evaluate("""async () => {
        await new Promise((resolve) => {
            let totalHeight = 0;
            let distance = 100;
            let timer = setInterval(() => {
                let scrollHeight = document.body.scrollHeight;
                window.scrollBy(0, distance);
                totalHeight += distance;
                if (totalHeight >= scrollHeight || totalHeight >= 10_000) {
                    window.scrollTo(0, 0);
                    clearInterval(timer);
                    resolve();
                }
            }, 100);
        });
    }""")


def upload_screenshot_to_s3(b: bytes, key: str):
    try:
        print(f"ℹ️  S3 Upload Screenshot:{key}")
        res = s3_client.put_object(
            Body=b,
            Bucket="bytelyon-private",
            Key=key,
            ContentType="image/png",
        )
        print(f"✅  S3 Upload Screenshot:{key} - ETag: {res['ETag']}")
    except ClientError as err:
        print(f"❌ S3 UploadScreenshot:{key} - Error: {err}")


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
class Sitemap:
    id: int
    domain: str
    bot_id: int
    headless: bool
    max_concurrency: int = 5
    max_retries: int = 3
    retry_backoff: float | int = 3.0

    urls: set[str] = field(default_factory=set)
    nodes: list[Node] = field(default_factory=list)

    bounder: Semaphore = field(init=False)
    queue: asyncio.Queue[T] = field(init=False)
    lock: asyncio.Lock = field(init=False)

    def __post_init__(self):
        self.bounder = asyncio.Semaphore(self.max_concurrency)
        self.queue = asyncio.Queue()
        self.lock = asyncio.Lock()


    async def extract_links(self, page: Page) -> list[str]:
        """Returns normalized, same-domain absolute URLs found on the page."""
        print(f"[ ] Extracting links from page: {page.url}")
        try:
            links = set[str]()
            exp: str = "() => Array.from(document.querySelectorAll('a[href]')).map(a => a.href)"
            hrefs: list[str | None] = cast(list[str | None], await page.evaluate(exp))
            for href in hrefs:
                if href is not None and str(urlparse(href).netloc).removeprefix("www.") == self.bot.query:
                    parsed = urlparse(href)
                    links.add(
                        f"{parsed.scheme}://{parsed.netloc}{parsed.path}".rstrip("/")
                    )
            print(f"[+] Extracted {len(links)} links from page: {page.url}")
            return list(links)
        except Error as e:
            print(f"[-] Error extracting links from {page.url}: {e}")
            return []


    async def crawl_page(self, context: BrowserContext, u: str) -> list[str]:
        for attempt in range(1, self.max_retries + 2):
            async with self.bounder:
                suffix = (
                    f" (attempt {attempt}/{self.max_retries + 1})"
                    if attempt > 1
                    else ""
                )
                print(f"[+] Crawling: {u}{suffix}")
                page = await context.new_page()
                try:
                    _ = await page.goto(u, wait_until="networkidle", timeout=30000)

                    scroll_page(page)

                    title = await page.title()
                    src = await page.content()
                    img = await page.screenshot(full_page=True)
                    meta = Doc(html=src).meta

                    self.pages.append(Node(u, title, meta, img))

                    return await self.extract_links(page)
                except Error as e:
                    if attempt <= self.max_retries:
                        delay = self.retry_backoff * attempt
                        print(f"[!] {u} failed ({e}); retrying in {delay:.0f}s")
                        await asyncio.sleep(delay)
                finally:
                    await page.close()
        print(f"[-] Giving up on {u} after {self.max_retries + 1} attempts")
        return []


    async def task(self, context: BrowserContext):
        while True:
            u = await self.queue.get()
            try:
                async with self.lock:
                    if url not in self.urls:
                        self.urls.add(url)

                discovered_links = await self.crawl_page(context, u)

                async with self.lock:
                    for link in discovered_links:
                        if link not in self.urls:
                            await self.queue.put(link)
            finally:
                self.queue.task_done()


    async def run(self):
        driver = await cdp_driver.start_async(headless=self.headless)
        endpoint_url = driver.get_endpoint_url()

        await self.queue.put(f"https://{self.bot.query}")

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


@dataclass
class Node:
    url: str
    title: str
    meta: dict[str, list[str]]
    img: bytes
    domain: str = field(init=False)
    screenshot_key: str = field(init=False)

    def __post_init__(self):
        self.domain = str(urlparse(self.url).netloc).removeprefix("www.")
        self.screenshot_key = f"{app_env}/{bot_id}/{uuid.uuid5(uuid.NAMESPACE_URL, url)}.png"


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Run a Sitemap 🤖")
    _ = parser.add_argument("-i", "--id", type=int, help="ID of the bot", required=True)
    _ = parser.add_argument("-x", "--sitemap_id", type=int, help="ID of the bot", required=True)
    _ = parser.add_argument("-q", "--domain", type=str, help="Query for the bot", required=True)
    _ = parser.add_argument("-k", "--key", type=str, help="ID of the bot", required=False,
                            default="my-random-32-character-x-api-key")
    _ = parser.add_argument(
        "--headless", action="store_true", help="Run in headless mode"
    )
    args = parser.parse_args()

    api_headers["x-api-key"] = args.key
    if len(args.key) == 36:
        api_url = "https://bytelyon.com/api"
        app_env = "production"

    sitemap = Sitemap(
        id=args.sitemap_id,
        bot_id=args.bot_id,
        domain=args.domain,
        headless=args.headless,
    )

    asyncio.run(sitemap.run())
    url = f"bots/{sitemap.bot_id}"
    body = {
        "domain": sitemap.domain,
        "urls": sorted(list(sitemap.urls)),
        "pages": sitemap.urls,
    }
    try:
        body = json.dumps(data, indent=4).encode("utf-8")
        print(f"ℹ️  API Request: {url}", body)
        response = requests.put(url, data=body, headers=api_headers)
        response.raise_for_status()  # Raises an error for bad status codes (4xx, 5xx)
        print("✅ API Request:", response.json())
    except requests.exceptions.RequestException as e:
        print(f"❌ API Request: {e}")

