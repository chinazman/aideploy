#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
aideploy - AI 原型快速部署工具（Python 独立客户端）

仅依赖 Python 3 标准库，不依赖原有 Go 客户端。
与 Go 客户端共用 ~/.aideploy/config.json 和 ~/.aideploy/tracking/ 目录，可互操作。

用法示例:
  python aideploy.py config --json
  python aideploy.py config-set server http://localhost:8080/api
  python aideploy.py config-set username admin
  python aideploy.py config-set password admin123
  python aideploy.py config-set site my-prototype F:/work/demo/dist
  python aideploy.py create my-prototype --desc "首页原型"
  python aideploy.py list --json
  python aideploy.py deploy my-prototype -m "第一次发布"
  python aideploy.py deploy my-prototype --dir ./dist --mode full
  python aideploy.py versions my-prototype --json
  python aideploy.py rollback my-prototype abc1234 --yes -m "回滚测试"
  python aideploy.py pull my-prototype
  python aideploy.py delete my-prototype --yes
"""

import ssl
import argparse
import hashlib
import io
import json
import os
import shutil
import sys
import tarfile
import time
import urllib.error
import urllib.parse
import urllib.request

_SSL_CTX = None  # insecure=true 时为跳过证书校验的 context

HOME = os.path.expanduser("~")
CONFIG_DIR = os.path.join(HOME, ".aideploy")
CONFIG_PATH = os.path.join(CONFIG_DIR, "config.json")
TRACKING_DIR = os.path.join(CONFIG_DIR, "tracking")
DEFAULT_SERVER = "http://localhost:8080/api"


# ---------------------------------------------------------------------------
# 配置
# ---------------------------------------------------------------------------

def default_config():
    return {"server_url": DEFAULT_SERVER, "username": "", "password": "", "api_key": "", "site_paths": {}}


def load_config():
    try:
        with open(CONFIG_PATH, "r", encoding="utf-8") as f:
            cfg = json.load(f)
    except (OSError, ValueError):
        return default_config()
    if not cfg.get("server_url"):
        cfg["server_url"] = DEFAULT_SERVER
    if not isinstance(cfg.get("site_paths"), dict):
        cfg["site_paths"] = {}
    return cfg


def save_config(cfg):
    os.makedirs(CONFIG_DIR, exist_ok=True)
    with open(CONFIG_PATH, "w", encoding="utf-8") as f:
        json.dump(cfg, f, ensure_ascii=False, indent=2)


# ---------------------------------------------------------------------------
# HTTP
# ---------------------------------------------------------------------------

def auth_headers(cfg):
    headers = {}
    if cfg.get("username") and cfg.get("password"):
        headers["X-Username"] = cfg["username"]
        headers["X-Password"] = cfg["password"]
    elif cfg.get("api_key"):
        headers["X-API-Key"] = cfg["api_key"]
    return headers


def request_json(cfg, method, path, body=None, query=None):
    """发送 JSON 请求，返回解析后的对象。失败时抛 SystemExit。"""
    url = cfg["server_url"].rstrip("/") + path
    if query:
        url += "?" + urllib.parse.urlencode(query)
    data = None
    headers = auth_headers(cfg)
    if body is not None:
        data = json.dumps(body, ensure_ascii=False).encode("utf-8")
        headers["Content-Type"] = "application/json"
    req = urllib.request.Request(url, data=data, headers=headers, method=method)
    return _do_request(req, expect_json=True)


def request_stream(cfg, path, query=None):
    """发送 GET 请求，返回原始字节。"""
    url = cfg["server_url"].rstrip("/") + path
    if query:
        url += "?" + urllib.parse.urlencode(query)
    req = urllib.request.Request(url, headers=auth_headers(cfg), method="GET")
    return _do_request(req, expect_json=False)


def multipart_upload(cfg, path, fields, file_field, filename, file_bytes):
    """multipart/form-data 上传（手工构建，不依赖 requests）。"""
    boundary = "----aideploy" + hashlib.md5(str(time.time()).encode()).hexdigest()
    buf = io.BytesIO()
    for k, v in fields.items():
        buf.write(("--%s\r\n" % boundary).encode())
        buf.write(('Content-Disposition: form-data; name="%s"\r\n\r\n' % k).encode())
        buf.write(str(v).encode("utf-8"))
        buf.write(b"\r\n")
    buf.write(("--%s\r\n" % boundary).encode())
    buf.write(('Content-Disposition: form-data; name="%s"; filename="%s"\r\n' % (file_field, filename)).encode())
    buf.write(b"Content-Type: application/octet-stream\r\n\r\n")
    buf.write(file_bytes)
    buf.write(("\r\n--%s--\r\n" % boundary).encode())

    url = cfg["server_url"].rstrip("/") + path
    headers = auth_headers(cfg)
    headers["Content-Type"] = "multipart/form-data; boundary=%s" % boundary
    req = urllib.request.Request(url, data=buf.getvalue(), headers=headers, method="POST")
    return _do_request(req, expect_json=True)


def _do_request(req, expect_json):
    try:
        with urllib.request.urlopen(req, timeout=300, context=_SSL_CTX) as resp:
            raw = resp.read()
            if expect_json:
                try:
                    return json.loads(raw.decode("utf-8"))
                except ValueError:
                    return {"raw": raw.decode("utf-8", "replace")}
            return raw
    except urllib.error.HTTPError as e:
        body = ""
        try:
            body = e.read().decode("utf-8", "replace")
        except Exception:
            pass
        try:
            err = json.loads(body).get("error", body)
        except ValueError:
            err = body
        fail("服务器错误 (HTTP %d): %s" % (e.code, err))
    except urllib.error.URLError as e:
        fail("无法连接服务器 %s: %s" % (req.full_url, e.reason))


def fail(msg, code=1):
    print(json.dumps({"ok": False, "error": msg}, ensure_ascii=False), file=sys.stderr)
    sys.exit(code)


# ---------------------------------------------------------------------------
# 文件扫描 / 打包
# ---------------------------------------------------------------------------

def scan_files(site_path):
    """扫描目录，返回 {相对路径: {path, hash, size}}。跳过隐藏文件/目录。"""
    files = {}
    site_path = os.path.abspath(site_path)
    for root, dirs, names in os.walk(site_path):
        dirs[:] = [d for d in dirs if not d.startswith(".")]
        for n in names:
            if n.startswith("."):
                continue
            full = os.path.join(root, n)
            rel = os.path.relpath(full, site_path).replace(os.sep, "/")
            h = hashlib.md5()
            with open(full, "rb") as f:
                for chunk in iter(lambda: f.read(65536), b""):
                    h.update(chunk)
            files[rel] = {"path": rel, "hash": h.hexdigest(), "size": os.path.getsize(full)}
    return files


def create_package(site_path, rel_paths):
    """将指定文件打包为 tar.gz（内存中），路径使用 POSIX 分隔符，与 Go 服务端兼容。"""
    buf = io.BytesIO()
    with tarfile.open(fileobj=buf, mode="w:gz") as tf:
        for rel in rel_paths:
            full = os.path.join(os.path.abspath(site_path), rel.replace("/", os.sep))
            if not os.path.isfile(full):
                continue
            info = tarfile.TarInfo(rel)
            info.size = os.path.getsize(full)
            info.mtime = int(os.path.getmtime(full))
            info.mode = 0o644
            with open(full, "rb") as f:
                tf.addfile(info, f)
    return buf.getvalue()


# ---------------------------------------------------------------------------
# 跟踪信息（与 Go 客户端 ~/.aideploy/tracking/<site>.json 兼容）
# ---------------------------------------------------------------------------

def now_iso():
    return time.strftime("%Y-%m-%dT%H:%M:%S") + (".%09d" % (time.time_ns() % 10**9)) + "+08:00"


def tracking_path(site):
    return os.path.join(TRACKING_DIR, site + ".json")


def load_tracking(site):
    try:
        with open(tracking_path(site), "r", encoding="utf-8") as f:
            return json.load(f)
    except (OSError, ValueError):
        return None


def save_tracking(site, files):
    os.makedirs(TRACKING_DIR, exist_ok=True)
    data = {
        "site_name": site,
        "last_sync": now_iso(),
        "files": [
            {
                "path": f["path"],
                "hash": f["hash"],
                "size": f["size"],
                "mod_time": f.get("mod_time", now_iso()),
                "last_deployed": f.get("last_deployed", now_iso()),
            }
            for f in files.values()
        ],
    }
    with open(tracking_path(site), "w", encoding="utf-8") as f:
        json.dump(data, f, ensure_ascii=False, indent=2)


# ---------------------------------------------------------------------------
# 子命令
# ---------------------------------------------------------------------------

def emit(payload, args, human):
    if getattr(args, "json", False):
        print(json.dumps(payload, ensure_ascii=False, indent=2))
    else:
        print(human)


def cmd_config(args):
    cfg = load_config()
    out = {
        "config_path": CONFIG_PATH,
        "server_url": cfg["server_url"],
        "username": cfg.get("username", ""),
        "password_set": bool(cfg.get("password")),
        "api_key_set": bool(cfg.get("api_key")),
        "site_paths": cfg["site_paths"],
    }
    emit(out, args, "当前配置:\n  服务器地址: %s\n  用户名: %s\n  网站绑定目录: %s\n  配置文件: %s" % (
        cfg["server_url"], cfg.get("username") or "(未设置)",
        json.dumps(cfg["site_paths"], ensure_ascii=False) if cfg["site_paths"] else "(未配置)", CONFIG_PATH))


def cmd_config_set(args):
    cfg = load_config()
    key, values = args.key, args.value
    if key == "server":
        if not values:
            fail("缺少服务器地址，用法: config-set server <url>")
        cfg["server_url"] = values[0]
        msg = "服务器地址已设置: %s" % values[0]
    elif key == "username":
        if not values:
            fail("缺少用户名，用法: config-set username <name>")
        cfg["username"] = values[0]
        msg = "用户名已设置: %s" % values[0]
    elif key == "password":
        if not values:
            fail("缺少密码，用法: config-set password <pwd>")
        cfg["password"] = values[0]
        msg = "密码已设置"
    elif key == "api-key":
        if not values:
            fail("缺少 API Key，用法: config-set api-key <key>")
        cfg["api_key"] = values[0]
        msg = "API Key 已设置"
    elif key == "insecure":
        if not values or values[0] not in ("true", "false"):
            fail("用法: config-set insecure <true|false>")
        cfg["insecure"] = values[0] == "true"
        msg = "TLS 证书校验已%s" % ("跳过" if cfg["insecure"] else "开启")
    elif key == "site":
        if len(values) < 2:
            fail("用法: config-set site <name> <directory>")
        site, path = values[0], os.path.abspath(values[1])
        if not os.path.isdir(path):
            fail("目录不存在: %s" % path)
        cfg["site_paths"][site] = path
        msg = "网站 '%s' 已绑定目录: %s" % (site, path)
    else:
        fail("未知配置项 '%s'，支持: server / username / password / api-key / site" % key)
    save_config(cfg)
    emit({"ok": True, "message": msg}, args, "✓ " + msg)


def cmd_config_remove(args):
    cfg = load_config()
    site = args.value
    if site not in cfg["site_paths"]:
        fail("网站 '%s' 未绑定目录" % site)
    del cfg["site_paths"][site]
    save_config(cfg)
    msg = "已移除网站 '%s' 的目录绑定" % site
    emit({"ok": True, "message": msg}, args, "✓ " + msg)


def cmd_test(args):
    cfg = load_config()
    result = request_json(cfg, "GET", "/sites/list")
    sites = result.get("sites", []) if isinstance(result, dict) else []
    emit({"ok": True, "server": cfg["server_url"], "auth_user": cfg.get("username", ""),
          "site_count": len(sites)},
         args, "✓ 连接成功: %s（可见 %d 个网站，用户: %s）" % (cfg["server_url"], len(sites), cfg.get("username") or "(匿名)"))


def cmd_create(args):
    cfg = load_config()
    result = request_json(cfg, "POST", "/sites/create", {"name": args.name, "desc": args.desc or ""})
    emit(result, args, "✓ 网站创建成功!\n  名称: %s\n  地址: %s" % (result.get("name"), result.get("url") or result.get("domain")))


def cmd_delete(args):
    if not args.yes:
        fail("删除是危险操作，请加 --yes 确认执行")
    cfg = load_config()
    request_json(cfg, "POST", "/sites/delete", {"name": args.name})
    emit({"ok": True, "message": "删除成功"}, args, "✓ 网站 '%s' 删除成功" % args.name)


def cmd_update(args):
    cfg = load_config()
    body = {"name": args.name, "desc": args.desc or "",
            "users": [u.strip() for u in (args.users or "").split(",") if u.strip()]}
    request_json(cfg, "POST", "/sites/update", body)
    emit({"ok": True, "message": "更新成功"}, args, "✓ 网站 '%s' 信息已更新" % args.name)


def cmd_list(args):
    cfg = load_config()
    result = request_json(cfg, "GET", "/sites/list")
    sites = result.get("sites", []) if isinstance(result, dict) else []
    emit({"sites": sites}, args,
         "网站列表 (%d):\n" % len(sites) + "\n".join(
             "  %-20s %s" % (s.get("name", ""), s.get("url") or s.get("domain", "")) for s in sites))


def cmd_versions(args):
    cfg = load_config()
    versions = request_json(cfg, "GET", "/sites/versions", query={"name": args.name})
    if versions:
        human = "网站 '%s' 版本历史:\n" % args.name + "\n".join(
            "  %s  %s  (%s)" % (v.get("hash", "")[:8], v.get("message", ""), v.get("date", ""))
            for v in versions)
    else:
        human = "网站 '%s' 版本历史: 暂无版本记录" % args.name
    emit({"versions": versions}, args, human)


def cmd_rollback(args):
    if not args.yes:
        fail("回滚会覆盖线上版本，请加 --yes 确认执行")
    cfg = load_config()
    request_json(cfg, "POST", "/sites/rollback",
                 {"name": args.name, "hash": args.hash, "message": args.message or "回滚版本"})
    emit({"ok": True, "message": "回滚成功"}, args, "✓ 网站 '%s' 已回滚到 %s" % (args.name, args.hash))


def resolve_dir(args, site, cfg, create=False):
    dir_path = args.dir or cfg["site_paths"].get(site)
    if not dir_path:
        fail("网站 '%s' 未绑定目录，请先执行: aideploy.py config-set site %s <目录>，或使用 --dir 参数" % (site, site))
    if not os.path.isdir(dir_path):
        if not create:
            fail("目录不存在: %s" % dir_path)
        os.makedirs(dir_path, exist_ok=True)
    return dir_path


def cmd_deploy(args):
    cfg = load_config()
    site = args.name
    dir_path = resolve_dir(args, site, cfg)
    current = scan_files(dir_path)

    mode = args.mode
    tracking = load_tracking(site)
    if mode == "auto":
        mode = "incremental" if tracking else "full"

    if mode == "incremental":
        prev = {f["path"]: f["hash"] for f in (tracking or {}).get("files", [])}
        changed = [f["path"] for f in current.values() if prev.get(f["path"]) != f["hash"]]
        if not changed:
            emit({"ok": True, "mode": "incremental", "changed": 0,
                  "message": "没有文件变更，无需部署"}, args, "✓ 没有文件变更，无需部署")
            return
        pkg = create_package(dir_path, changed)
        endpoint = "/sites/deploy-incremental"
    else:
        changed = list(current.keys())
        pkg = create_package(dir_path, changed)
        endpoint = "/sites/deploy-full"

    result = multipart_upload(cfg, endpoint,
                              {"name": site, "message": args.message or ""}, "package",
                              "deploy.tar.gz", pkg)
    save_tracking(site, current)
    out = {"ok": True, "mode": mode, "deployed_files": len(changed), "server_response": result}
    emit(out, args, "✓ %s部署成功 (%d 个文件) — 网站: %s" % ("全量" if mode == "full" else "增量", len(changed), site))


def cmd_pull(args):
    cfg = load_config()
    site = args.name
    dir_path = resolve_dir(args, site, cfg, create=True)
    data = request_stream(cfg, "/sites/export", query={"name": site})

    os.makedirs(dir_path, exist_ok=True)
    # 清空本地目录（保留隐藏文件），与 Go 客户端行为一致
    for entry in os.listdir(dir_path):
        if entry.startswith("."):
            continue
        full = os.path.join(dir_path, entry)
        if os.path.isdir(full) and not os.path.islink(full):
            shutil.rmtree(full, ignore_errors=True)
        else:
            os.remove(full)

    # 解压（带路径安全检查）
    import gzip
    count = 0
    with gzip.GzipFile(fileobj=io.BytesIO(data)) as gz:
        with tarfile.open(fileobj=gz, mode="r:") as tf:
            dest_abs = os.path.abspath(dir_path)
            for member in tf.getmembers():
                if not member.isfile():
                    continue
                target = os.path.abspath(os.path.join(dest_abs, member.name.replace("/", os.sep)))
                if not (target == dest_abs or target.startswith(dest_abs + os.sep)):
                    fail("部署包中包含非法路径: %s" % member.name)
                os.makedirs(os.path.dirname(target), exist_ok=True)
                with tf.extractfile(member) as src, open(target, "wb") as dst:
                    dst.write(src.read())
                count += 1

    save_tracking(site, scan_files(dir_path))
    emit({"ok": True, "pulled_files": count}, args, "✓ 已从服务器下载 %d 个文件到 %s" % (count, dir_path))


# ---------------------------------------------------------------------------
# 入口
# ---------------------------------------------------------------------------

def add_common(p, with_site=True):
    p.add_argument("--json", action="store_true", help="以 JSON 格式输出（便于程序解析）")
    if with_site:
        p.add_argument("--dir", help="网站本地目录（未绑定 site_paths 时使用）")
        p.add_argument("-m", "--message", help="版本说明")


def main():
    if sys.platform == "win32":
        try:
            sys.stdout.reconfigure(encoding="utf-8", errors="replace")
            sys.stderr.reconfigure(encoding="utf-8", errors="replace")
        except Exception:
            pass

    global _SSL_CTX
    if load_config().get("insecure"):
        _SSL_CTX = ssl.create_default_context()
        _SSL_CTX.check_hostname = False
        _SSL_CTX.verify_mode = ssl.CERT_NONE

    ap = argparse.ArgumentParser(prog="aideploy", description="AI 原型快速部署工具（Python 客户端）")
    sub = ap.add_subparsers(dest="command", required=True)

    p = sub.add_parser("config", help="查看当前配置")
    add_common(p, with_site=False)
    p.set_defaults(func=cmd_config)

    p = sub.add_parser("config-set", help="设置配置项: server/username/password/api-key/site")
    p.add_argument("key", choices=["server", "username", "password", "api-key", "insecure", "site"])
    p.add_argument("value", nargs="*")
    add_common(p, with_site=False)
    p.set_defaults(func=cmd_config_set)

    p = sub.add_parser("config-remove", help="移除网站目录绑定")
    p.add_argument("value", help="网站名称")
    add_common(p, with_site=False)
    p.set_defaults(func=cmd_config_remove)

    p = sub.add_parser("test", help="测试服务器连接与认证")
    add_common(p, with_site=False)
    p.set_defaults(func=cmd_test)

    p = sub.add_parser("create", help="创建网站")
    p.add_argument("name")
    p.add_argument("--desc", help="网站描述")
    add_common(p, with_site=False)
    p.set_defaults(func=cmd_create)

    p = sub.add_parser("delete", help="删除网站（需 --yes）")
    p.add_argument("name")
    p.add_argument("--yes", action="store_true")
    add_common(p, with_site=False)
    p.set_defaults(func=cmd_delete)

    p = sub.add_parser("update", help="更新网站描述/授权用户")
    p.add_argument("name")
    p.add_argument("--desc")
    p.add_argument("--users", help="授权用户列表，逗号分隔")
    add_common(p, with_site=False)
    p.set_defaults(func=cmd_update)

    p = sub.add_parser("list", help="列出所有网站")
    add_common(p, with_site=False)
    p.set_defaults(func=cmd_list)

    p = sub.add_parser("versions", help="查看版本历史")
    p.add_argument("name")
    add_common(p, with_site=False)
    p.set_defaults(func=cmd_versions)

    p = sub.add_parser("rollback", help="回滚到指定版本（需 --yes）")
    p.add_argument("name")
    p.add_argument("hash")
    p.add_argument("--yes", action="store_true")
    add_common(p, with_site=False)
    p.set_defaults(func=cmd_rollback)

    p = sub.add_parser("deploy", help="发布网站（auto: 有跟踪信息走增量，否则全量）")
    p.add_argument("name")
    p.add_argument("--mode", choices=["auto", "full", "incremental"], default="auto")
    add_common(p)
    p.set_defaults(func=cmd_deploy)

    p = sub.add_parser("pull", help="从服务器下载网站并覆盖本地目录")
    p.add_argument("name")
    add_common(p)
    p.set_defaults(func=cmd_pull)

    args = ap.parse_args()
    args.func(args)


if __name__ == "__main__":
    main()
