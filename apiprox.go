package apiprox

import (
	"io"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

// Based of gist: https://gist.github.com/yowu/f7dc34bd4736a65ff28d
var hopHeaders = []string{
	"Connection",
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Te", // canonical version of "TE"
	"Trailers",
	"Transfer-Encoding",
	"Upgrade",
}

type Server struct {
	Router *chi.Mux
}

func New() *Server {
	return &Server{
		Router: newRouter(),
	}
}

func newRouter() *chi.Mux {
	r := chi.NewRouter()

	// A good base middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(cors.AllowAll().Handler)

	r.HandleFunc("/*", apiProxy())

	return r
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func delHopHeaders(header http.Header) {
	for _, h := range hopHeaders {
		header.Del(h)
	}
}

func apiProxy() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fwdURL := r.URL.Query().Get("fwd")
		req, err := http.NewRequestWithContext(r.Context(), r.Method, fwdURL, r.Body)
		if err != nil {
			log.Printf("new request: %v", err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Printf("http do: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		log.Println(req.RemoteAddr, " ", resp.Status)

		delHopHeaders(resp.Header)
		copyHeader(w.Header(), resp.Header)
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	}
}
