package app

import (
	"errors"
	"sync"
)

var ErrTodoNotFound = errors.New("todo not found")

type TodoStore struct {
	mu     sync.RWMutex
	todos  []Todo
	nextID int
}

func NewTodoStore() *TodoStore {
	return &TodoStore{
		todos: []Todo{
			{ID: 1, Title: "Run todo API with Docker", Completed: false},
			{ID: 2, Title: "Open Kibana and inspect logs", Completed: false},
		},
		nextID: 3,
	}
}

func (s *TodoStore) List() []Todo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	todos := make([]Todo, len(s.todos))
	copy(todos, s.todos)
	return todos
}

func (s *TodoStore) Create(input TodoCreate) Todo {
	s.mu.Lock()
	defer s.mu.Unlock()

	todo := Todo{
		ID:        s.nextID,
		Title:     input.Title,
		Completed: false,
	}
	s.nextID++
	s.todos = append(s.todos, todo)
	return todo
}

func (s *TodoStore) Update(id int, input TodoUpdate) (Todo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for index := range s.todos {
		if s.todos[index].ID == id {
			if input.Title != nil {
				s.todos[index].Title = *input.Title
			}
			if input.Completed != nil {
				s.todos[index].Completed = *input.Completed
			}
			return s.todos[index], nil
		}
	}

	return Todo{}, ErrTodoNotFound
}

func (s *TodoStore) Delete(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for index := range s.todos {
		if s.todos[index].ID == id {
			s.todos = append(s.todos[:index], s.todos[index+1:]...)
			return nil
		}
	}

	return ErrTodoNotFound
}
