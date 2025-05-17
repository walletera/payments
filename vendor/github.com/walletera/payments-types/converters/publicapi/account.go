package publicapi

import (
    "github.com/walletera/payments-types/privateapi"
    "github.com/walletera/payments-types/publicapi"
)

// Convert transforms a publicapi.Account to a privateapi.Account.
func Convert(account publicapi.Account) privateapi.Account {
    var out privateapi.Account
    out.InstitutionName = privateapi.OptString(account.InstitutionName)
    out.InstitutionId = privateapi.OptString(account.InstitutionId)
    out.Currency = privateapi.Currency(account.Currency)

    details := account.AccountDetails.OneOf
    switch details.Type {
    case publicapi.CvuAccountDetailsAccountAccountDetailsSum:
        if cvu, ok := details.GetCvuAccountDetails(); ok {
            out.AccountDetails = privateapi.AccountAccountDetails{
                OneOf: privateapi.NewCvuAccountDetailsAccountAccountDetailsSum(
                    convertCvuAccountDetails(cvu),
                ),
            }
        }
    case publicapi.DinopayAccountDetailsAccountAccountDetailsSum:
        if dinopay, ok := details.GetDinopayAccountDetails(); ok {
            out.AccountDetails = privateapi.AccountAccountDetails{
                OneOf: privateapi.NewDinopayAccountDetailsAccountAccountDetailsSum(
                    convertDinopayAccountDetails(dinopay),
                ),
            }
        }
    }
    return out
}

func convertCvuAccountDetails(cvu publicapi.CvuAccountDetails) privateapi.CvuAccountDetails {
    pubCvu := privateapi.CvuAccountDetails{
        AccountType: cvu.AccountType,
        Cuit:        privateapi.OptString(cvu.Cuit),
        RoutingInfo: convertCvuRoutingInfo(cvu.RoutingInfo.OneOf),
    }
    return pubCvu
}

func convertCvuRoutingInfo(ri publicapi.CvuAccountDetailsRoutingInfoSum) privateapi.CvuAccountDetailsRoutingInfo {
    var pubRouting privateapi.CvuAccountDetailsRoutingInfo
    switch ri.Type {
    case publicapi.CvuCvuRoutingInfoCvuAccountDetailsRoutingInfoSum:
        pubRouting.OneOf = privateapi.CvuAccountDetailsRoutingInfoSum{
            Type: privateapi.CvuCvuRoutingInfoCvuAccountDetailsRoutingInfoSum,
            CvuCvuRoutingInfo: privateapi.CvuCvuRoutingInfo{
                Cvu: ri.CvuCvuRoutingInfo.Cvu,
            },
        }
    case publicapi.AliasCvuRoutingInfoCvuAccountDetailsRoutingInfoSum:
        pubRouting.OneOf = privateapi.CvuAccountDetailsRoutingInfoSum{
            Type: privateapi.AliasCvuRoutingInfoCvuAccountDetailsRoutingInfoSum,
            AliasCvuRoutingInfo: privateapi.AliasCvuRoutingInfo{
                CvuRoutingInfoType: ri.AliasCvuRoutingInfo.CvuRoutingInfoType,
                Alias:              ri.AliasCvuRoutingInfo.Alias,
            },
        }
    }
    return pubRouting
}

func convertDinopayAccountDetails(dinopay publicapi.DinopayAccountDetails) privateapi.DinopayAccountDetails {
    return privateapi.DinopayAccountDetails{
        AccountType:   dinopay.AccountType,
        AccountHolder: dinopay.AccountHolder,
        AccountNumber: dinopay.AccountNumber,
    }
}
