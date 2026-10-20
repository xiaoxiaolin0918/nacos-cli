# -*- coding: utf-8 -*-
from __future__ import annotations

import threading
import urllib.parse
import urllib.request

from nacos_cli.infrastructure.web_login import form_html, run_web_login


def test_form_html_escapes():
    html = form_html('http://x<script>', 'a"b', 'bad<x>')
    assert "<script>" not in html
    assert "bad<x>" not in html
    assert "bad&lt;x&gt;" in html


def test_run_web_login_posts_and_returns():
    def do_login(user, password):
        assert user == "nacos"
        assert password == "p@ss"
        return {"auth_mode": "token", "has_token": True}

    def open_browser(url: str) -> None:
        def post():
            data = urllib.parse.urlencode(
                {"username": "nacos", "password": "p@ss"}
            ).encode()
            req = urllib.request.Request(url, data=data, method="POST")
            with urllib.request.urlopen(req, timeout=5) as resp:
                body = resp.read().decode()
                assert "Login saved" in body

        threading.Thread(target=post, daemon=True).start()

    out = run_web_login(
        server_hint="http://127.0.0.1:8848/nacos",
        prefill_user="nacos",
        do_login=do_login,
        open_browser=open_browser,
        timeout_s=10,
    )
    assert out["auth_mode"] == "token"
    assert out["via"] == "web"
    assert "p@ss" not in str(out)
