package result

import (
	"time"

	"github.com/google/uuid"
)

type GetChats struct {
	ID        uuid.UUID
	Title     string
	Couple    []Couple
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Couple struct {
	Command Command
	Result  Result
}

type Model struct {
	ID   uuid.UUID
	Name string
}

type Tool struct {
	ID   uuid.UUID
	Name string
}

type Command struct {
	ID        uuid.UUID
	Text      string
	Settings  map[string]string
	Flags     []string
	Model     *Model
	Tool      *Tool
	Status    string
	Media     []string
	CreatedAt time.Time
}

type Result struct {
	ID        uuid.UUID
	Text      string
	Model     Model
	Tool      Tool
	Media     []string
	CreatedAt time.Time
}
