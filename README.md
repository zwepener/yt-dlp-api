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

## 📜 Requirements

> [!IMPORTANT]
> Although it is not strictly required to have a redis service, it is highly recommended as the API uses it to cache resolved urls to reduce calls to the internet (YouTube, Instagram, Facebook, etc.).

### When deploying using Docker (or Docker Compose)

* [Docker Engine](https://docs.docker.com/engine/install/) installed on host machine
* [Docker Compose](https://docs.docker.com/compose/install/linux/) plugin for docker engine (only if you plan on using docker compose)

### When deploying directly to host machine
* [Go 1.22+](https://go.dev/doc/install) installed on host machine
* [yt-dlp](https://github.com/yt-dlp/yt-dlp) installed on host machine

---

## 📦 Getting Started

### 1. Clone the repository

```bash
git clone https://github.com/zwepener/yt-dlp-api.git
cd yt-dlp-api
```

### 2. Configure

Create a copy of the included [`.env.template`](.env.template) file:
```bash
cp .env.template .env
```
You can edit your `.env` file according to your needs.

### 3. Deploy

> [!IMPORTANT]
> If you don't want to use docker compose, you will have to either host your own redis service or manually start and configure a redis container seperately.
> If you have your own redis service and you want to use docker compose, you will need to modify the existing [`docker-compose.yml`](docker-compose.yml) file to remove my redis service or use your own `docker-compose.yml` file entirely.

#### Using Docker Compose (Recommended)

```bash
docker compose up --detach --build
docker compose logs --follow api
```
This will also start a `redis` container. The api uses this service to cache resolved urls for a set period.

#### Using Docker

```bash
docker build --tag yt-dlp-api .
docker run --detach --publish 8080:8080 yt-dlp-api
docker logs --follow yt-dlp-api
```

#### Directly on Host Machine

```bash
go build -o ./api ./src
chmod +x ./api
./api
```

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

## 🔮 Future Plans

Some ideas I’d like to work on in the future:
- [x] Hash urls for redis cache keys.
- [ ] Adopt [gin](https://gin-gonic.com/) for easier maintentance.
- [ ] Implement proper logging.
- [ ] Add health checks and better error handling.

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
