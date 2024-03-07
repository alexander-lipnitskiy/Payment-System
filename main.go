package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jacoelho/banking/iban"
	"slices"
)

type AccountType int
type AccountStatus int

const (
	Emission AccountType = iota
	Destruction
	Individual
	LegalEntity
)

const (
	Active AccountStatus = iota
	Blocked
)

func (accountType AccountType) String() string {
	switch accountType {
	case Emission:
		return "Специальный счет для эмиссии"
	case Destruction:
		return "Специальный счет для уничтожения"
	case Individual:
		return "Физическое лицо"
	case LegalEntity:
		return "Юридическое лицо"

	default:
		return "Неверный тип аккаунта"
	}
}

func (accountStatus AccountStatus) String() string {
	switch accountStatus {
	case Active:
		return "Активен"
	case Blocked:
		return "Заблокирован"

	default:
		return "Неверный аккаунт статус"
	}
}

var bankAccounts []*Account

type User struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

type Account struct {
	IBAN          string        `json:"iban"`
	Amount        float64       `json:"amount"`
	Currency      string        `json:"currency"`
	AccountStatus AccountStatus `json:"accountStatus"`
	AccountType   AccountType   `json:"accountType"`
	User          User          `json:"user"`
}

type JsonTransfer struct {
	IbanFrom string  `json:"ibanFrom"`
	IbanTo   string  `json:"ibanTo"`
	Amount   float64 `json:"amount"`
}

func GenerateIban(countryCode string) string {
	ibn, err := iban.Generate("BY")
	if err != nil {
		fmt.Println(err)
	}

	return ibn
}

func newAccount(amount float64, currency string, user User, accountType AccountType,
	accountStatus AccountStatus) *Account {
	ibn := GenerateIban("BY")
	return &Account{
		IBAN:          ibn,
		Amount:        amount,
		Currency:      currency,
		AccountStatus: accountStatus,
		User:          user,
		AccountType:   accountType,
	}
}

func OpenNewBankAccount(user User, amount float64) *Account {
	newAccount := newAccount(amount, "BYN", user, Individual, Active)
	bankAccounts = append(bankAccounts, newAccount)

	return newAccount
}

func (account *Account) DisplayAccount() {
	fmt.Printf("account = %+v\n", account)
}

func (user *User) DisplayUser() {
	fmt.Printf("user = %+v\n", user)
}

func ShowEmissionAccountInfo() (string, error) {
	idxEmission := slices.IndexFunc(bankAccounts, func(c *Account) bool { return c.AccountType == Emission })

	if idxEmission == -1 {
		return "", errors.New("не найден специальный счет для эмиссии")
	}

	return bankAccounts[idxEmission].IBAN, nil
}

func ShowDestructionAccountInfo() (string, error) {
	idxDestruction := slices.IndexFunc(bankAccounts, func(c *Account) bool { return c.AccountType == Destruction })

	if idxDestruction == -1 {
		return "", errors.New("не найден специальный счет для уничтожения")
	}

	return bankAccounts[idxDestruction].IBAN, nil
}

func MoneyTransferToEmission(ibanEmission string, amount float64) (bool, error) {
	idxTo := slices.IndexFunc(bankAccounts, func(c *Account) bool {
		return c.IBAN == ibanEmission && c.AccountType == Emission
	})

	if idxTo == -1 {
		return false, errors.New("отправка денег на несуществующий эмисионный счет")
	}

	bankAccounts[idxTo].Amount = bankAccounts[idxTo].Amount + amount

	return true, nil
}

func MoneyTransferToDestruction(ibanFrom string, ibanDestruction string, amount float64) (bool, error) {
	idxFrom := slices.IndexFunc(bankAccounts, func(c *Account) bool { return c.IBAN == ibanFrom })
	if idxFrom == -1 {
		return false, errors.New("указан неверный счет отправителя")
	}

	idxTo := slices.IndexFunc(bankAccounts, func(c *Account) bool {
		return c.IBAN == ibanDestruction && c.AccountType == Destruction
	})

	if idxTo == -1 {
		return false, errors.New("указан неверный счет уничтожения")
	}

	if bankAccounts[idxFrom].Amount < amount {
		return false, errors.New("недостаточно средств")
	}

	bankAccounts[idxFrom].Amount = bankAccounts[idxFrom].Amount - amount
	bankAccounts[idxTo].Amount = bankAccounts[idxTo].Amount + amount

	return true, nil
}

func MoneyTransfer(ibanFrom string, ibanTo string, amount float64) (bool, error) {
	idxFrom := slices.IndexFunc(bankAccounts, func(c *Account) bool { return c.IBAN == ibanFrom })
	if idxFrom == -1 {
		return false, errors.New("указан неверный счет отправителя")
	}

	idxTo := slices.IndexFunc(bankAccounts, func(c *Account) bool { return c.IBAN == ibanTo })

	if idxTo == -1 {
		return false, errors.New("указан неверный счет получателя")
	}

	if bankAccounts[idxFrom].Amount < amount {
		return false, errors.New("недостаточно средств")
	}

	bankAccounts[idxFrom].Amount = bankAccounts[idxFrom].Amount - amount
	bankAccounts[idxTo].Amount = bankAccounts[idxTo].Amount + amount

	return true, nil
}

func MoneyTransferJson(transactionJson []byte) (bool, error) {
	var jsonTransfer JsonTransfer

	if err := json.Unmarshal(transactionJson, &jsonTransfer); err != nil {
		return false, errors.New("неверное декодирование данных JSON")
	}

	idxFrom := slices.IndexFunc(bankAccounts, func(c *Account) bool { return c.IBAN == jsonTransfer.IbanFrom })
	if idxFrom == -1 {
		return false, errors.New("указан неверный счет отправителя")
	}

	idxTo := slices.IndexFunc(bankAccounts, func(c *Account) bool { return c.IBAN == jsonTransfer.IbanTo })

	if idxTo == -1 {
		return false, errors.New("указан неверный счет получателя")
	}

	if bankAccounts[idxFrom].Amount < jsonTransfer.Amount {
		return false, errors.New("недостаточно средств")
	}

	bankAccounts[idxFrom].Amount = bankAccounts[idxFrom].Amount - jsonTransfer.Amount
	bankAccounts[idxTo].Amount = bankAccounts[idxTo].Amount + jsonTransfer.Amount

	return true, nil
}

func main() {
	// инициализация специальных счетов
	bankAdmin := User{LastName: "Иванов", FirstName: "Иван"}

	emissionAccount := newAccount(100, "BYN", bankAdmin, Emission, Active)
	destructionAccount := newAccount(0, "BYN", bankAdmin, Destruction, Active)

	bankAccounts = append(bankAccounts, emissionAccount)
	bankAccounts = append(bankAccounts, destructionAccount)

	// номер специального счета для “эмиссии”
	emissionAccountIban, err := ShowEmissionAccountInfo()
	if err != nil {
		return
	}

	fmt.Println("Номер специального счета для эмиссии:", emissionAccountIban)

	// номер специального счета для “уничтожения”
	destructionAccountIban, err := ShowDestructionAccountInfo()
	if err != nil {
		return
	}

	fmt.Println("Номер специального счета для уничтожения:", destructionAccountIban)

	// добавлению на счет “эмиссии” указанной суммы
	const emissionSum float64 = 25

	_, err = MoneyTransferToEmission(emissionAccountIban, emissionSum)
	if err != nil {
		fmt.Printf("что-то пошло не так")
	}

	fmt.Printf("Добавлению на счет “эмиссии” указанной суммы (%f BYN):\n", emissionSum)

	// отправка определенной суммы денег с указанного счета на счет “уничтожения”
	const destructionSum float64 = 21

	_, errDestruction := MoneyTransferToDestruction(emissionAccountIban, destructionAccountIban, destructionSum)
	if errDestruction != nil {
		fmt.Printf("что-то пошло не так")
	}

	fmt.Printf("Отправка денег с указанного счета на счет “уничтожения” (%f BYN):\n", destructionSum)

	// открытие нового счета
	const account1Sum float64 = 100
	const account2Sum float64 = 0
	const account3Sum float64 = 40

	account1 := OpenNewBankAccount(User{FirstName: "Василь", LastName: "Быков"}, account1Sum)
	account2 := OpenNewBankAccount(User{FirstName: "Светлана", LastName: "Александровна"}, account2Sum)
	account3 := OpenNewBankAccount(User{FirstName: "Линда", LastName: "Комарова"}, account3Sum)

	// перевод заданной суммы денег между двумя указанными счетами (с несколькими параметрами)
	_, errTransfer := MoneyTransfer(account1.IBAN, account2.IBAN, 22)
	if errTransfer != nil {
		fmt.Println("Error:", err)
	}

	// перевод заданной суммы денег между двумя указанными счетами (с единственным параметром в формате json)
	const transferAmount float64 = 15

	jsonString := fmt.Sprintf(`{"ibanFrom": "%s","ibanTo": "%s", "amount": %f}`,
		account1.IBAN, account3.IBAN, transferAmount)
	transactionJson := []byte(jsonString)

	_, errTransferJson := MoneyTransferJson(transactionJson)
	if errTransferJson != nil {
		fmt.Println("Error:", err)
	}

	// вывод списка всех счетов
	for _, s := range bankAccounts {
		fmt.Printf("\n%s %s\n%s\n%f%s\n%s\n%s\n", s.User.LastName, s.User.FirstName, s.IBAN, s.Amount,
			s.Currency, s.AccountStatus, s.AccountType)
	}
}
