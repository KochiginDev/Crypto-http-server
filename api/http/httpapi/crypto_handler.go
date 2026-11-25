package httpapi

import (
	"crypthttp/domain"
	"crypthttp/usecases"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type CryptoHandler struct {
	crypto usecases.Crypto
}

func NewCryptoHandler(crypto usecases.Crypto) *CryptoHandler {
	return &CryptoHandler{crypto: crypto}
}

func (h *CryptoHandler) Add(w http.ResponseWriter, r *http.Request) {
	var symbol AddCryptoRequest
	err := decodeJSON(r, w, &symbol)
	if err != nil {
		ErrorWriter(w, r, err)
		return
	}
	coin, err := h.crypto.Add(r.Context(), symbol.Symbol)
	if err != nil {
		ErrorWriter(w, r, err)
		return
	}
	_ = WriteJSON(w, http.StatusCreated, CryptoResponse{Crypto: coin})
	return
}

func (h *CryptoHandler) Get(w http.ResponseWriter, r *http.Request) {
	symbol := chi.URLParam(r, "symbol")
	coin, err := h.crypto.Get(r.Context(), symbol)
	if err != nil {
		ErrorWriter(w, r, err)
		return
	}
	_ = WriteJSON(w, 200, coin)
	return
}
func (h *CryptoHandler) GetList(w http.ResponseWriter, r *http.Request) {
	list, err := h.crypto.GetList(r.Context())
	if err != nil {
		ErrorWriter(w, r, err)
		return
	}
	if list == nil {
		list = []domain.Crypto{}
	}
	_ = WriteJSON(w, 200, CryptoListResponse{Cryptos: list})
	return
}

func (h *CryptoHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	symbol := chi.URLParam(r, "symbol")
	coin, err := h.crypto.Refresh(r.Context(), symbol)
	if err != nil {
		ErrorWriter(w, r, err)
		return
	}
	_ = WriteJSON(w, 200, CryptoResponse{Crypto: coin})
	return
}
func (h *CryptoHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	symbol := chi.URLParam(r, "symbol")
	list, err := h.crypto.GetHistory(r.Context(), symbol)
	if err != nil {
		ErrorWriter(w, r, err)
		return
	}
	if list == nil {
		list = []domain.PricePoint{}
	}
	_ = WriteJSON(w, 200, HistoryResponse{Symbol: symbol, History: list})
	return
}

func (h *CryptoHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	symbol := chi.URLParam(r, "symbol")
	coin, err := h.crypto.Get(r.Context(), symbol)
	if err != nil {
		ErrorWriter(w, r, err)
		return
	}
	var stat StatsResponse
	statistic, err := h.crypto.GetStats(r.Context(), symbol)
	if err != nil {
		ErrorWriter(w, r, err)
		return
	}
	stat.Stats = statistic
	stat.Price = coin.CurrentPrice
	stat.Symbol = coin.Symbol
	_ = WriteJSON(w, 200, stat)
	return
}

func (h *CryptoHandler) Delete(w http.ResponseWriter, r *http.Request) {
	symbol := chi.URLParam(r, "symbol")
	err := h.crypto.Delete(r.Context(), symbol)
	if err != nil {
		ErrorWriter(w, r, err)
		return
	}
	_ = WriteJSON(w, 200, struct{}{})
	return
}
