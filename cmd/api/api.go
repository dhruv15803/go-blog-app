package main

import (
	"github.com/dhruv15803/go-blog-app/internal/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
	"time"
)

type server struct {
	addr                string
	readRequestTimeout  time.Duration
	writeRequestTimeout time.Duration
	handler             *handlers.Handler
}

func newServer(addr string, readRequestTimeout time.Duration, writeRequestTimeout time.Duration, handler *handlers.Handler) *server {
	return &server{
		addr:                addr,
		readRequestTimeout:  readRequestTimeout,
		writeRequestTimeout: writeRequestTimeout,
		handler:             handler,
	}
}

func (s *server) mount() *chi.Mux {

	r := chi.NewRouter()

	r.Route("/api", func(r chi.Router) {
		r.Use(middleware.Logger)
		r.Get("/health", s.handler.HealthCheckHandler)

		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", s.handler.RegisterUserHandler)
			r.Put("/activate/{token}", s.handler.ActivateUserHandler)
			r.Post("/login", s.handler.LoginUserHandler)
			r.With(s.handler.AuthMiddleware).Get("/user", s.handler.GetUserHandler)
		})

		r.Route("/blog", func(r chi.Router) {
			r.Group(func(r chi.Router) {
				r.Use(s.handler.AuthMiddleware)
				r.Post("/", s.handler.CreateBlogHandler)
				r.Delete("/{blogId}", s.handler.DeleteBlogHandler)
				r.Patch("/{blogId}/status", s.handler.UpdateBlogStatusHandler)
				r.Post("/{blogId}/like", s.handler.LikeBlogHandler)
				r.Post("/{blogId}/bookmark", s.handler.BookmarkBlogHandler)
			})
			//	get blog posts feed for a topic handler - unauthenticated
			r.Get("/{topicId}/blogs", s.handler.GetBlogsFeedByTopicHandler)
			r.With(s.handler.OptionalAuthMiddleware).Get("/blogs/feed", s.handler.GetBlogsFeedHandler)
		})

		r.Route("/blog-comment", func(r chi.Router) {
			r.Group(func(r chi.Router) {
				r.Use(s.handler.AuthMiddleware)
				r.Post("/", s.handler.CreateBlogCommentHandler)
				r.Delete("/{blogCommentId}", s.handler.DeleteBlogCommentHandler)
				r.Put("/{blogCommentId}", s.handler.UpdateBlogCommentHandler)
				r.Post("/{blogCommentId}/like", s.handler.LikeBlogCommentHandler)
			})

			r.Get("/{blogId}/blog-comments", s.handler.GetBlogCommentsHandler)
			r.Get("/{blogCommentId}/comments", s.handler.GetBlogCommentCommentsHandler)
		})

		r.Route("/topic", func(r chi.Router) {

			r.Get("/topics", s.handler.GetTopicsHandler)

			r.Group(func(r chi.Router) {
				//	add , delete and edit blog topics (admin routes)
				r.Use(s.handler.AuthMiddleware)
				r.Use(s.handler.AdminAuthMiddleware)
				r.Post("/", s.handler.CreateTopicHandler)
				r.Put("/{topicId}", s.handler.UpdateTopicHandler)
				r.Delete("/{topicId}", s.handler.DeleteTopicHandler)
			})

			r.Group(func(r chi.Router) {
				r.Use(s.handler.AuthMiddleware)
				r.Post("/{topicId}/follow", s.handler.FollowTopicHandler)
			})
		})
	})

	return r
}

func (s *server) run() error {

	// attach routes to handler

	r := s.mount()

	server := http.Server{
		Addr:         ":" + s.addr,
		Handler:      r,
		ReadTimeout:  s.readRequestTimeout,
		WriteTimeout: s.writeRequestTimeout,
	}

	return server.ListenAndServe()
}
