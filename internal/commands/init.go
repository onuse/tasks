package commands

import (
	"fmt"

	"github.com/onuse/tasks/internal/store"
)

func Init() error {
	s := store.New(".")

	if err := s.Init(); err != nil {
		return err
	}

	fmt.Println("Initialized task tracking in .tasks/")
	fmt.Println("Add to git with: git add .tasks && git commit -m \"Initialize task tracking\"")
	return nil
}
