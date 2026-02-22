#!/usr/bin/env python3
"""
Load test script for the relay proxy.
Sends requests concurrently at a given rate and tracks which backend
(by response "host") handled each request.
"""

import argparse
import asyncio
import json
import sys
import time
from collections import defaultdict

try:
    import aiohttp
except ImportError:
    print("Install aiohttp: pip install aiohttp", file=sys.stderr)
    sys.exit(1)


async def fetch(
    session: aiohttp.ClientSession,
    url: str,
    request_id: int,
    timeout: float,
    verbose: bool,
) -> tuple[int, str | None, str | None]:
    """Return (request_id, backend host or None, error or None)."""
    try:
        async with session.get(url, timeout=aiohttp.ClientTimeout(total=timeout)) as r:
            r.raise_for_status()
            data = await r.json()
            raw = data.get("host") or "unknown"
            backend = str(raw) if raw is not None else "unknown"
            if verbose:
                print(f"  {request_id:5d}  backend={backend}")
            return (request_id, backend, None)
    except (aiohttp.ClientError, asyncio.TimeoutError) as e:
        if verbose:
            print(f"  {request_id:5d}  error: {e}", file=sys.stderr)
        return (request_id, None, str(e))
    except (KeyError, json.JSONDecodeError) as e:
        if verbose:
            print(f"  {request_id:5d}  bad response: {e}", file=sys.stderr)
        return (request_id, None, str(e))


async def run_request(
    session: aiohttp.ClientSession,
    url: str,
    request_id: int,
    start_at: float,
    timeout: float,
    verbose: bool,
) -> tuple[int, str | None, str | None]:
    """Wait until start_at then run the request."""
    now = time.perf_counter()
    delay = start_at - now
    if delay > 0:
        await asyncio.sleep(delay)
    return await fetch(session, url, request_id, timeout, verbose)


async def main_async(args: argparse.Namespace) -> None:
    interval = 1.0 / args.rate if args.rate > 0 else 0
    backend_counts: dict[str, int] = defaultdict(int)
    errors = 0

    start = time.perf_counter()
    # High connection limit so we can achieve high request rate (default 100 caps throughput)
    connector = aiohttp.TCPConnector(limit=0, limit_per_host=0)
    async with aiohttp.ClientSession(connector=connector) as session:
        tasks = []
        for i in range(args.requests):
            url = args.urls[i % len(args.urls)]
            start_at = start + i * interval
            tasks.append(
                asyncio.create_task(
                    run_request(
                        session,
                        url,
                        i + 1,
                        start_at,
                        args.timeout,
                        args.verbose,
                    )
                )
            )
        results = await asyncio.gather(*tasks, return_exceptions=True)

    for r in results:
        if isinstance(r, BaseException):
            errors += 1
            if args.verbose:
                print(f"  task error: {r}", file=sys.stderr)
            continue
        req_id, backend, err = r
        if err:
            errors += 1
        elif backend:
            backend_counts[backend] += 1

    total_time = time.perf_counter() - start

    # Summary
    print()
    print("--- Summary ---")
    print(f"  Total requests:  {args.requests}")
    print(f"  Errors:          {errors}")
    print(f"  Duration:        {total_time:.2f}s")
    print(f"  Actual rate:     {args.requests / total_time:.1f} req/s")
    print()
    print("  Backend distribution (host):")
    for host in sorted(backend_counts.keys(), key=lambda h: (h == "unknown", str(h))):
        pct = 100.0 * backend_counts[host] / args.requests
        print(f"    {host:>10s}  {backend_counts[host]:5d}  ({pct:5.1f}%)")
    if errors:
        sys.exit(1)


def main() -> None:
    parser = argparse.ArgumentParser(
        description="Send requests at a fixed rate (concurrent) and track which backend handled each."
    )
    parser.add_argument(
        "urls",
        nargs="+",
        help="One or more target URLs (e.g. http://localhost:8080/). If multiple, requests are round-robined.",
    )
    parser.add_argument(
        "-r",
        "--rate",
        type=float,
        required=True,
        metavar="REQ_PER_SEC",
        help="Request rate in requests per second.",
    )
    parser.add_argument(
        "-n",
        "--requests",
        type=int,
        default=100,
        metavar="N",
        help="Total number of requests to send (default: 100).",
    )
    parser.add_argument(
        "-t",
        "--timeout",
        type=float,
        default=10.0,
        metavar="SEC",
        help="Request timeout in seconds (default: 10).",
    )
    parser.add_argument(
        "-v",
        "--verbose",
        action="store_true",
        help="Print each request response (backend host).",
    )
    args = parser.parse_args()

    asyncio.run(main_async(args))


if __name__ == "__main__":
    main()
