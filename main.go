package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/storage"
)

func main() {
	// Parse command-line flags
	port := flag.String("port", "3000", "Port to run the server on")
	dir := flag.String("dir", "./downloads", "Directory to store downloaded files")
	flag.Parse()

	// Configure the torrent client
	clientConfig := torrent.NewDefaultClientConfig()
	clientConfig.DataDir = *dir
	clientConfig.DefaultStorage = storage.NewFileByInfoHash(clientConfig.DataDir)

	// Create the torrent client
	client, err := torrent.NewClient(clientConfig)
	if err != nil {
		log.Fatalf("Error creating torrent client: %v", err)
	}
	defer client.Close()

	// Handle the /stream endpoint
	http.HandleFunc("/stream", func(w http.ResponseWriter, r *http.Request) {
		// Extract the magnet link from the query parameters
		magnetLink := r.URL.Query().Get("magnet")
		if magnetLink == "" {
			http.Error(w, "Magnet link is required", http.StatusBadRequest)
			return
		}

		log.Printf("Processing magnet link: %s", magnetLink)

		// Add the magnet link to the torrent client
		t, err := client.AddMagnet(magnetLink)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error adding magnet: %v", err), http.StatusInternalServerError)
			log.Printf("Error adding magnet: %v", err)
			return
		}
		<-t.GotInfo()

		// Find the largest video file in the torrent
		videoFile := findLargestVideoFile(t.Files())
		if videoFile == nil {
			http.Error(w, "No video file found in the torrent", http.StatusNotFound)
			log.Println("No video file found in the torrent")
			return
		}

		log.Printf("Streaming video file: %s", videoFile.Path())

		// Determine the content type of the video file
		contentType := getContentType(videoFile.Path())
		// Handle range requests and stream the video file
		handleRangeRequest(w, r, videoFile, contentType)
	})

	log.Printf("Server started at http://localhost:%s", *port)

	// Periodically clear the downloads folder
	go func() {
		for {
			time.Sleep(3 * time.Hour)
			clearDownloadsFolder(*dir)
		}
	}()

	// Start the HTTP server
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}

// findLargestVideoFile finds the largest video file in the list of torrent files.
func findLargestVideoFile(files []*torrent.File) *torrent.File {
	var videoFile *torrent.File
	for _, file := range files {
		if isVideoFile(file.Path()) && (videoFile == nil || file.Length() > videoFile.Length()) {
			videoFile = file
		}
	}
	return videoFile
}

// isVideoFile checks if the file has a video extension (mp4 or mkv).
func isVideoFile(path string) bool {
	ext := filepath.Ext(path)
	return ext == ".mp4" || ext == ".mkv"
}

// getContentType returns the content type based on the file extension.
func getContentType(path string) string {
	switch filepath.Ext(path) {
	case ".mkv":
		return "video/x-matroska"
	default:
		return "video/mp4"
	}
}

// handleRangeRequest handles HTTP range requests and streams the video file.
func handleRangeRequest(w http.ResponseWriter, r *http.Request, videoFile *torrent.File, contentType string) {
	rangeHeader := r.Header.Get("Range")
	start, end, contentLength := parseRangeHeader(rangeHeader, videoFile.Length())

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", contentLength))
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, videoFile.Length()))
	w.WriteHeader(http.StatusPartialContent)

	streamVideoFile(w, videoFile, start, end)
}

// parseRangeHeader parses the HTTP Range header and returns the start, end, and content length.
func parseRangeHeader(rangeHeader string, fileLength int64) (int64, int64, int64) {
	var start, end int64
	if rangeHeader != "" {
		fmt.Sscanf(rangeHeader, "bytes=%d-%d", &start, &end)
		if end == 0 {
			end = fileLength - 1
		}
	} else {
		start = 0
		end = fileLength - 1
	}
	return start, end, end - start + 1
}

// streamVideoFile streams the video file to the client.
func streamVideoFile(w http.ResponseWriter, videoFile *torrent.File, start, end int64) {
	reader := videoFile.NewReader()
	defer reader.Close()

	_, err := reader.Seek(start, io.SeekStart)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error seeking in file: %v", err), http.StatusInternalServerError)
		log.Printf("Error seeking in file: %v", err)
		return
	}

	buf := make([]byte, 1024*1024) // 1MB buffer
	for pos := start; pos <= end; {
		n, err := reader.Read(buf)
		if err != nil && err != io.EOF {
			http.Error(w, fmt.Sprintf("Error reading file: %v", err), http.StatusInternalServerError)
			log.Printf("Error reading file: %v", err)
			return
		}
		if n > 0 {
			_, err := w.Write(buf[:n])
			if err != nil {
				http.Error(w, fmt.Sprintf("Error writing to client: %v", err), http.StatusInternalServerError)
				log.Printf("Error writing to client: %v", err)
				return
			}
			w.(http.Flusher).Flush()
		}
		if err == io.EOF || n == 0 {
			break
		}
		pos += int64(n)
	}

	log.Println("Stream completed")
}

// clearDownloadsFolder clears the downloads folder.
func clearDownloadsFolder(dir string) {
	err := os.RemoveAll(dir)
	if err != nil {
		log.Printf("Error clearing downloads folder: %v", err)
	} else {
		log.Println("Downloads folder cleared")
	}
}
