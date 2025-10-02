# 🎵 yt-dlp Stream Resolver API

A lightweight Golang API that uses [yt-dlp](https://github.com/yt-dlp/yt-dlp) to resolve a given video URL (YouTube, etc.) into a direct stream URL.

I built this primarily to power two of my personal projects:

* 📱 A WhatsApp bot
* 🎶 A Discord music bot

But it can be used in any project that needs direct access to media streams without dealing with YouTube’s player logic.

---

## 🚀 Features

* Exposes a simple REST API to resolve media links.
* Uses `yt-dlp` under the hood for wide compatibility.
* Lightweight, containerized with Docker.
* Ready-to-use with `docker-compose`.

---

## 📦 Getting Started

1. **Clone the repository:**
    ```bash
    git clone https://github.com/zwepener/yt-dlp-api.git
    cd yt-dlp-api
    ```

2. **Build with Docker:**<br/>
You can build and run the API using Docker:
    ```bash
    docker build -t yt-dlp-api .
    docker run -p 8080:8080 yt-dlp-api
    ```
    The API will be available at: 👉 `http://localhost:8080`

4. **Using docker-compose**<br/>
    If you prefer using `docker-compose`, simply run:
    ```bash
    docker compose up --build
    ```
    This will also start a `redis` container. The api uses this service to cache resolved urls for a set period.

---

## 🔌 API Usage

**Endpoint**: `POST /resolve`<br/><br/>
**Request Body:**  
The request body must consist of an array of urls.
```json
[
  "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
]
```
**Response:**  
If a given url could not be resolved into its streaming url, it will be omitted from the result.
```json
{
  "https://www.youtube.com/watch?v=dQw4w9WgXcQ": "https://rr2---sn-woc7kn7y.googlevideo.com/videoplayback?..."
}
```

---

## 🛠 Configuration

Environment variables can be set in `.env` or via Docker (see [`.env.template`](.env.template) for all available environment variables).
```bash
cp .env.template .env
nano .env
```

---

## 🔮 Future Plans

Some ideas I’d like to work on in the future:
- [x] Add support for batch URL resolution.
- [x] Implement caching to reduce repeated yt-dlp calls.
- [x] Clean urls before caching for more consistent caching.
- [ ] Hash urls for redis cache keys.
- [ ] Adopt [gin](https://gin-gonic.com/) for easier maintentance.
- [ ] Implement proper logging.
- [ ] Add health checks and better error handling.

---

## 📜 Requirements

> [!IMPORTANT]
> Although it is not strictly required to have a redis service, it is highly recommended as the API uses it to cache resolved urls to reduce calls to the internet (YouTube, Instagram, Facebook, etc.).

### Using Docker

* [Docker Engine](https://docs.docker.com/engine/install/)
* [Docker Compose](https://docs.docker.com/compose/install/linux/) plugin for docker engine (only if you plan on using docker compose)

### No Docker
* [Go 1.22+](https://go.dev/doc/install)
* [yt-dlp](https://github.com/yt-dlp/yt-dlp)

---

## ⚠️ Disclaimer

This project is provided for **educational and personal use only**.

Please note:
* Extracting or streaming content directly from platforms like YouTube and others may **violate their Terms of Service**.
* Abuse of this API against external services could result in your account(s) being banned, or even **legal repercussions**.
* The author(s) of this project assume **no responsibility** for misuse.

---

## 🤝 Contributing
Contributions are welcome! Feel free to open issues or submit PRs if you’d like to add features or fix bugs.

---

## 📄 License
This project is licensed under the MIT License.
