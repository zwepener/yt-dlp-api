# 🎵 yt-dlp Stream Resolver API
A lightweight Golang API that uses [yt-dlp](https://github.com/yt-dlp/yt-dlp) to resolve a given video URL (YouTube, etc.) into a direct stream URL.

I built this primarily to power two of my personal projects:
* 📱 A WhatsApp bot
* 🎶 A Discord music bot
But it can be used in any project that needs direct access to media streams without dealing with YouTube’s player logic.

# 🚀 Features
* Exposes a simple REST API to resolve media links.
* Uses `yt-dlp` under the hood for wide compatibility.
* Lightweight, containerized with Docker.
* Ready-to-use with `docker-compose`.

# 📦 Getting Started
1. Clone the repository
`git clone https://github.com/zwepener/yt-dlp-api.git
cd yt-dlp-api`
2. Build with Docker
You can build and run the API using Docker:
`docker build -t yt-dlp-api .
docker run -p 8080:8080 yt-dlp-api`
The API will be available at:
👉 `http://localhost:8080`
3. Using docker-compose
If you prefer using `docker-compose`, simply run:
`docker-compose up --build`

# 🔌 API Usage
Endpoint
`POST /resolve`
Request body
`[
  "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
]`
Responses
`{
  "https://www.youtube.com/watch?v=dQw4w9WgXcQ": "https://rr2---sn-woc7kn7y.googlevideo.com/videoplayback?..."
}`

# 🛠 Configuration
Environment variables can be set in `.env` or via Docker (see `.env.template` for all available environment variables).

# 📜 Requirements
* [Go 1.22+](https://go.dev/doc/install) (if building locally without Docker)
* [yt-dlp](https://github.com/yt-dlp/yt-dlp) (if deploying locally without Docker)

# 🤝 Contributing
Contributions are welcome! Feel free to open issues or submit PRs if you’d like to add features or fix bugs.

# 📄 License
This project is licensed under the MIT License.
