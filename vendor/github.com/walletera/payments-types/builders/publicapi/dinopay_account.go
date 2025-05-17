package publicapi

import (
    "github.com/walletera/payments-types/publicapi"
)

type DinopayAccountBuilder struct {
    institutionName publicapi.OptString
    institutionId   publicapi.OptString
    currency        publicapi.Currency
    accountHolder   string
    accountNumber   string
}

func NewDinopayAccountBuilder() *DinopayAccountBuilder {
    return &DinopayAccountBuilder{
        currency: publicapi.CurrencyUSD,
    }
}

func (b *DinopayAccountBuilder) WithInstitutionName(name string) *DinopayAccountBuilder {
    b.institutionName = publicapi.OptString{
        Value: name,
        Set:   true,
    }
    return b
}

func (b *DinopayAccountBuilder) WithInstitutionId(id string) *DinopayAccountBuilder {
    b.institutionId = publicapi.OptString{
        Value: id,
        Set:   true,
    }
    return b
}

func (b *DinopayAccountBuilder) WithCurrency(currency publicapi.Currency) *DinopayAccountBuilder {
    b.currency = currency
    return b
}

func (b *DinopayAccountBuilder) WithAccountHolder(accountHolder string) *DinopayAccountBuilder {
    b.accountHolder = accountHolder
    return b
}

func (b *DinopayAccountBuilder) WithAccountNumber(accountNumber string) *DinopayAccountBuilder {
    b.accountNumber = accountNumber
    return b
}

func (b *DinopayAccountBuilder) Build() publicapi.Account {
    return publicapi.Account{
        InstitutionName: b.institutionName,
        InstitutionId:   b.institutionId,
        Currency:        b.currency,
        AccountDetails: publicapi.AccountAccountDetails{
            OneOf: publicapi.NewDinopayAccountDetailsAccountAccountDetailsSum(
                publicapi.DinopayAccountDetails{
                    AccountType:   "dinopay",
                    AccountHolder: b.accountHolder,
                    AccountNumber: b.accountNumber,
                },
            ),
        },
    }
}
