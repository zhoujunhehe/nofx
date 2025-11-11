package kline

import (
	"context"
	"log"
)

// RestReconciler fetches recent windows via REST and reconciles into store.
type RestReconciler interface {
	ReconcileWindow(symbol, interval string, limit int) error
}

type restReconciler struct {
	rate  *Governor
	store Store
	http  *KlineHTTPClient
}

func NewRestReconciler(rate *Governor, store Store) *restReconciler {
	return &restReconciler{rate: rate, store: store, http: NewKlineHTTPClient()}
}

func (r *restReconciler) ReconcileWindow(symbol, interval string, limit int) error {
	if r.store == nil {
		return nil // No error, just skip
	}
	if err := r.rate.Acquire(context.Background(), 1, "reconcile"); err != nil {
		log.Printf("rate limit exceeded for %s %s", symbol, interval)
		return err
	}
	kl, err := r.http.GetKlines(context.Background(), symbol, interval, limit)
	if err != nil {
		log.Printf("reconcile REST error %s %s: %v", symbol, interval, err)
		return err
	}
	r.store.UpsertFinalBatch(symbol, interval, kl)
	return nil
}
