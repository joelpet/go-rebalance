# go-rebalance

`go-rebalance` assists in rebalancing investment portfolios towards to a desired distribution.


## Usage

1. Fetch your data from Avanza, e.g.:

```
rebalance avanza fetch --username 1111111
```

2. Calculate transfers for rebalancing, e.g.:

```
rebalance avanza --username 1111111 calculate --account-id 2222222
```

The account id is normally the same as the account number.

3. Sanity check the suggested transfers.

4. Carry out the transfers using your favorite Avanza UI.
