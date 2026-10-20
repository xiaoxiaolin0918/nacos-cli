# -*- coding: utf-8 -*-
"""Local loopback web login form (plumelog-style). Binds 127.0.0.1 only."""
from __future__ import annotations

import html
import secrets
import threading
import webbrowser
from http.server import BaseHTTPRequestHandler, HTTPServer
from typing import Callable
from urllib.parse import parse_qs

WEB_LOGIN_TIMEOUT_S = 120


def open_system_browser(url: str) -> None:
    webbrowser.open(url)


def form_html(server_hint: str, username: str = "", error: str = "") -> str:
    err = (
        f'<p style="color:#b00020">{html.escape(error)}</p>' if error else ""
    )
    return f"""<!doctype html>
<html><head><meta charset="utf-8"><title>nacos-cli login</title>
<style>
body{{font-family:system-ui,sans-serif;max-width:420px;margin:40px auto;padding:0 16px}}
label{{display:block;margin:12px 0 4px}}
input{{width:100%;padding:8px;box-sizing:border-box}}
button{{margin-top:16px;padding:10px 16px}}
.hint{{color:#555;font-size:13px}}
</style></head><body>
<h1>nacos-cli web login</h1>
<p class="hint">Target: <code>{html.escape(server_hint)}</code></p>
<p class="hint">Password stays in this form; the CLI never prints it.</p>
{err}
<form method="POST">
<label>Username</label>
<input name="username" value="{html.escape(username)}" autocomplete="username" required>
<label>Password</label>
<input name="password" type="password" autocomplete="current-password" required>
<button type="submit">Sign in</button>
</form>
</body></html>"""


def ok_html(server: str, auth_mode: str) -> str:
    return f"""<!doctype html>
<html><head><meta charset="utf-8"><title>nacos-cli login ok</title></head>
<body style="font-family:system-ui;margin:40px">
<h1>Login saved</h1>
<p>Server: <code>{html.escape(server)}</code></p>
<p>Auth mode: <code>{html.escape(auth_mode)}</code></p>
<p>You can close this tab.</p>
</body></html>"""


def fail_html(message: str) -> str:
    return f"""<!doctype html>
<html><head><meta charset="utf-8"><title>nacos-cli login failed</title></head>
<body style="font-family:system-ui;margin:40px">
<h1>Login failed</h1>
<p style="color:#b00020">{html.escape(message)}</p>
</body></html>"""


def run_web_login(
    *,
    server_hint: str,
    prefill_user: str,
    do_login: Callable[[str, str], dict],
    open_browser: Callable[[str], None] | None = None,
    timeout_s: float = WEB_LOGIN_TIMEOUT_S,
    on_ready: Callable[[str], None] | None = None,
) -> dict:
    """Serve a one-shot login form on 127.0.0.1 and return login result dict.

    do_login(username, password) must return a dict with at least auth_mode.
    Raises TimeoutError or RuntimeError on failure.
    """
    token = secrets.token_hex(16)
    form_path = f"/cli-login/{token}/"
    outcome: dict = {}
    err_box: list[BaseException] = []
    done = threading.Event()

    class Handler(BaseHTTPRequestHandler):
        def log_message(self, fmt: str, *args) -> None:  # silence
            return

        def do_GET(self) -> None:  # noqa: N802
            if self.path.rstrip("/") != form_path.rstrip("/"):
                self.send_error(404)
                return
            body = form_html(server_hint, prefill_user).encode("utf-8")
            self.send_response(200)
            self.send_header("Content-Type", "text/html; charset=utf-8")
            self.send_header("Cache-Control", "no-store")
            self.send_header("Content-Length", str(len(body)))
            self.end_headers()
            self.wfile.write(body)

        def do_POST(self) -> None:  # noqa: N802
            if self.path.rstrip("/") != form_path.rstrip("/"):
                self.send_error(404)
                return
            length = int(self.headers.get("Content-Length", "0") or 0)
            raw = self.rfile.read(length).decode("utf-8", errors="replace")
            form = parse_qs(raw, keep_blank_values=True)
            user = (form.get("username") or [""])[0].strip()
            password = (form.get("password") or [""])[0]
            try:
                result = do_login(user, password)
                body = ok_html(server_hint, str(result.get("auth_mode") or "")).encode(
                    "utf-8"
                )
                outcome.update(result)
                outcome["via"] = "web"
                outcome["username"] = user
            except Exception as e:  # noqa: BLE001
                err_box.append(e)
                body = fail_html(str(e)).encode("utf-8")
            self.send_response(200)
            self.send_header("Content-Type", "text/html; charset=utf-8")
            self.send_header("Cache-Control", "no-store")
            self.send_header("Content-Length", str(len(body)))
            self.end_headers()
            self.wfile.write(body)
            done.set()

    httpd = HTTPServer(("127.0.0.1", 0), Handler)
    host, port = httpd.server_address
    if host not in ("127.0.0.1", "::1"):
        httpd.server_close()
        raise RuntimeError(f"refusing non-loopback bind {host}")
    form_url = f"http://127.0.0.1:{port}{form_path}"

    thread = threading.Thread(target=httpd.serve_forever, daemon=True)
    thread.start()
    try:
        if on_ready:
            on_ready(form_url)
        opener = open_browser or open_system_browser
        try:
            opener(form_url)
        except Exception:
            pass
        if not done.wait(timeout_s):
            raise TimeoutError("web login timed out; local listener closed")
        if err_box:
            raise RuntimeError(str(err_box[0])) from err_box[0]
        if not outcome:
            raise RuntimeError("web login produced no result")
        return outcome
    finally:
        httpd.shutdown()
        httpd.server_close()
