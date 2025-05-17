Feature: Create inbound payment

  Background: the payments service is up and running
    Given a running payments service
      And a running payments events consumer with queueName: "createPaymentTestQueue"

  Scenario: a payment is created successfully
    Given an authorized dinopay-gateway service
    When  the dinopay gateway sends the following payment to the payments endpoint:
    """json
    {
      "id": "3b3315ea-38c1-40a4-b7f9-149cc9807096",
      "customerId": "997e76a7-c18d-4337-90ea-9a5cb0c5b54e",
      "amount": 100,
      "currency": "ARS",
      "gateway": "bind",
      "direction": "inbound",
      "status": "confirmed",
      "debtor": {
        "institutionName": "dinopay",
        "institutionId": "dinopay",
        "currency": "ARS",
        "accountDetails": {
          "accountType": "cvu",
          "cuit": "23679876453",
          "routingInfo": {
            "cvuRoutingInfoType": "cvu",
            "cvu": "1122334455667788554433"
          }
        }
      },
      "beneficiary": {
        "institutionName": "dinopay",
        "institutionId": "dinopay",
        "currency": "ARS",
        "accountDetails": {
          "accountType": "cvu",
          "cuit": "23679876453",
           "routingInfo": {
            "cvuRoutingInfoType": "cvu",
            "cvu": "1122334455667788554433"
          }
        }
      }
    }
    """
    Then the endpoint returns the http status code 201
    Then the payments service publish the following event:
    """json
    {
      "id": "${json-unit.regex}^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$",
      "type": "PaymentCreated",
      "data": {
        "id": "3b3315ea-38c1-40a4-b7f9-149cc9807096",
        "amount": 100,
        "currency": "ARS",
        "customerId": "997e76a7-c18d-4337-90ea-9a5cb0c5b54e",
        "direction": "inbound",
        "status": "confirmed",
        "debtor": {
          "institutionName": "dinopay",
          "institutionId": "dinopay",
          "currency": "ARS",
          "accountDetails": {
            "accountType": "cvu",
            "cuit": "23679876453",
            "routingInfo": {
              "cvuRoutingInfoType": "cvu",
              "cvu": "1122334455667788554433"
            }
          }
        },
        "beneficiary": {
          "institutionName": "dinopay",
          "institutionId": "dinopay",
          "currency": "ARS",
          "accountDetails": {
            "accountType": "cvu",
            "cuit": "23679876453",
             "routingInfo": {
              "cvuRoutingInfoType": "cvu",
              "cvu": "1122334455667788554433"
            }
          }
        },
        "createdAt": "${json-unit.any-string}"
      }
    }
    """
