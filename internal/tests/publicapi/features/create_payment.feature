Feature: Create outbound payment

  Background: the payments service is up and running
    Given a running payments service
      And a running payments events consumer with queueName: "createPaymentTestQueue"

  Scenario: a payment is created successfully
    Given an authorized walletera customer
    When  the customer sends the following payment to the payments endpoint:
    """json
    {
      "id": "bdf48329-d870-4fb4-882a-0fa0aef28a63",
      "amount": 100,
      "currency": "ARS",
      "gateway": "bind",
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
        "id": "${json-unit.regex}^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$",
        "amount": 100,
        "currency": "ARS",
        "direction": "outbound",
        "customerId": "${json-unit.regex}^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$",
        "status": "pending",
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

  Scenario: payment creation failed due to missing authentication token
    Given an unauthorized walletera customer
    When  the customer sends the following payment to the payments endpoint:
    """json
    {
      "id": "bdf48329-d870-4fb4-882a-0fa0aef28a63",
      "amount": 100,
      "currency": "ARS",
      "gateway": "bind",
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
    Then the endpoint returns the http status code 401

  Scenario: payment creation failed due to invalid authentication token
    Given a walletera customer with an invalid token
    When  the customer sends the following payment to the payments endpoint:
    """json
    {
      "id": "bdf48329-d870-4fb4-882a-0fa0aef28a63",
      "amount": 100,
      "currency": "ARS",
      "gateway": "bind",
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
    Then the endpoint returns the http status code 401