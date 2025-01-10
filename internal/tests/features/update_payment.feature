Feature: Update outbound payment
  An outbound payment in pending status can be updated to confirmed or failed status

  Background: the payments service is up and running
    Given a running payments service
      And a running payments events consumer with queueName: "updatePaymentTestQueue"

  Scenario: a pending payment is successfully updated to confirmed
    Given an authorized walletera customer
      And a payment in pending status
     When the payments service receive a PATCH request to update the payment to status: "confirmed"
     Then the payment is updated to status: "confirmed"
      And the payments service publish the following event:
    """json
    {
      "id": "${json-unit.regex}^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$",
      "type": "PaymentUpdated",
      "data": {
        "paymentId": "${json-unit.regex}^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$",
        "externalId": "${json-unit.regex}^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$",
        "status": "confirmed"
      }
    }
    """

  Scenario: a pending payment is successfully updated to failed
    Given an authorized walletera customer
      And a payment in pending status
     When the payments service receive a PATCH request to update the payment to status: "failed"
     Then the payment is updated to status: "failed"
      And the payments service publish the following event:
    """json
    {
      "id": "${json-unit.regex}^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$",
      "type": "PaymentUpdated",
      "data": {
        "paymentId": "${json-unit.regex}^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$",
        "externalId": "${json-unit.regex}^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$",
        "status": "failed"
      }
    }
    """