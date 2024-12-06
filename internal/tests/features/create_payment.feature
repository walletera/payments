Feature: Create outbound payment
  Walletera customers with funds on their accounts can create outbound payments

  Background: the payments service is up and running
    Given a running payments service
      And a running payments events consumer with queueName: "createPaymentTestQueue"

  Scenario: a payment is created successfully
    Given a walletera customer
    When  the customer sends the following payment to the payments endpoint:
    """json
    {
      "id": "bdf48329-d870-4fb4-882a-0fa0aef28a63",
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
