echo "== Compiling assets"

GOFLOW_ASSETS_PATH=$GOPATH/src/github.com/fieldryand/goflow/assets

cd $GOFLOW_ASSETS_PATH && npm install && npm run build

echo "== Done compiling assets"
