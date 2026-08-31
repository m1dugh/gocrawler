package queue

import (
	"errors"
	"sync"

	"github.com/m1dugh/gocrawler/internal/crawler/scope"
)

// ErrEmpty is returned by Pop when the queue has no elements left.
var ErrEmpty = errors.New("queue: empty")

// ErrOutOfScope is returned by Push when the url does not match the
// queue's scope and was therefore not enqueued.
var ErrOutOfScope = errors.New("queue: url out of scope")

type node struct {
	url  string
	next *node
}

// Queue is a mutex-protected linked-list FIFO queue of urls, bound to a
// Scope that every pushed url must match.
type Queue struct {
	mu     sync.Mutex
	scope  *scope.Scope
	head   *node
	tail   *node
	length int
}

// New returns an empty Queue that validates pushed urls against s.
func New(s *scope.Scope) *Queue {
	return &Queue{scope: s}
}

// Length returns the number of urls currently queued.
func (q *Queue) Length() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.length
}

// Push validates url against the queue's scope and appends it to the tail
// of the queue. It returns ErrOutOfScope without enqueuing if url does not
// match the scope.
func (q *Queue) Push(url string) error {
	if !q.scope.IsInScope(url) {
		return ErrOutOfScope
	}

	n := &node{url: url}

	q.mu.Lock()
	defer q.mu.Unlock()

	if q.tail == nil {
		q.head = n
		q.tail = n
	} else {
		q.tail.next = n
		q.tail = n
	}
	q.length++

	return nil
}

// Pop removes and returns the url at the head of the queue. It returns
// ErrEmpty if the queue has no elements.
func (q *Queue) Pop() (string, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.head == nil {
		return "", ErrEmpty
	}

	n := q.head
	q.head = n.next
	if q.head == nil {
		q.tail = nil
	}
	q.length--

	return n.url, nil
}
