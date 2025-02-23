package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"strconv"
	"sync"

	"github.com/audunhov/innleggsblekka/internal"
	"github.com/audunhov/innleggsblekka/views"
	_ "github.com/mattn/go-sqlite3"
)

func expect(test string, err error) {
	if err != nil {
		log.Fatal("Expect failed:", test, "\nerror:", err.Error())
	}
}

type API struct {
	*internal.Queries
	db *sql.DB
	mu *sync.Mutex
}

func (api *API) Begin(r *http.Request) (*internal.Queries, *sql.Tx) {
	tx, err := api.db.BeginTx(r.Context(), nil)
	expect("can create tx", err)
	api.mu.Lock()

	return api.WithTx(tx), tx
}

func (api *API) Close(r *http.Request, tx *sql.Tx) {
	tx.Rollback()
	api.mu.Unlock()
}

func (api *API) handlerApprovePost(w http.ResponseWriter, r *http.Request) {
	q, tx := api.Begin(r)
	defer api.Close(r, tx)

	user := r.PathValue("user")
	post := r.PathValue("post")

	uid, err := strconv.ParseInt(user, 10, 64)
	expect("valid user id", err)

	pid, err := strconv.ParseInt(post, 10, 64)
	expect("valid post id", err)

	if uid == pid {
		http.Error(w, "Cannot approve own post", http.StatusBadRequest)
		return
	}

	p, err := q.ApprovePost(r.Context(), internal.ApprovePostParams{
		ID:         pid,
		Approvedby: sql.NullInt64{Int64: uid, Valid: true},
	})

	j, err := json.Marshal(p)
	expect("Convert json", err)

	_, err = q.CreateLogEntry(r.Context(), internal.CreateLogEntryParams{
		Tablename: "posts",
		User:      uid,
		Newvalue:  j,
	})

	expect("Can create log entry", err)

	tx.Commit()

	respondWithJson(w, http.StatusOK, p)
}
func (api *API) handlerRemoveApproval(w http.ResponseWriter, r *http.Request) {
	a, tx := api.Begin(r)
	defer api.Close(r, tx)

	post := r.PathValue("post")

	pid, err := strconv.ParseInt(post, 10, 64)
	expect("valid post id", err)

	p, err := a.RemoveApprovalFromPost(r.Context(), pid)
	expect("Convert json", err)

	tx.Commit()

	respondWithJson(w, http.StatusOK, p)
}

func (api *API) handlerGetFavourites(w http.ResponseWriter, r *http.Request) {
	posts, err := api.ListMostFavouritedPosts(r.Context())
	expect("Can get stats", err)
	respondWithJson(w, http.StatusOK, posts)
}

func respondWithJson(w http.ResponseWriter, status int, data any) {
	j, err := json.Marshal(data)
	expect("Can convert json", err)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(j)
}

func getUidFromReq(r *http.Request) (int64, error) {
	// TODO: Implement this
	fmt.Println("Mocking id for request to", r.URL.Path)
	return 1, nil
}

func main() {
	port := flag.String("port", ":5050", "port to serve")
	file := flag.String("file", "./innlegg.db", "db file path")
	flag.Parse()

	slog.Debug("Connecting to DB")
	conn, err := sql.Open("sqlite3", *file)

	if err != nil {
		log.Fatal("Could not connect to db")
	}

	slog.Debug("Successfully connected to DB")
	defer conn.Close()

	api := API{
		db:      conn,
		Queries: internal.New(conn),
		mu:      &sync.Mutex{},
	}

	mux := http.NewServeMux()
	v1 := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		twpc, err := api.ListTagsWithPostCount(r.Context())
		expect("Can get tags and counts", err)
		fmt.Println(twpc)

		tags, err := api.ListTags(r.Context())
		expect("Can get tags and counts", err)

		expect("Can get tags and counts", err)
		views.HomePage(twpc, tags).Render(r.Context(), w)
	})

	mux.HandleFunc("/tag/{id}/", func(w http.ResponseWriter, r *http.Request) {

		tid, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		expect("Can convert param", err)

		tag, err := api.GetTagById(r.Context(), tid)
		expect("Tag found", err)

		posts, err := api.ListPostsByTag(r.Context(), tid)
		expect("Posts found", err)

		views.TagPage(tag, posts).Render(r.Context(), w)

	})

	mux.HandleFunc("GET /posts/{id}/", func(w http.ResponseWriter, r *http.Request) {

		pid, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		expect("Can convert param", err)

		post, err := api.GetPostById(r.Context(), pid)
		expect("Can convert post", err)

		views.PostPage(post).Render(r.Context(), w)
	})

	v1.HandleFunc("GET /users/", func(w http.ResponseWriter, r *http.Request) {
		users, err := api.ListUsers(r.Context())
		expect("Can find users", err)
		respondWithJson(w, http.StatusOK, users)
	})

	v1.HandleFunc("POST /search/", func(w http.ResponseWriter, r *http.Request) {
		if searchval := r.FormValue("search"); searchval != "" {
			posts, err := api.SearchPosts(r.Context(), r.FormValue("search"))
			expect("Can search post", err)

			for _, post := range posts {
				views.SearchResult(post.ID, post.Title, post.Highlight).Render(r.Context(), w)
			}
			return
		}
		respondWithJson(w, 200, "Ingen innlegg")

	})

	v1.HandleFunc("GET /users/{id}/", func(w http.ResponseWriter, r *http.Request) {
		uid, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		expect("Can get id", err)

		u, err := api.GetUserById(r.Context(), uid)
		expect("Can find user", err)
		respondWithJson(w, http.StatusOK, u)
	})

	v1.HandleFunc("GET /posts/", func(w http.ResponseWriter, r *http.Request) {
		posts, err := api.ListPosts(r.Context())
		expect("Can find posts", err)
		respondWithJson(w, http.StatusOK, posts)
	})
	v1.HandleFunc("POST /posts/", func(w http.ResponseWriter, r *http.Request) {
		q, tx := api.Begin(r)
		defer api.Close(r, tx)
		defer tx.Commit()
		err := r.ParseForm()
		expect("Can parse form", err)

		title := r.FormValue("title")
		body := r.FormValue("body")
		tags := r.Form["tags"]

		post, err := q.CreatePost(r.Context(), internal.CreatePostParams{
			Title:     title,
			Body:      body,
			Creatorid: 1,
		})
		for _, tag := range tags {
			tid, err := strconv.ParseInt(tag, 10, 64)

			expect("can convert tid", err)

			tp, err := q.AddTagToPost(r.Context(), internal.AddTagToPostParams{
				Postid: post.ID,
				Tagid:  tid,
			})

			fmt.Println("Added tp", tp)

			expect("Can add tag to post", err)

		}

		respondWithJson(w, http.StatusOK, post)
	})
	v1.HandleFunc("GET /posts/{id}/favourite/", func(w http.ResponseWriter, r *http.Request) {
		pid, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		expect("Can get id", err)

		uid, err := getUidFromReq(r)
		expect("Can get user from req", err)

		posts, err := api.AddFavourite(r.Context(), internal.AddFavouriteParams{
			Postid: pid,
			Userid: uid,
		})
		expect("Can find posts", err)
		respondWithJson(w, http.StatusOK, posts)
	})

	v1.HandleFunc("GET /posts/{id}/", func(w http.ResponseWriter, r *http.Request) {
		pid, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		expect("Can get id", err)
		post, err := api.GetPostById(r.Context(), pid)
		if err != nil {
			http.Error(w, "No post found", http.StatusNotFound)
			return
		}
		respondWithJson(w, http.StatusOK, post)
	})

	// v1.HandleFunc("GET /posts/tag/{tag}/", func(w http.ResponseWriter, r *http.Request) {
	// 	pid, err := strconv.ParseInt(r.PathValue("tag"), 10, 64)
	// 	expect("Can get id", err)
	// 	posts, err := api.ListPostsByTag(r.Context(), pid)
	// 	expect("Can find post", err)
	// 	respondWithJson(w, http.StatusOK, posts)
	// })

	v1.HandleFunc("GET /tags/", func(w http.ResponseWriter, r *http.Request) {
		tags, err := api.ListTags(r.Context())
		expect("Can find tags", err)
		respondWithJson(w, http.StatusOK, tags)
	})
	v1.HandleFunc("POST /tags/", func(w http.ResponseWriter, r *http.Request) {
		tag, err := api.CreateTag(r.Context(), r.FormValue("name"))
		expect("Can find tags", err)
		views.AddPostTagOption(tag).Render(r.Context(), w)
	})

	mux.Handle("/v1/", http.StripPrefix("/v1", v1))

	server := http.Server{
		Addr:    *port,
		Handler: mux,
	}

	slog.Info("Started server", "port", *port)

	if err := server.ListenAndServe(); err != nil {
		log.Fatal("Server shut down:", err)
	}
}
