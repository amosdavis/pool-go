Feature: High-level pool networking API
  As a Go developer
  I want to use net.Conn and net.Listener with POOL
  So that I can write POOL applications idiomatically

  Background:
    Given the POOL kernel module is loaded

  Scenario: Dial and communicate
    Given a POOL echo server on ":9253"
    When I dial "pool" "127.0.0.1:9253"
    And I write "hello world"
    Then I should read "hello world"

  Scenario: Dial IPv6 address
    Given a POOL echo server on ":9253"
    When I dial "pool6" "[::1]:9253"
    And I write "ipv6 test"
    Then I should read "ipv6 test"

  Scenario: Dial with timeout
    When I dial "pool" "192.0.2.1:9999" with a 1 second timeout
    Then the dial should fail with a timeout error

  Scenario: Listen and accept
    Given I listen on "pool" ":9254"
    When a client connects to "127.0.0.1:9254"
    Then Accept should return a connection
    And the remote address should be "127.0.0.1"

  Scenario: Conn implements net.Conn
    Given I have a connected pool.Conn
    Then it should implement net.Conn
    And LocalAddr should return a pool address
    And RemoteAddr should return a pool address

  Scenario: Read deadline exceeded
    Given I have a connected pool.Conn
    When I set a read deadline 100ms in the past
    Then Read should return a timeout error

  Scenario: Write deadline exceeded
    Given I have a connected pool.Conn
    When I set a write deadline 100ms in the past
    Then Write should return a timeout error

  Scenario: Close connection
    Given I have a connected pool.Conn
    When I close the connection
    Then subsequent writes should return ErrClosed
    And subsequent reads should return ErrClosed

  Scenario: Session telemetry
    Given I have a connected pool.Conn
    When I request telemetry
    Then I should receive RTT, jitter, loss, and throughput values

  Scenario: Session state
    Given I have a connected pool.Conn
    When I query the session state
    Then it should be "ESTABLISHED"

  Scenario: Multi-channel communication
    Given I have a connected pool.Conn
    When I open channel 5
    And I write "channel5 data" on channel 5
    And the peer echoes on channel 5
    Then I should read "channel5 data" on channel 5

  Scenario: Close channel
    Given I have a connected pool.Conn
    And I have an open channel 5
    When I close channel 5
    Then writes to channel 5 should return ErrClosed

  Scenario: Address parsing
    When I resolve "pool" "10.0.0.1:9253"
    Then the address network should be "pool"
    And the address string should be "10.0.0.1:9253"

  Scenario: IPv6 address parsing
    When I resolve "pool6" "[2001:db8::1]:9253"
    Then the address network should be "pool"
    And the address string should be "[2001:db8::1]:9253"

  Scenario: Error mapping
    Given a session-full errno
    Then the error should be ErrSessionFull

  Scenario: Concurrent read and write
    Given I have a connected pool.Conn
    When 10 goroutines write concurrently
    And 10 goroutines read concurrently
    Then no data races should occur
