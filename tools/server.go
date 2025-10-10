package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"math"
	"math/big"
	"net/http"
	"strconv"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// randomString generates a random string of exactly n bytes
func randomString(n int) string {
	if n <= 0 {
		return ""
	}

	b := make([]byte, n)
	for i := range b {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			panic(err)
		}
		b[i] = letters[num.Int64()]
	}
	return string(b)
}

// MiBToBytes converts MiB (can be fractional) to bytes
func MiBToBytes(mib float64) int {
	return int(math.Round(mib * 1024 * 1024))
}

func main() {
	port := flag.String("port", "8080", "Port to listen on")
	urlPath := flag.String("url", "/", "URL path to handle")
	defaultSizeMiB := flag.Float64("body", 1, "Default size of the response body (in MiB, can be fractional)")

	flag.Parse()

	http.HandleFunc(*urlPath, func(w http.ResponseWriter, r *http.Request) {
		size := MiBToBytes(*defaultSizeMiB)

		// Parse ?size= query parameter (also in MiB)
		if sizeParam := r.URL.Query().Get("size"); sizeParam != "" {
			if parsed, err := strconv.ParseFloat(sizeParam, 64); err == nil && parsed > 0 {
				size = MiBToBytes(parsed)
			}
		}

		// Safety limit: max 10 MiB per response
		if size > MiBToBytes(10) {
			size = MiBToBytes(10)
		}

		body := randomString(size)

		fmt.Printf("Received %s %s (size=%d bytes)\n", r.Method, r.URL.Path, size)
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte(body))
	})

	addr := fmt.Sprintf(":%s", *port)
	fmt.Printf("Serving random responses on http://localhost%s%s (default size: %.2f MiB)\n",
		addr, *urlPath, *defaultSizeMiB)

	if err := http.ListenAndServe(addr, nil); err != nil {
		panic(err)
	}
}
