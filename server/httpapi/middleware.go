// Package httpapi owns HTTP mechanics shared by authenticated and public APIs.
package httpapi

import (
	"container/list"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"net"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
)

var ErrInvalidMiddleware = errors.New("invalid HTTP middleware configuration")

type bucket struct {
	tokens   float64
	last     time.Time
	lastSeen time.Time
	element  *list.Element
}

// TokenBuckets is a bounded, concurrency-safe LRU token-bucket collection.
// A regressed clock never refills or moves a bucket forward.
type TokenBuckets struct {
	mu       sync.Mutex
	buckets  map[string]*bucket
	recency  *list.List
	capacity float64
	perMS    float64
	idleTTL  time.Duration
	max      int
}

func NewTokenBuckets(capacity, perMinute, maxEntries int) (*TokenBuckets, error) {
	if capacity < 1 || perMinute < 1 || maxEntries < 1 {
		return nil, ErrInvalidMiddleware
	}
	refillMinutes := (capacity + perMinute - 1) / perMinute
	if refillMinutes < 1 {
		refillMinutes = 1
	}
	return &TokenBuckets{
		buckets:  make(map[string]*bucket),
		recency:  list.New(),
		capacity: float64(capacity),
		perMS:    float64(perMinute) / float64(time.Minute/time.Millisecond),
		idleTTL:  time.Duration(refillMinutes) * time.Minute,
		max:      maxEntries,
	}, nil
}

func (buckets *TokenBuckets) Allow(key string, now time.Time) bool {
	if buckets == nil || key == "" || now.IsZero() {
		return false
	}
	buckets.mu.Lock()
	defer buckets.mu.Unlock()
	buckets.evictExpired(now)
	current := buckets.buckets[key]
	if current == nil {
		if len(buckets.buckets) >= buckets.max {
			buckets.removeOldest()
		}
		current = &bucket{tokens: buckets.capacity, last: now, lastSeen: now}
		current.element = buckets.recency.PushFront(key)
		buckets.buckets[key] = current
	} else if !now.Before(current.lastSeen) {
		current.lastSeen = now
		buckets.recency.MoveToFront(current.element)
	}
	elapsed := now.Sub(current.last).Milliseconds()
	if elapsed > 0 {
		current.tokens += float64(elapsed) * buckets.perMS
		if current.tokens > buckets.capacity {
			current.tokens = buckets.capacity
		}
		current.last = now
	}
	if current.tokens < 1 {
		return false
	}
	current.tokens--
	return true
}

func (buckets *TokenBuckets) Len() int {
	if buckets == nil {
		return 0
	}
	buckets.mu.Lock()
	defer buckets.mu.Unlock()
	return len(buckets.buckets)
}

func (buckets *TokenBuckets) Contains(key string) bool {
	if buckets == nil {
		return false
	}
	buckets.mu.Lock()
	defer buckets.mu.Unlock()
	return buckets.buckets[key] != nil
}

func (buckets *TokenBuckets) evictExpired(now time.Time) {
	for element := buckets.recency.Back(); element != nil; element = buckets.recency.Back() {
		key := element.Value.(string)
		current := buckets.buckets[key]
		if now.Before(current.lastSeen) || now.Sub(current.lastSeen) < buckets.idleTTL {
			return
		}
		buckets.recency.Remove(element)
		delete(buckets.buckets, key)
	}
}

func (buckets *TokenBuckets) removeOldest() {
	element := buckets.recency.Back()
	if element == nil {
		return
	}
	delete(buckets.buckets, element.Value.(string))
	buckets.recency.Remove(element)
}

func ClientIP(request *http.Request, trustedProxyHops int) string {
	if request == nil || trustedProxyHops < 0 {
		return ""
	}
	host, _, err := net.SplitHostPort(request.RemoteAddr)
	if err != nil {
		host = request.RemoteAddr
	}
	if parsed := net.ParseIP(strings.TrimSpace(host)); parsed != nil {
		host = parsed.String()
	}
	if trustedProxyHops == 0 {
		return host
	}
	var forwarded []string
	for _, value := range request.Header.Values("X-Forwarded-For") {
		for _, entry := range strings.Split(value, ",") {
			forwarded = append(forwarded, strings.TrimSpace(entry))
		}
	}
	if len(forwarded) < trustedProxyHops {
		return host
	}
	candidate := forwarded[len(forwarded)-trustedProxyHops]
	if parsed := net.ParseIP(candidate); parsed != nil {
		return parsed.String()
	}
	return host
}

type RequestIDs struct {
	pattern  *regexp.Regexp
	maxBytes int
	random   io.Reader
}

func NewRequestIDs(pattern string, maxBytes int, random io.Reader) (*RequestIDs, error) {
	compiled, err := regexp.Compile(pattern)
	if err != nil || pattern == "" || maxBytes < 1 || random == nil {
		return nil, ErrInvalidMiddleware
	}
	return &RequestIDs{pattern: compiled, maxBytes: maxBytes, random: random}, nil
}

func Phase0RequestIDs(pattern string, maxBytes int) (*RequestIDs, error) {
	return NewRequestIDs(pattern, maxBytes, rand.Reader)
}

func (ids *RequestIDs) Resolve(candidate string, now time.Time) (string, error) {
	if ids == nil || now.IsZero() {
		return "", ErrInvalidMiddleware
	}
	if len(candidate) <= ids.maxBytes && ids.pattern.MatchString(candidate) {
		return candidate, nil
	}
	return newUUIDv7(now, ids.random)
}

func newUUIDv7(now time.Time, random io.Reader) (string, error) {
	if now.IsZero() || random == nil || now.UnixMilli() < 0 {
		return "", ErrInvalidMiddleware
	}
	var raw [16]byte
	if _, err := io.ReadFull(random, raw[:]); err != nil {
		return "", err
	}
	milliseconds := uint64(now.UnixMilli())
	for index := 5; index >= 0; index-- {
		raw[index] = byte(milliseconds)
		milliseconds >>= 8
	}
	raw[6] = raw[6]&0x0f | 0x70
	raw[8] = raw[8]&0x3f | 0x80
	var encoded [36]byte
	hex.Encode(encoded[0:8], raw[0:4])
	encoded[8] = '-'
	hex.Encode(encoded[9:13], raw[4:6])
	encoded[13] = '-'
	hex.Encode(encoded[14:18], raw[6:8])
	encoded[18] = '-'
	hex.Encode(encoded[19:23], raw[8:10])
	encoded[23] = '-'
	hex.Encode(encoded[24:36], raw[10:16])
	return string(encoded[:]), nil
}
