/*
proxy-basecoin is an example script that sets up a proxy
that knows about the basic basecoin types (sendtx and accounts).

It can be extended to add support for basecoin plugins,
or as a source of inspiration to configure a proxy for your
own app.

If you run the basecoin demo app locally (from the data dir),
try the following:

# get the tm chain id:
curl http://localhost:46657/status | jq .result[1].node_info.network

proxy-basecoin -tmchain=test-chain-A8iHWI -chain=test_chain_id -rpc=localhost:46657

curl http://localhost:8108/keys/
curl -XPOST http://localhost:8108/keys/ -d '{"name": "alice", "passphrase": "1234567890"}'
curl -XPOST http://localhost:8108/keys/ -d '{"name": "bobby", "passphrase": "1234567890"}'
curl http://localhost:8108/keys/

# -> at this point, grab that address, but it in the genesis for
# basecoin, so you are rich, and restart the basecoin server ;)

## TODO: working examples here

# query no data
curl http://localhost:8108/query/store/01234567

# get an account (by path, or knowing the special prefix)
curl http://localhost:8108/query/account/1B1BE55F969F54064628A63B9559E7C21C925165
curl http://localhost:8108/query/store/626173652f612f1B1BE55F969F54064628A63B9559E7C21C925165

# 626173652f612f <- this is the magic base/a/ prefix for accounts in hex
# 1B1BE55F969F54064628A63B9559E7C21C925165 <- address with coins

# get proof by complete key
# TODO: currently fails, complaining about validator sigs
curl http://localhost:8108/proof/626173652f612f1B1BE55F969F54064628A63B9559E7C21C925165

# post a tx
# use addressed returned from your keys call above
# input is alice, output is bob
curl -XPOST http://localhost:8108/txs/ -d '{
  "name": "alice",
  "passphrase": "1234567890",
  "data": {
    "type": "sendtx",
    "data": {
      "gas": 22,
      "fee": {"denom": "ETH", "amount": 1},
      "inputs": [{
        "address": "4D8908785EC867139CA02E71A717C01FA506B96A",
        "coins": [{"denom": "ETH", "amount": 21}],
        "sequence": 1,
        "pub_key": {
          "type": "ed25519",
          "data": "D7FB176319AF0C126C4C4C7851CF7C66340E7DF8763F0AA9700EBAE32A955401"
        }
      }],
      "outputs": [{
        "address": "9F31A3AC6B1468402AAC5593AE9E82A041457F5F",
        "coins": [{"denom": "ETH", "amount": 20}]
      }]
    }
  }
}'

# and try a special escrow type.... using the trader plugins
curl -XPOST http://localhost:8108/txs/ -d '{
  "name": "alice",
  "passphrase": "1234567890",
  "data": {
    "type": "apptx",
    "data": {
      "name": "escrow",
      "gas": 22,
      "fee": {"denom": "ETH", "amount": 1},
      "input": {
        "address": "4D8908785EC867139CA02E71A717C01FA506B96A",
        "coins": [{"denom": "ETH", "amount": 21}],
        "sequence": 2,
      },
      "type": "create",
      "appdata": {
        "recipient": "9F31A3AC6B1468402AAC5593AE9E82A041457F5F",
        "arbiter": "12468402AAC55931A3AC6B1468E82A04145"
      },
    }
  }
}'


*/
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/tendermint/go-keys/cryptostore"
	"github.com/tendermint/go-keys/storage/filestorage"
	"github.com/tendermint/light-client/extensions/basecoin"
	"github.com/tendermint/light-client/proxy"
	"github.com/tendermint/tendermint/rpc/client"
)

var (
	port      = flag.Int("port", 8108, "port for proxy server")
	rpcAddr   = flag.String("rpc", "", "url for tendermint rpc server")
	chainID   = flag.String("chain", "", "id of the basecoin app (from basecoin genesis.json)")
	tmChainID = flag.String("tmchain", "", "chain id from tendermint (for signing blocks)")
	keydir    = flag.String("keydir", ".keys", "directory to store private keys")
)

// TODO: add cors and unix-socket support like over in verifier
func main() {
	flag.Parse()

	// TODO: make these actually do something
	vr := basecoin.NewBasecoinValues()
	sr := basecoin.NewBasecoinTx(*chainID)

	if *chainID == "" {
		fmt.Println("You must specify -chain with the basecoin chain_id")
		return
	}
	if *tmChainID == "" {
		fmt.Println("You must specify -tmchain with the tendermint chain_id")
		return
	}
	if *rpcAddr == "" {
		fmt.Println("You must specify -rpc with the location of a tendermint node")
		return
	}

	// set up all the pieces based on config
	r := mux.NewRouter()
	cstore := cryptostore.New(
		cryptostore.SecretBox,
		filestorage.New(*keydir),
	)
	node := client.NewHTTP(*rpcAddr, "/websocket")
	proxy.RegisterDefault(r, cstore, node, sr, vr, *tmChainID)

	// TODO: add some awesome middlewares...

	// now, just run the server and bind to localhost for security
	bind := fmt.Sprintf("127.0.0.1:%d", *port)
	log.Fatal(http.ListenAndServe(bind, r))

}
