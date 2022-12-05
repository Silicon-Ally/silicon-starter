// Package graphutil provides helpers for working with GraphQL/gqlgen. This is
// heavily inspired by https://gqlgen.com/reference/field-collection/
package graphutil

import (
	"context"
	"time"

	"github.com/99designs/gqlgen/graphql"
)

type Fields struct {
	fields map[string]struct{}
}

func (f *Fields) ContainsAnyOf(targets ...string) bool {
	for _, t := range targets {
		if _, ok := f.fields[t]; ok {
			return true
		}
	}
	return false
}

func ContainsAnyOf(ctx context.Context, targets ...string) bool {
	return GetPreloads(ctx).ContainsAnyOf(targets...)
}

func GetPreloads(ctx context.Context) *Fields {
	if !graphql.HasOperationContext(ctx) {
		return &Fields{fields: make(map[string]struct{})}
	}

	return GetNestedPreloads(
		graphql.GetOperationContext(ctx),
		graphql.CollectFieldsCtx(ctx, nil),
		"",
	)
}

func GetNestedPreloads(ctx *graphql.OperationContext, fields []graphql.CollectedField, prefix string) *Fields {
	f := make(map[string]struct{})
	for _, column := range getNestedPreloads(ctx, fields, prefix) {
		f[column] = struct{}{}
	}
	return &Fields{fields: f}
}

func getNestedPreloads(ctx *graphql.OperationContext, fields []graphql.CollectedField, prefix string) (preloads []string) {
	for _, column := range fields {
		prefixColumn := GetPreloadString(prefix, column.Name)
		preloads = append(preloads, prefixColumn)
		preloads = append(preloads, getNestedPreloads(ctx, graphql.CollectFields(ctx, column.Selections, nil), prefixColumn)...)
	}
	return
}

func GetPreloadString(prefix, name string) string {
	if len(prefix) > 0 {
		return prefix + "." + name
	}
	return name
}

func TimeToPtr(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

func StringToPtr[S ~string](s S) *string {
	if s == "" {
		return nil
	}
	asStr := string(s)
	return &asStr
}
