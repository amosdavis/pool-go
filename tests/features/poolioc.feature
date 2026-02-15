Feature: Low-level poolioc device operations
  As a developer using the poolioc package
  I want to interact with the POOL kernel module via ioctls
  So that I can build POOL-based applications

  Background:
    Given the POOL kernel module is loaded

  Scenario: Open and close device
    When I open the POOL device
    Then the device file descriptor should be valid
    And I close the device without error

  Scenario: Start and stop listener
    Given I have an open POOL device
    When I start listening on port 9253
    Then the listener should be active
    When I stop the listener
    Then the listener should be stopped

  Scenario: Connect to a remote peer
    Given I have an open POOL device
    And a remote POOL peer is listening on "127.0.0.1:9253"
    When I connect to "127.0.0.1:9253"
    Then the session should be established
    And the session index should be non-negative

  Scenario: Connect to an IPv6 peer
    Given I have an open POOL device
    And a remote POOL peer is listening on "[::1]:9253"
    When I connect to "[::1]:9253"
    Then the session should be established

  Scenario: Send and receive data
    Given I have an established session
    When I send "hello pool" on channel 0
    And the peer echoes the data back
    Then I should receive "hello pool" on channel 0

  Scenario: List sessions
    Given I have an established session
    When I list sessions
    Then the session list should contain at least 1 session
    And the session state should be "ESTABLISHED"

  Scenario: Close a session
    Given I have an established session
    When I close the session
    Then the session should be removed from the list

  Scenario: Subscribe to a channel
    Given I have an established session
    When I subscribe to channel 5
    Then channel 5 should be active
    When I unsubscribe from channel 5
    Then channel 5 should be inactive

  Scenario: Channel bitmap listing
    Given I have an established session
    And I subscribe to channels 1, 3, and 7
    When I list channels
    Then the bitmap should show channels 1, 3, and 7 as active

  Scenario: Send large payload
    Given I have an established session
    When I send a 4000-byte payload
    And the peer echoes the data back
    Then I should receive a 4000-byte payload

  Scenario: Connection refused
    Given I have an open POOL device
    When I connect to "192.0.2.1:9999"
    Then the connection should fail with a timeout or unreachable error

  Scenario: IPv4-mapped helper functions
    When I convert "10.0.0.1" to an IPv4-mapped IPv6 address
    Then the result should be "::ffff:10.0.0.1"
    And IsV4Mapped should return true

  Scenario: Ioctl number encoding
    Then POOL_IOC_LISTEN should have type byte 0x50
    And POOL_IOC_CONNECT should have direction bits set to WRITE
