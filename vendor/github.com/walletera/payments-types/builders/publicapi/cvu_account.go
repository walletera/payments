package publicapi

import (
    "github.com/walletera/payments-types/publicapi"
    "github.com/walletera/werrors"
)

type CVUAccountBuilder struct {
    InstitutionName publicapi.OptString
    InstitutionId   publicapi.OptString
    cvu             string
    alias           string
}

func NewCVUAccountBuilder() *CVUAccountBuilder {
    return &CVUAccountBuilder{}
}

func (b *CVUAccountBuilder) WithInstitutionName(name string) *CVUAccountBuilder {
    b.InstitutionName = publicapi.NewOptString(name)
    return b
}

func (b *CVUAccountBuilder) WithInstitutionId(id string) *CVUAccountBuilder {
    b.InstitutionId = publicapi.NewOptString(id)
    return b
}

func (b *CVUAccountBuilder) WithCVU(cvu string) *CVUAccountBuilder {
    b.cvu = cvu
    return b
}

func (b *CVUAccountBuilder) WithAlias(alias string) *CVUAccountBuilder {
    b.alias = alias
    return b
}

func (b *CVUAccountBuilder) Build() (publicapi.Account, werrors.WError) {
    var accountDetails publicapi.AccountAccountDetails
    if len(b.cvu) > 0 {
        accountDetails = publicapi.AccountAccountDetails{
            OneOf: publicapi.NewCvuAccountDetailsAccountAccountDetailsSum(
                publicapi.CvuAccountDetails{
                    AccountType: "cvu",
                    RoutingInfo: publicapi.CvuAccountDetailsRoutingInfo{
                        OneOf: publicapi.NewCvuCvuRoutingInfoCvuAccountDetailsRoutingInfoSum(
                            publicapi.CvuCvuRoutingInfo{
                                CvuRoutingInfoType: "cvu",
                                Cvu:                b.cvu,
                            },
                        ),
                    },
                },
            ),
        }
    } else if len(b.alias) > 0 {
        accountDetails = publicapi.AccountAccountDetails{
            OneOf: publicapi.NewCvuAccountDetailsAccountAccountDetailsSum(
                publicapi.CvuAccountDetails{
                    AccountType: "cvu",
                    RoutingInfo: publicapi.CvuAccountDetailsRoutingInfo{
                        OneOf: publicapi.NewAliasCvuRoutingInfoCvuAccountDetailsRoutingInfoSum(
                            publicapi.AliasCvuRoutingInfo{
                                CvuRoutingInfoType: "alias",
                                Alias:              b.alias,
                            },
                        ),
                    },
                },
            ),
        }
    } else {
        return publicapi.Account{}, werrors.NewNonRetryableInternalError("either cvu or alias must be set in a cvu account")
    }
    return publicapi.Account{
        InstitutionName: b.InstitutionName,
        InstitutionId:   b.InstitutionId,
        Currency:        publicapi.CurrencyARS,
        AccountDetails:  accountDetails,
    }, nil
}
