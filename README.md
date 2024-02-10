# go-rebalance

`go-rebalance` assists in rebalancing an investment portfolio towards a desired distribution.

## Installation

```
go install ./cmd/rebalance
```

### Dependencies

#### COIN-OR linear programming solver

Refer to the upstream documentation for installation instructions.
See https://www.coin-or.org/downloading/.


## Usage

1. Fetch your data from Avanza, e.g.:

```
rebalance avanza --username 1111111 fetch
```

2. Calculate transfers for rebalancing, e.g.:

```
rebalance avanza --username 1111111 calculate --account-id 2222222
```

The account id is normally the same as the account number.

3. Sanity check the suggested transfers.

4. Carry out the transfers using your favorite Avanza UI.
