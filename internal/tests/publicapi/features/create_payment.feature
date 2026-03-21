Feature: Create outbound payment

  Background: the payments service is up and running
    Given a running payments service
      And a running payments events consumer with queueName: "createPaymentTestQueue"

  Scenario: a cvu payment is created successfully
    Given an authorized walletera customer
    When  the customer sends the following payment to the payments endpoint:
    """
    testdata/successful_outbound_bind_payment_bdf4.json
    """
    Then the endpoint returns the http status code 201
    Then the payments service publish the following event:
    """
    testdata/payment_created_event.json
    """

  Scenario: payment creation failed due to missing authentication token
    Given an unauthorized walletera customer
    When  the customer sends the following payment to the payments endpoint:
    """
    testdata/successful_outbound_bind_payment_bdf4.json
    """
    Then the endpoint returns the http status code 401

  Scenario: payment creation failed due to invalid authentication token
    Given a walletera customer with an invalid token
    When  the customer sends the following payment to the payments endpoint:
    """
    testdata/successful_outbound_bind_payment_bdf4.json
    """
    Then the endpoint returns the http status code 401

  Scenario: payment creation fails with bad request when amount is zero
    Given an authorized walletera customer
    When  the customer sends the following payment to the payments endpoint:
    """
    testdata/payment_with_zero_amount.json
    """
    Then the endpoint returns the http status code 400

  Scenario: payment creation fails with bad request when amount is negative
    Given an authorized walletera customer
    When  the customer sends the following payment to the payments endpoint:
    """
    testdata/payment_with_negative_amount.json
    """
    Then the endpoint returns the http status code 400

  Scenario: payment creation fails with conflict when payment already exists
    Given an authorized walletera customer
    When  the customer sends the following payment to the payments endpoint:
    """
    testdata/successful_outbound_bind_payment_bdf4.json
    """
    And the customer sends the following payment to the payments endpoint:
    """
    testdata/successful_outbound_bind_payment_bdf4.json
    """
    Then the endpoint returns the http status code 409