The purpose of this application is to provide a simple and efficient way to stream video files directly from torrent magnet links over HTTP. Here are the key objectives and functionalities of the application:

Key Objectives
Stream Video Files from Torrents:

The application allows users to stream video files (like .mp4 and .mkv) directly from torrent magnet links without needing to download the entire file first.

HTTP-Based Streaming:

It serves the video files over HTTP, making it compatible with standard web browsers and media players that support HTTP streaming.

Support for Range Requests:

The application supports HTTP range requests, enabling clients to seek within the video file, which is essential for smooth playback in media players.

Minimal Storage Usage:

By streaming the video file directly, the application minimizes the storage usage on the server, as it only stores the torrent data needed for streaming.

Periodic Cleanup:

The application periodically clears the downloads folder to free up space, ensuring that the server does not run out of storage due to accumulated torrent data.

Functionalities
Command-Line Interface:

Allows users to specify the port and directory for the server and downloads via command-line flags.

Torrent Client Integration:

Integrates with a torrent client to handle magnet links and manage torrent data.

HTTP Server:

Provides an HTTP endpoint (/stream) to accept requests with magnet links and stream the corresponding video files.

Magnet Link Processing:

Extracts and validates magnet links from HTTP requests.

Adds the magnet link to the torrent client and waits for the torrent metadata.

Video File Identification:

Identifies the largest video file in the torrent, which is typically the main video content.

Content Type Detection:

Determines the content type of the video file based on its extension (e.g., video/mp4 or video/x-matroska).

Range Request Handling:

Parses and handles HTTP range requests to support seeking within the video file.

Video Streaming:

Streams the video file to the client, handling reading from the torrent file and writing to the HTTP response concurrently.

Error Handling:

Provides robust error handling for various stages of the process, including adding magnet links, finding video files, and streaming the video.

Periodic Task:

Periodically clears the downloads folder to manage storage usage.

Logging:

Logs important events and errors to help with debugging and monitoring.

Summary
The primary purpose of this application is to enable users to stream video files directly from torrent magnet links over HTTP, providing a seamless and efficient way to watch videos without the need for full downloads. It leverages Go's concurrency model to handle multiple requests and tasks efficiently, ensuring a smooth streaming experience.


how to run
go run main.go -port 8080 -dir /tmp/downloads

ubuntu 22.04/debian 12 64bit
to run binary
./rsd -port 8080 -dir /tmp/downloads


build static binary for linx, you can also build for Windows Mac etc
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o rsd .

