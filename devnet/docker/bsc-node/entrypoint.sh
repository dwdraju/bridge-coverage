#!/bin/sh
geth --datadir node1 \
    --syncmode 'full' \
    --port 30310 \
    --rpc --rpcaddr 0.0.0.0 \
    --rpccorsdomain "*" \
    --rpcport 8545 \
    --ws --wsaddr 0.0.0.0 \
    --wsport 8546 \
    --rpcapi personal,eth,net,web3,txpool,miner \
    --networkid 97 \
    --mine \
    --unlock '0x48948297C3236ec3eA6c95f4eEc22fDb18255E55' \
    --password .secret \
    --allow-insecure-unlock