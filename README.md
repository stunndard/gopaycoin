# gopaycoin

![](https://github.com/stunndard/gopaycoin/blob/master/public/img/screenshot.png?raw=true)

## What is it
gopaycoin is a Bitcoin payment gateway allowing anyone to accept Bitcoins payments on their website without relying on any 3rd party
to accept your payments. Why trust your money to someone else even if they look good? And pay them fees too? Thanks, but no.

## Features
- Accepts BTC payments
- Shows nice invoice page to the customer
- Processes refunds automatically if payment is not enough or overpaid
- Processes withdrawals to your cold wallet automatically
- Very easy integration with your website or service

## How it works
- You create a payment with required amount in USD by making just one API call
- It will convert the USD amount to BTC using the current market rates
- You can redirect the user to the gopaycoin invoice page so they can complete the payment and see the payment progress. Or you can show them the payment progress by yourself
- You will receive callbacks when the payment is pending and when it is complete
- Payments, refunds and withdrawals to your cold wallet are processed automatically with no manual intervention

## Requirements
- Full Bitcoin Core node version > 0.14.1 is required. We are completely self-hosted, ain't we?
- Mysql for gopaycoin database

## API documentation
- Soon

## Work in progress!!!
- Not recommended to use in production yet.

## Planned:
- Docker and docker-compose. Deploy it with just one command
- Admin web interface

