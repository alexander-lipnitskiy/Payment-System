package main

import (
	"errors"
	"fmt"
	"sync"
	"testing"
)

var ErrEmissionIbanNotExist = errors.New("отправка денег на несуществующий эмисионный счет")
var ErrInsufficientFunds = errors.New("недостаточно средств")
var ErrSenderIbanNotExist = errors.New("указан неверный счет отправителя")
var ErrReceiverIbanNotExist = errors.New("указан неверный счет уничтожения")

func TestMoneyTransferToEmission(t *testing.T) {
	bankAdmin := User{LastName: "Иванов", FirstName: "Иван"}

	account1 := OpenNewBankAccount(bankAdmin, 100, Emission, Active)

	const emissionSum float64 = 25

	_, err := MoneyTransferToEmission(account1.IBAN, emissionSum)
	if err != nil {
		fmt.Printf("что-то пошло не так")
	}

	got := account1.Amount
	want := float64(125)

	if got != want {
		t.Errorf("got %f, wanted %f", got, want)
	}
}

func TestMoneyTransferToEmissionIbanNotExist(t *testing.T) {
	bankAdmin := User{LastName: "Иванов", FirstName: "Иван"}

	account1 := OpenNewBankAccount(bankAdmin, 100, Emission, Active)

	const emissionSum float64 = 25

	_, err := MoneyTransferToEmission(account1.IBAN+"23", emissionSum)

	assertError(t, err, ErrEmissionIbanNotExist)
}

func TestMoneyTransferToDestruction(t *testing.T) {
	bankAdmin := User{LastName: "Иванов", FirstName: "Иван"}

	account1 := OpenNewBankAccount(bankAdmin, 100, Emission, Active)
	account2 := OpenNewBankAccount(bankAdmin, 0, Destruction, Active)

	const destructionSum float64 = 21

	_, errDestruction := MoneyTransferToDestruction(account1.IBAN, account2.IBAN, destructionSum)
	if errDestruction != nil {
		fmt.Println("Error:", errDestruction)
	}

	gotEmission := account1.Amount
	wantEmission := float64(79)

	gotDestruction := account2.Amount
	wantDestruction := float64(21)

	if gotEmission != wantEmission {
		t.Errorf("got %f, wanted %f", gotEmission, wantEmission)
	}

	if gotDestruction != wantDestruction {
		t.Errorf("got %f, wanted %f", gotDestruction, wantDestruction)
	}
}

func TestMoneyTransferToDestructionInsufficientFunds(t *testing.T) {
	bankAdmin := User{LastName: "Иванов", FirstName: "Иван"}

	account1 := OpenNewBankAccount(bankAdmin, 100, Emission, Active)
	account2 := OpenNewBankAccount(bankAdmin, 0, Destruction, Active)

	const destructionSum float64 = 101

	_, errDestruction := MoneyTransferToDestruction(account1.IBAN, account2.IBAN, destructionSum)

	assertError(t, errDestruction, ErrInsufficientFunds)
}

func TestMoneyTransferToDestructionIncorrectReceiverIban(t *testing.T) {
	bankAdmin := User{LastName: "Иванов", FirstName: "Иван"}

	account2 := OpenNewBankAccount(bankAdmin, 0, Destruction, Active)

	const destructionSum float64 = 101

	_, errDestruction := MoneyTransferToDestruction("incorrectIban", account2.IBAN, destructionSum)

	assertError(t, errDestruction, ErrReceiverIbanNotExist)
}

func TestMoneyTransferToDestructionIncorrectSenderIban(t *testing.T) {
	bankAdmin := User{LastName: "Иванов", FirstName: "Иван"}

	account1 := OpenNewBankAccount(bankAdmin, 100, Emission, Active)

	const destructionSum float64 = 101

	_, errDestruction := MoneyTransferToDestruction(account1.IBAN, "incorrectIban", destructionSum)

	assertError(t, errDestruction, ErrSenderIbanNotExist)
}

func TestMoneyTransfer(t *testing.T) {
	bankAdmin := User{LastName: "Иванов", FirstName: "Иван"}

	account1 := OpenNewBankAccount(bankAdmin, 100, Emission, Active)
	account2 := OpenNewBankAccount(bankAdmin, 0, Destruction, Active)

	const destructionSum float64 = 21

	_, errDestruction := MoneyTransfer(account1.IBAN, account2.IBAN, destructionSum)
	if errDestruction != nil {
		fmt.Println("Error:", errDestruction)
	}

	gotEmission := account1.Amount
	wantEmission := float64(79)

	gotDestruction := account2.Amount
	wantDestruction := float64(21)

	if gotEmission != wantEmission {
		t.Errorf("got %f, wanted %f", gotEmission, wantEmission)
	}

	if gotDestruction != wantDestruction {
		t.Errorf("got %f, wanted %f", gotDestruction, wantDestruction)
	}
}

func TestMoneyTransferRunSafelyConcurrently(t *testing.T) {
	bankAdmin := User{LastName: "Иванов", FirstName: "Иван"}

	account1 := OpenNewBankAccount(bankAdmin, 1000, Emission, Active)
	account2 := OpenNewBankAccount(bankAdmin, 0, Destruction, Active)

	var wg sync.WaitGroup
	wg.Add(1000)
	for i := 0; i < 1000; i++ {
		go func() {
			_, errDestruction := MoneyTransfer(account1.IBAN, account2.IBAN, float64(1))
			if errDestruction != nil {
				fmt.Println("Error:", errDestruction)
			}
			wg.Done()
		}()

	}
	wg.Wait()

	gotEmission := account1.Amount
	wantEmission := float64(0)

	gotDestruction := account2.Amount
	wantDestruction := float64(1000)

	if gotEmission != wantEmission {
		t.Errorf("got %f, wanted %f", gotEmission, wantEmission)
	}

	if gotDestruction != wantDestruction {
		t.Errorf("got %f, wanted %f", gotDestruction, wantDestruction)
	}
}

func assertError(t testing.TB, got, want error) {
	t.Helper()
	if got == nil {
		t.Fatal("не получил ошибку")
	}
}
