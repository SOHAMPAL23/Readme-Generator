<div align="center">

# 📄 ReadMeAI

### AI-Powered GitHub README Generator

[![Go](https://img.shields.io/badge/Go-1.24-00ADD8?style=flat-square&logo=go)](https://golang.org)
[![Gin](https://img.shields.io/badge/Gin-Framework-00B4AB?style=flat-square)](https://gin-gonic.com)
[![OpenAI](https://img.shields.io/badge/OpenAI-GPT--4o--mini-412991?style=flat-square&logo=openai)](https://openai.com)
[![Groq](https://img.shields.io/badge/Groq-Fallback--Ready-f55036?style=flat-square)](https://groq.com)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=flat-square&logo=docker)](https://docker.com)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg?style=flat-square)](LICENSE)

**Paste a GitHub URL → Get a professional README in seconds.**

</div>

---

## 🎯 Overview

ReadMeAI is a production-ready SaaS application that automatically generates high-quality README files by analyzing GitHub repositories.

It fetches repository metadata, detects your tech stack from dependency files, calculates a health score, and sends structured data to OpenAI (or fallback providers like Groq) to produce a professional, recruiter-ready README — all in one click.

---

## ✨ Features

- 🔍 **Deep Repository Analysis** — parses `package.json`, `go.mod`, `requirements.txt`, `Cargo.toml`, and more
- 🤖 **Multi-Provider AI** — primary generation via OpenAI GPT-4o-mini, with auto-fallback to Groq (Llama-3.3-70b)
- 🎨 **4 README Styles** — Developer, Portfolio, Startup, Open Source
- 📊 **Health Score** — rates your repo 0-100 with actionable improvement suggestions
- 👀 **Live Preview** — split-screen rendered Markdown with syntax highlighting
- 📥 **One-click Download** — saves as `README.md` ready to commit
- 📋 **Copy to Clipboard** — paste directly into GitHub
- 🌙 **Dark / Light Mode** — persisted in localStorage
- 🚦 **Rate Limiting** — per-IP protection for API keys
- 🐳 **Docker Ready** — multi-stage build, non-root user

---

## 🛠️ Tech Stack

| Layer | Technology |
|---|---|
| Backend | Go 1.24, Gin, go-resty |
| Frontend | Vanilla JS, Marked.js, Highlight.js |
| AI | OpenAI GPT-4o-mini |
| GitHub | GitHub REST API v3 |
| Deployment | Docker, Docker Compose |

---

## 🚀 Quick Start

### Prerequisites
- Go 1.24+
- OpenAI API Key ([get one here](https://platform.openai.com/api-keys))
- GitHub Personal Access Token (optional, but recommended)

### 1. Clone & Configure

```bash
git clone https://github.com/your-username/readmeai
cd readmeai
cp .env.example .env
```

Edit `.env`:

```env
OPENAI_API_KEY=sk-your_key_here
GROQ_API_KEY=gsk_your_key_here     # optional fallback (resiliency against rate limits)
GITHUB_TOKEN=ghp_your_token_here   # optional, raises rate limit to 5000/hr
PORT=8080
OPENAI_MODEL=gpt-4o-mini           # or gpt-4o for higher quality
GROQ_MODEL=llama-3.3-70b-versatile  # default Groq fallback model
```

### 2. Run Locally

```bash
go run ./cmd/main.go
```

Open **http://localhost:8080**

### 3. Run with Docker

```bash
docker compose up --build
```

---

## 📁 Project Structure

```
readmeai/
├── cmd/
│   └── main.go                # Entry point, Gin router, middleware
├── core/
│   ├── github/
│   │   ├── client.go          # GitHub API client (Resty)
│   │   ├── parser.go          # Response parser, health scoring
│   │   └── analyzer.go        # Tech stack detection from file tree
│   ├── ai/
│   │   ├── generator.go       # AI client (OpenAI + Groq fallback)
│   │   └── prompts.go         # 4 style-specific prompt templates
│   ├── handlers/
│   │   ├── generate.go        # POST /api/generate
│   │   ├── repository.go      # GET /api/repository
│   │   └── handlers.go        # Handler utilities
│   ├── services/
│   │   └── readme_service.go  # Orchestration layer
│   ├── models/
│   │   ├── repository.go      # RepoInfo, TechStack, HealthScore
│   │   └── readme.go          # Request/Response structs
│   └── utils/
│       ├── validator.go        # GitHub URL validation
│       └── markdown.go         # Markdown helpers
├── static/
│   ├── css/
│   │   └── style.css          # Full design system (dark/light)
│   └── js/
│       └── app.js             # Vanilla JS SPA logic
├── index.html                 # SPA shell
├── Dockerfile                 # Multi-stage build
├── docker-compose.yml
├── .env.example
├── go.mod
└── go.sum
```

---

## 🌐 API Reference

### `POST /api/generate`

Generate a full README for a repository.

**Request:**
```json
{
  "repo_url": "https://github.com/gin-gonic/gin",
  "style": "developer"
}
```

**Styles:** `developer` | `portfolio` | `startup` | `opensource`

**Response:**
```json
{
  "repository": { "name": "gin", "stars": 79000, "language": "Go", ... },
  "tech_stack": { "backend": ["Go"], "devops": ["Docker"] },
  "health_score": { "score": 88, "suggestions": [...] },
  "readme": "# gin\n\n...",
  "style": "developer",
  "generated_at": "2025-01-01T00:00:00Z"
}
```

### `GET /api/repository?url=<github_url>`

Fetch repository metadata and tech stack without calling AI.

### `GET /health`

Returns service health status.

---

## 🔐 Environment Variables

| Variable | Required | Description |
|---|---|---|
| `OPENAI_API_KEY` | ✅ Yes | OpenAI API key for README generation (either OpenAI or Groq is required) |
| `GROQ_API_KEY` | No | Groq API key for fallback generation (protects against OpenAI 429s) |
| `GITHUB_TOKEN` | ⚠️ Recommended | Raises API limit from 60 to 5000 req/hr |
| `PORT` | No | Server port (default: `8080`) |
| `OPENAI_MODEL` | No | Model to use (default: `gpt-4o-mini`) |
| `GROQ_MODEL` | No | Model to use for Groq fallback (default: `llama-3.3-70b-versatile`) |
| `GIN_MODE` | No | `release` for production |

---

## ☁️ Deployment

### Render

1. Connect your GitHub repo on [render.com](https://render.com)
2. Set build command: `go build -o readmeai ./cmd/main.go`
3. Set start command: `./readmeai`
4. Add environment variables in the Render dashboard

### Railway

```bash
railway login
railway init
railway up
railway variables set OPENAI_API_KEY=sk-...
railway variables set GITHUB_TOKEN=ghp_...
```

### Fly.io

```bash
fly launch
fly secrets set OPENAI_API_KEY=sk-...
fly secrets set GITHUB_TOKEN=ghp_...
fly deploy
```

---

## 🗺️ Roadmap

- [ ] GitHub OAuth — generate READMEs for private repos
- [ ] One-click commit — push README directly to the repository
- [ ] README versioning — compare current vs generated
- [ ] Multi-language generation — Portuguese, Spanish, Chinese
- [ ] Streaming generation — token-by-token live preview
- [ ] SQLite history — track previously generated READMEs

---

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feat/your-feature`
3. Commit your changes: `git commit -m 'feat: add feature'`
4. Push and open a Pull Request

Please follow Go conventions and add tests for new functionality.

---

## 📄 License

MIT License — see [LICENSE](LICENSE) for details.
