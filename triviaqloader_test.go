package main

import (
	"fmt"
	"os"
	"testing"
)

func TestLoadFile(t *testing.T) {
	example, err := os.ReadFile("./assets/trivia/example.trivia")
	if err != nil {
		t.Fatalf(err.Error())
	}
	//fmt.Println(string(example))
	bank, err := loadTriviaBankFromString(string(example))
	if err != nil {
		t.Fatalf(err.Error())
	}
	fmt.Println("Question bank: ", *bank)

	if bank.Questions[0].Q != "What is the capital city of Germany?" {
		t.Fatalf("Not loaded properly 1")
	}

	if bank.Questions[1].Q != "Who was the first US President?" {
		t.Fatalf("Not loaded properly 2")
	}

	if len(bank.Questions[0].A) != 2 {
		t.Fatalf(("Not loaded properly 3"))
	}
}
