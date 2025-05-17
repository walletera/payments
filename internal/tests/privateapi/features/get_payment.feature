Feature: Get payment
  Walletera customers can retrieve a specific payment
  using walletera API

  Background: the payments service is up and running
    Given a running payments service

  Scenario: A customer can retrieve an outbound payment by id
    Given an authorized dinopay-gateway service
      And the following payment:
    """json
    {
      "id": "63c3b924-ddc6-4a28-9baf-6eb0aa4110f0",
      "customerId": "abbb8aa3-87f9-4b2b-889f-8962cf708cfc",
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
    When the payments service receive a GET request to retrieve the payment with id: "63c3b924-ddc6-4a28-9baf-6eb0aa4110f0"
    Then the payments service returns the following response:
    """json
    {
      "id": "63c3b924-ddc6-4a28-9baf-6eb0aa4110f0",
      "customerId": "abbb8aa3-87f9-4b2b-889f-8962cf708cfc",
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
      },
      "createdAt": "${json-unit.any-string}"
    }
    """

  Scenario: A customer try to retrieve a non existent outbound payment
    Given an authorized dinopay-gateway service
     When the payments service receive a GET request to retrieve the payment with id: "01939cfe-849e-79c4-b2aa-285522817e69"
     Then the payments service returns 404 Not Found

