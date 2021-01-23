package router_handler

import (
	"context"
	"encoding/json"
	"net/http"

	"cloud.google.com/go/firestore"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"google.golang.org/api/iterator"
)

type Handler struct {
	logger   *zap.SugaredLogger
	router   *mux.Router
	database *firestore.Client
}

func New(logger *zap.SugaredLogger, router *mux.Router, database *firestore.Client) *Handler {
	h := Handler{logger, router, database}
	h.registerRoutes()

	return &h
}

func (h *Handler) registerRoutes() {
	h.router.HandleFunc("/beans", h.getBeans).Methods("GET")
	h.router.HandleFunc("/beans", h.addBean).Methods("POST")
}

type Bean struct {
	Flavors []string `json:"flavors"`
	Name    string   `json:"name"`
	Roaster string   `json:"roaster"`
	Shade   string   `json:"shade"`
}

type BeansResp struct {
	Beans []Bean `json:"beans"`
}

type AddBeanReq struct {
	Flavors []string `json:"flavors"`
	Name    string   `json:"name"`
	Roaster string   `json:"roaster"`
	Shade   string   `json:"shade"`
}

type AddBeanResp struct {
	ID string `json:"id"`
}

func (h *Handler) getBeans(w http.ResponseWriter, r *http.Request) {
	var resp = &BeansResp{}

	iter := h.database.Collection("beans").Documents(context.TODO())
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			h.logger.Fatalf("Failed to iterate: %v", err)
		}

		var b Bean
		doc.DataTo(&b)
		resp.Beans = append(resp.Beans, b)
	}

	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) addBean(w http.ResponseWriter, r *http.Request) {
	var (
		req  AddBeanReq       // The HTTP request
		resp = &AddBeanResp{} // The HTTP response
		ctx  = context.TODO()
		err  error
	)

	// Make sure the JSON is valid
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Make sure roaster exists - we'll talk about this below
	iter := h.database.Collection("roasters").Where("name", "==", req.Roaster).Documents(ctx)
	for {
		doc, err := iter.Next()
		if doc == nil {
			http.Error(w, "invalid roaster", http.StatusBadRequest)
			return
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		break
	}

	// Add the bean
	doc, _, err := h.database.Collection("beans").Add(ctx, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp.ID = doc.ID

	// Return the response as JSON
	json.NewEncoder(w).Encode(resp)
}
