Feature: Get payment
  Walletera customers can retrieve a specific payment
  using walletera API

  Background: the payments service is up and running
    Given a running payments service

  Scenario: A customer can retrieve an outbound payment by id
    Given the following payment:
    """json
    {
      "id": "8e38b2f9-af7d-4a80-a9ed-6f5f395004dd",
      "amount": 100,
      "currency": "ARS",
      "beneficiary": {
        "bankName": "Banco Macro",
        "bankId": "123456",
        "accountHolder": "Ricardo Alfonsin",
        "accountNumber": "123456789",
        "accountNumberType": "ARBank",
        "routingKey": "123456789123456789"
       }
    }
    """
    And the payments service receive a PATCH request to update the payment to status: "confirmed"
    When the payments service receive a GET request to retrieve the payment with id: "8e38b2f9-af7d-4a80-a9ed-6f5f395004dd"
    Then the payments service returns the following response:
    """json
    {
      "id": "8e38b2f9-af7d-4a80-a9ed-6f5f395004dd",
      "amount": 100,
      "currency": "ARS",
      "direction": "outbound",
      "customerId": "${json-unit.regex}^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$",
      "externalId": "${json-unit.regex}^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$",
      "beneficiary": {
          "bankName": "Banco Macro",
          "bankId": "123456",
          "accountHolder": "Ricardo Alfonsin",
          "accountNumber": "123456789",
          "accountNumberType": "ARBank",
          "routingKey": "123456789123456789"
      },
      "status": "confirmed",
      "createdAt": "${json-unit.any-string}"
    }
    """

  Scenario: A customer try to retrieve a non existent outbound payment
    When the payments service receive a GET request to retrieve the payment with id: "01939cfe-849e-79c4-b2aa-285522817e69"
    Then the payments service returns 404 Not Found

