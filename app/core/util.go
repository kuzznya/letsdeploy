package core

import "database/sql"

func toMap[T any, K comparable, V any](items []T, keyExtractor func(T) K, valueExtractor func(T) V) map[K]V {
	m := make(map[K]V)
	for _, item := range items {
		m[keyExtractor(item)] = valueExtractor(item)
	}
	return m
}

func toMapSelf[T any, K comparable](items []T, keyExtractor func(T) K) map[K]T {
	return toMap(items, keyExtractor, func(t T) T { return t })
}

func contains[K comparable, V any](m map[K]V, key K) bool {
	_, found := m[key]
	return found
}

func remove[T any](items []T, idx int) []T {
	return append(items[:idx], items[idx+1:]...)
}

func removeFirstByPredicate[T any](items []T, predicate func(T) bool) []T {
	for i, item := range items {
		if predicate(item) {
			return remove(items, i)
		}
	}
	return items
}

func mapItems[T any, R any](items []T, mapper func(T) R) []R {
	r := make([]R, len(items))
	for i, item := range items {
		r[i] = mapper(item)
	}
	return r
}

func filter[T any](items []T, predicate func(T) bool) []T {
	var r []T
	for _, item := range items {
		if predicate(item) {
			r = append(r, item)
		}
	}
	return r
}

func fromNullString(s sql.NullString) *string {
	if s.Valid {
		return &s.String
	} else {
		return nil
	}
}

func toNullString(s *string) sql.NullString {
	if s != nil {
		return sql.NullString{String: *s, Valid: true}
	} else {
		return sql.NullString{Valid: false}
	}
}
