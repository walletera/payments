Feature: Get payment
  Walletera customers can retrieve a specific payment
  using walletera API

  Background: the payments service is up and running
    Given a running payments service

  Scenario: A customer can retrieve an outbound payment by id
    Given an authorized walletera customer
      And the following payment:
    """
    testdata/successful_outbound_bind_payment_af2e.json
    """
    When the payments service receive a GET request to retrieve the payment with id: "af2e70dd-bd96-4be3-9f7b-4c2ef9d72c2e"
    Then the payments service returns the following response:
    """
    testdata/successful_outbound_bind_payment_af2e_response.json
    """

  Scenario: A customer try to retrieve a non existent outbound payment
    Given an authorized walletera customer
     When the payments service receive a GET request to retrieve the payment with id: "01939cfe-849e-79c4-b2aa-285522817e69"
     Then the payments service returns 404 Not Found

