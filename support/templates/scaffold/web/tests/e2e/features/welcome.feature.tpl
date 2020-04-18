Feature: welcome
  As a new appy user, I want to be able to understand how it works.

  Scenario: open and read official documentation
    Given I know the documentation URL is "https://appist.gitbook.io/appy/"
    When I open the URL
    Then I should see the introduction page
