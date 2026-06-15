---
title: "Introduction"
description: "What lemmy is and how it is put together."
weight: 10
---

A command line for the Lemmy federated forum.

lemmy is a single binary. It speaks to lemmy.world over plain HTTPS,
shapes the responses into clean records, and gets out of your way. There is
no API key, nothing to sign up for, and nothing to run alongside it.

## What it can do

- **`lemmy posts`** — list posts from lemmy.world, sorted by activity, heat, or newness
- **`lemmy communities`** — list communities by subscriber count or activity
- **`lemmy comments <post-id>`** — list comments on a post
- **`lemmy search <query>`** — search posts across the instance
- **`lemmy site`** — show instance statistics: users, posts, comments, communities

## How it is built

- A **library package** (`lemmy`) holds the HTTP client and the typed
  data models. It paces requests, sets an honest User-Agent, and retries the
  transient failures any public site throws under load.
- A **domain** (`lemmy/domain.go`) declares each operation once on the
  [any-cli/kit](https://github.com/tamnd/any-cli) framework. That single
  declaration becomes a CLI command, an HTTP route, an MCP tool, and a
  resource-URI dereference. It is the one place you add to the tool.
- A thin **`cmd/lemmy`** hands the assembled app to `kit.Run`, which
  builds the command tree and the serve and mcp surfaces.

## Scope

lemmy is a read-only client over data Lemmy already serves
publicly. It reads that data and shapes it for you. That narrow scope keeps it a
single small binary with no database, no daemon, and no setup.

Next: [install it](/getting-started/installation/), then take the
[quick start](/getting-started/quick-start/).
