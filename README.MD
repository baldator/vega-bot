# 🤖 VEGA EVENTS  BOT 📣
The bot is publishing real-time Vega network notification to social media platforms.

## Docker container
You can run the solution in a Docker container:
```
docker run baldator/vega-bot -v src/config.yaml:config.yaml -v src/ethereum.conf:data/ethereum.conf -v src/bots.conf:data/bots.conf
```

## Events
The following events triggers a message:

- New Market Proposal created, updated, enacted
- Market price monitoring auction started/ended
- Network has been reset (network ID has changed/block height reset)
- Rekt alerts (large liquidations)
- Whale alerts (large buys/sells etc)
- Loss socialisation alerts (distribution of funds generated by defaulting traders)

## Dependencies
Vega bot use [post to social API service](https://github.com/cdm/post-to-socials) to send message to socials. Please make sure you have a running instance of the service before running the bot. 

## Configuration
Edit the file `config.yaml` and fill in the required configuration values:
```
SocialServiceURL                => Post to social API URL
SocialTwitterEnabled            => true if you want to enable Twitter, false otherwise
SocialTelegramEnabled           => true if you want to enable Telegram, false otherwise
SocialDiscordEnabled            => true if you want to enable Discord, false otherwise
SocialSlackEnabled              => true if you want to enable Slack, false otherwise
SocialServiceKey                => Post to social API Key
SocialServiceSecret             => Post to social API secret
GrpcNodeUrl                     => URL of the Vega gRPC endpoint
SentryEnabled                   => true if you want to enable Sentry integration
SentryDsn:                      => The Sentry endpoint to send crash information to
PrometheusEnabled               => true if you want to expose Prometheus compatible APM endpoint
PrometheusPort                  => Prometheus endpoint port (default: 2112)
VegaEventsBatchSize             => Vega client default batch size (default value: 5000)
VegaOrdersEnabled               => true if you want the client to listen to orders events (needed if you want to enable Whale alerts)
VegaTradesEnabled               => true if you want the client to listen to trades events (needed if you want to enable Rekt alerts)
VegaProposalsEnabled            => true if you want the client to listen to proposals events
VegaAuctionsEnabled             => true if you want the client to listen to auctions events
VegaLossSocializationEnabled    => true if you want the client to listen to loss socialization events
VegaNetworkParametersEnabled    => true if you want the client to listen to network paramentes events
Debug                           => true if you want to print debug event information
```

## Vega protocol
Vega is a proof of stake blockchain for creating and trading derivatives. It provides infrastructure for decentralised markets that settle in assets held on Ethereum, and in future also Bitcoin and other major collateral blockchains. Vega facilitates high speed, permissionless derivatives markets. 
[More informations](https://vega.xyz/)

