Feature: Create outbound payment
  Walletera customers with funds on their accounts can create outbound payments

  Background: the dinopay-gateway is up and running
    Given a running payments service

  Scenario: a payment is created successfully
    Given a walletera customer
    When  the customer sends the following payment to the payments endpoint:
    """json
    {
      "amount": 100,
      "currency": "ARS",
      "beneficiary": {
        "bankName": "dinopay",
        "bankId": "dinopay",
        "accountHolder": "John Doe",
        "routingKey": "123456789123456"
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
        "beneficiary": {
          "bankName": "dinopay",
          "bankId": "dinopay",
          "accountHolder": "John Doe",
          "routingKey": "123456789123456"
        },
        "createdAt": "${json-unit.any-string}"
      }
    }
    """
