# User Stories

## Scenario 1

    - As a user I should be able to create account easily with my data

### Acceptance Criteria

    - User can create account on the website 
    - Account would be verified via code sent to the user
    - user should input the verification code within 60 seconds
---

## Scenario 2

    - As a user I should sign in with my email and password and apply for forgot password

### Acceptance Criteria

    - User can sign in to the website with the right email and password 
    - If user forgot the password, it should be changed
    - user would receive verification code within 60 seconds and then would be able to update the password
---

## Scenario 3

    - As a user I should be able to update my data anytime (name, password, ssh_key)

### Acceptance Criteria

    - User can login then go to the profile page to update his data such as name, password and ssh_key
    - User can't update his email
---

## Scenario 4

    - As a User I expect to deal with user-friendly interface with plain colors

### Acceptance Criteria

    - Website would be simple, user-friendly with eye pleasing colors
---

## Scenario 5

    - As a user I should be able to logout anytime and stay logged in multiple days 

### Acceptance Criteria

    - User can logout from the website anytime 
    - When user logs to the website he will stay logged in for a while as log as he didn't logout 
---

## Scenario 6

    - As a user I should be able to apply for a voucher to use the grid for deployment

### Acceptance Criteria

    - User can apply for voucher from the interface
    - User would receive a confirmation mail whether the application is accepted or not 
---

## Scenario 7

    - As a user I should be able to activate a voucher to use the grid for deployment

### Acceptance Criteria

    - User can activate a voucher from the interface using profile page
---

## Scenario 8

    - As a user I expect to get all information about the voucher, used resources, and remaining balance 

### Acceptance Criteria

    - User should get all information about the voucher and its available balance
    - Each user will have certain amount of money based on the voucher 
---

## Scenario 9

    - As a user I expect to know all about how to use the website

### Acceptance Criteria

    - User should know all about the website and how to use it, apply for vouchers, deploy and so on
---

## Scenario 10

    - As a user I expect to deploy fast and easily without any complex steps 

### Acceptance Criteria

    - User can deploy vm or k8s with choosing the name, resources to be deployed
    - User will be given all needed information to access vm or k8s as resources, public ip, and planetary network ip
    - If there's any error, all logs of deployment will be shown to the user
---

## Scenario 11

    - As a user I expect to cancel my deployments anytime 

### Acceptance Criteria

    - User can cancel any specific deployment or the whole deployment easily from the interface
    - If there's any error, all logs of deployment will be shown to the user
---
