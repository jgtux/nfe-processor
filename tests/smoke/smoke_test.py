#!/usr/bin/env python3
"""
NF-e Processor — Smoke Test
Usage: python3 smoke_test.py [--base-url http://localhost:8080]
"""

import argparse
import json
import os
import sys
import time
import tempfile
from pathlib import Path

try:
    import requests
except ImportError:
    print("ERROR: requests not installed. Run: pip install requests")
    sys.exit(1)

# ── Config ────────────────────────────────────────────────────────────────────

BASE_URL = "http://localhost:8080"
API      = f"{BASE_URL}/api/v1"

EXAMPLES_DIR = Path(__file__).parent / "examples"

GREEN  = "\033[92m"
RED    = "\033[91m"
YELLOW = "\033[93m"
BOLD   = "\033[1m"
RESET  = "\033[0m"

passed = 0
failed = 0

# ── Helpers ───────────────────────────────────────────────────────────────────

def ok(msg):
    global passed
    passed += 1
    print(f"  {GREEN}✓{RESET} {msg}")

def fail(msg, detail=""):
    global failed
    failed += 1
    detail_str = f"\n      {YELLOW}{detail}{RESET}" if detail else ""
    print(f"  {RED}✗{RESET} {msg}{detail_str}")

def section(title):
    print(f"\n{BOLD}── {title} {'─' * (50 - len(title))}{RESET}")

def check(condition, msg_ok, msg_fail, detail=""):
    if condition:
        ok(msg_ok)
    else:
        fail(msg_fail, detail)

def get(path, **kwargs):
    return requests.get(f"{API}{path}", timeout=10, **kwargs)

def post_files(files_dict):
    return requests.post(f"{API}/xml/upload", files=files_dict, timeout=10)

def wait_for_processing(seconds=2):
    time.sleep(seconds)

# ── Tests ─────────────────────────────────────────────────────────────────────

def test_health():
    section("Health Check")
    try:
        r = get("/health")
        check(r.status_code == 200, "GET /health returns 200", f"Expected 200, got {r.status_code}")
        data = r.json()
        check(data.get("data", {}).get("status") == "ok",
              "Response body has status=ok",
              f"Unexpected body: {data}")
    except requests.exceptions.ConnectionError:
        fail("Cannot connect to API", f"Is the server running at {BASE_URL}?")
        print(f"\n{RED}Cannot reach API — aborting remaining tests.{RESET}")
        sys.exit(1)


def test_upload_valid():
    section("Upload — Valid NF-es")

    for filename in ["nfe_compra.xml", "nfe_venda.xml", "nfe_nao_identificada.xml"]:
        path = EXAMPLES_DIR / filename
        if not path.exists():
            fail(f"Example file not found: {filename}", f"Expected at {path}")
            continue

        with open(path, "rb") as f:
            r = post_files({"files": (filename, f, "text/xml")})

        check(r.status_code == 202,
              f"POST {filename} returns 202",
              f"POST {filename} returned {r.status_code}",
              r.text[:200])

        data = r.json()
        results = data.get("data", {}).get("results", [])
        if results:
            check(results[0].get("id") and not results[0].get("error"),
                  f"{filename} enqueued successfully (id={results[0].get('id', '')[:8]}...)",
                  f"{filename} enqueue failed: {results[0].get('error')}")


def test_processing():
    section("Processing — Wait & Verify")

    print(f"  {YELLOW}waiting 3s for async processing...{RESET}")
    wait_for_processing(3)

    r = get("/nfe")
    check(r.status_code == 200, "GET /nfe returns 200", f"Got {r.status_code}")

    data = r.json()
    nfes = data.get("data", [])
    total = data.get("total", 0)

    check(total >= 3,
          f"At least 3 NF-es processed (got {total})",
          f"Expected ≥3 processed NF-es, got {total}")

    operations = {n["operation"] for n in nfes}
    check("purchase" in operations,
          "At least one 'purchase' operation classified",
          "No 'purchase' operation found")
    check("sale" in operations,
          "At least one 'sale' operation classified",
          "No 'sale' operation found")
    check("unidentified" in operations,
          "At least one 'unidentified' operation classified",
          "No 'unidentified' operation found")

    statuses = {n["status"] for n in nfes}
    check("error" not in statuses,
          "No error records in main listing (errors go to quarantine)",
          f"Found 'error' status in /nfe — should be in /nfe/quarantine")


def test_summary():
    section("Client Summary")

    r = get("/nfe/summary")
    check(r.status_code == 200, "GET /nfe/summary returns 200", f"Got {r.status_code}")

    data = r.json()
    rows = data.get("data", [])
    check(len(rows) > 0,
          f"Summary has {len(rows)} client row(s)",
          "Summary is empty — expected at least one client")

    if rows:
        row = rows[0]
        check("client" in row and "purchases" in row and "sales" in row,
              f"Summary row has correct fields: {row}",
              f"Missing fields in summary row: {row}")


def test_unidentified():
    section("Unidentified NF-es")

    r = get("/nfe/unidentified")
    check(r.status_code == 200, "GET /nfe/unidentified returns 200", f"Got {r.status_code}")

    data = r.json()
    items = data.get("data", [])
    check(len(items) >= 1,
          f"Found {len(items)} unidentified NF-e(s)",
          "Expected at least 1 unidentified NF-e")

    for item in items:
        check(item.get("operation") == "unidentified",
              f"Item {item.get('id','?')[:8]}... has operation=unidentified",
              f"Unexpected operation: {item.get('operation')}")
        check(item.get("status") != "error",
              f"Unidentified item has status={item.get('status')} (not error)",
              "Unidentified NF-e should not have status=error")


def test_quarantine_invalid_content():
    section("Quarantine — Invalid Content")

    with tempfile.NamedTemporaryFile(suffix=".xml", mode="w", delete=False) as f:
        f.write("this is not xml at all")
        tmp = f.name

    try:
        with open(tmp, "rb") as f:
            r = post_files({"files": ("invalid.xml", f, "text/xml")})

        check(r.status_code == 202, "Upload of invalid XML returns 202", f"Got {r.status_code}")
        data = r.json()
        results = data.get("data", {}).get("results", [])
        if results:
            # Handler pre-check should reject it before enqueuing
            check(results[0].get("error") is not None,
                  f"Invalid content rejected at handler: {results[0].get('error')}",
                  "Expected handler to reject non-XML content")
    finally:
        os.unlink(tmp)


def test_quarantine_malformed_xml():
    section("Quarantine — Malformed NF-e XML")

    # Valid XML but not a valid NF-e (wrong structure)
    bad_nfe = b"""<?xml version="1.0" encoding="UTF-8"?>
<nfeProc xmlns="http://www.portalfiscal.inf.br/nfe" versao="4.00">
  <NFe>
    <infNFe Id="NFe00000000000000000000000000000000000000000000" versao="4.00">
      <emit><CNPJ>00000000000000</CNPJ></emit>
    </infNFe>
  </NFe>
</nfeProc>"""

    with tempfile.NamedTemporaryFile(suffix=".xml", delete=False) as f:
        f.write(bad_nfe)
        tmp = f.name

    try:
        with open(tmp, "rb") as f:
            r = post_files({"files": ("bad_nfe.xml", f, "text/xml")})

        check(r.status_code == 202, "Upload of malformed NF-e returns 202", f"Got {r.status_code}")

        data = r.json()
        results = data.get("data", {}).get("results", [])
        upload_id = results[0].get("id") if results else None

        if upload_id:
            ok(f"Malformed NF-e enqueued (id={upload_id[:8]}...)")
            print(f"  {YELLOW}waiting 2s for consumer to process...{RESET}")
            wait_for_processing(2)

            r2 = get("/nfe/quarantine")
            items = r2.json().get("data", [])
            match = any(i.get("upload_id") == upload_id for i in items)
            check(match,
                  "Malformed NF-e found in quarantine",
                  "Malformed NF-e NOT found in quarantine",
                  f"upload_id={upload_id[:8]}...")
        else:
            # Handler rejected it (pre-check caught it)
            check(results[0].get("error") is not None,
                  f"Malformed NF-e rejected at handler: {results[0].get('error')}",
                  "Unexpected: no id and no error")
    finally:
        os.unlink(tmp)


def test_file_size_limit():
    section("Upload — File Size Limit (> 1 MB)")

    big = b"<?xml version='1.0'?><nfeProc>" + b"A" * (1024 * 1024 + 1) + b"</nfeProc>"

    with tempfile.NamedTemporaryFile(suffix=".xml", delete=False) as f:
        f.write(big)
        tmp = f.name

    try:
        with open(tmp, "rb") as f:
            r = post_files({"files": ("big.xml", f, "text/xml")})

        check(r.status_code == 202, "Request returns 202", f"Got {r.status_code}")
        data = r.json()
        results = data.get("data", {}).get("results", [])
        if results:
            check("too large" in (results[0].get("error") or "").lower(),
                  f"File too large error returned: {results[0].get('error')}",
                  f"Expected size error, got: {results[0]}")
    finally:
        os.unlink(tmp)


def test_file_count_limit():
    section("Upload — File Count Limit (> 10 files)")

    files = []
    tmp_files = []
    try:
        for i in range(11):
            f = tempfile.NamedTemporaryFile(suffix=".xml", mode="w", delete=False)
            f.write("<?xml version='1.0'?><nfeProc/>")
            f.close()
            tmp_files.append(f.name)
            files.append(("files", (f"file_{i}.xml", open(f.name, "rb"), "text/xml")))

        r = requests.post(f"{API}/xml/upload", files=files, timeout=10)
        check(r.status_code == 400,
              "11 files returns 400 Too Many Files",
              f"Expected 400, got {r.status_code}",
              r.text[:200])
    finally:
        for _, (_, fobj, _) in files:
            fobj.close()
        for p in tmp_files:
            os.unlink(p)


def test_rate_limit():
    section("Rate Limiting (POST /xml/upload)")

    statuses = []
    path = EXAMPLES_DIR / "nfe_compra.xml"

    if not path.exists():
        fail("nfe_compra.xml not found — skipping rate limit test")
        return

    print(f"  {YELLOW}sending 15 rapid requests...{RESET}")
    for _ in range(15):
        with open(path, "rb") as f:
            r = requests.post(f"{API}/xml/upload",
                              files={"files": ("nfe_compra.xml", f, "text/xml")},
                              timeout=10)
        statuses.append(r.status_code)

    hit_429 = 429 in statuses
    check(hit_429,
          f"Rate limiter triggered 429 after {statuses.index(429)+1} requests",
          "Rate limiter did NOT trigger — check RATE_LIMIT_BURST and RATE_LIMIT_RPS",
          f"Statuses: {statuses}")


def test_clients():
    section("Internal Clients")

    r = get("/clients")
    check(r.status_code == 200, "GET /clients returns 200", f"Got {r.status_code}")

    data = r.json()
    clients = data.get("data", [])
    check(len(clients) == 5,
          f"5 internal clients returned",
          f"Expected 5 clients, got {len(clients)}")

    if clients:
        c = clients[0]
        check(all(k in c for k in ("id", "name", "cnpj")),
              f"Client has correct fields: {c}",
              f"Missing fields in client: {c}")


def test_non_xml_file():
    section("Upload — Non-XML File")

    with tempfile.NamedTemporaryFile(suffix=".xml", mode="w", delete=False) as f:
        f.write("<html><body>not an nfe</body></html>")
        tmp = f.name

    try:
        with open(tmp, "rb") as f:
            r = post_files({"files": ("fake.xml", f, "text/xml")})

        check(r.status_code == 202, "Returns 202", f"Got {r.status_code}")
        data = r.json()
        results = data.get("data", {}).get("results", [])
        if results:
            has_error = results[0].get("error") is not None
            check(has_error,
                  f"Non-NF-e XML rejected: {results[0].get('error')}",
                  "Non-NF-e XML was not rejected")
    finally:
        os.unlink(tmp)


# ── Main ──────────────────────────────────────────────────────────────────────

def main():
    global BASE_URL, API

    parser = argparse.ArgumentParser(description="NF-e Processor smoke tests")
    parser.add_argument("--base-url", default="http://localhost:8080",
                        help="API base URL (default: http://localhost:8080)")
    args = parser.parse_args()

    BASE_URL = args.base_url.rstrip("/")
    API = f"{BASE_URL}/api/v1"

    print(f"\n{BOLD}NF-e Processor — Smoke Test{RESET}")
    print(f"Target: {BASE_URL}")
    print("=" * 55)

    test_health()
    test_clients()
    test_upload_valid()
    test_processing()
    test_summary()
    test_unidentified()
    test_quarantine_invalid_content()
    test_quarantine_malformed_xml()
    test_file_size_limit()
    test_file_count_limit()
    test_rate_limit()  # last — exhausts the rate limit bucket

    # ── Summary ───────────────────────────────────────────────────────────────
    total = passed + failed
    print(f"\n{'=' * 55}")
    print(f"{BOLD}Results: {passed}/{total} passed{RESET}", end="  ")
    if failed == 0:
        print(f"{GREEN}All tests passed ✓{RESET}")
    else:
        print(f"{RED}{failed} failed ✗{RESET}")
    print()

    sys.exit(0 if failed == 0 else 1)


if __name__ == "__main__":
    main()
