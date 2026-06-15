---
title: "Quick start"
description: "Fetch your first records with lemmy."
weight: 30
---

Once `lemmy` is on your `PATH`, list the active posts on lemmy.world:

```bash
lemmy posts
```

By default you get an aligned table. Ask for JSON when you want to pipe it:

```bash
$ lemmy posts --limit 3 -o json
[
  {
    "id": 123,
    "title": "Some interesting post",
    "community": "technology",
    "author": "alice",
    "score": 150,
    "comments": 42,
    "published": "2026-06-15T05:00:00.000000",
    "post_url": "https://lemmy.world/post/123"
  }
]
```

## Commands

```bash
lemmy posts --sort Active --limit 20         # active posts
lemmy posts --sort Hot --type Local          # hot local posts
lemmy communities --sort Active --limit 10   # top communities
lemmy comments 123 --limit 20               # comments on post 123
lemmy search "self-hosted" --limit 10        # search posts
lemmy site                                   # instance stats
```

## Shape the output

The same flags work on every command:

```bash
lemmy posts --fields id,title,score         # keep only these columns
lemmy posts -o jsonl | jq .community        # one object per line, into jq
lemmy posts -o csv                          # CSV for spreadsheets
```

`-o` takes `table`, `json`, `jsonl`, `csv`, `tsv`, `url`, or `raw`. Left to
`auto`, it prints a table to a terminal and JSONL into a pipe, so the same
command reads well by hand and parses cleanly downstream.

## Serve it instead

The same operations are available over HTTP and to agents over MCP:

```bash
lemmy serve --addr :7777 &
curl -s 'localhost:7777/v1/posts'    # NDJSON, one record per line
lemmy mcp                            # MCP over stdio
```
