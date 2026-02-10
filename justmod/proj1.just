set working-directory := '..'

# list project 1 just recipes
default:
    @just --list p1 --unsorted

deps:
    #!/usr/bin/env bash
    set -euo pipefail
    cd kvstore
    export PATH="$(go env GOPATH)/bin:$PATH"
    if ! command -v protoc >/dev/null 2>&1; then
        echo "protoc not found; please install Protocol Buffers (protoc) first."
        exit 1
    fi
    # Install protobuf plugins for Go
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
    # Initialize go module if not exists
    if [ ! -f go.mod ]; then
        go mod init madkv/kvstore
    fi
    # Get dependencies
    go get google.golang.org/grpc
    go get google.golang.org/protobuf
    go get github.com/google/btree
    # Generate protobuf code into gen/kvpb (module-aware output)
    mkdir -p gen/kvpb
    protoc --go_out=. --go_opt=module=madkv/kvstore \
           --go-grpc_out=. --go-grpc_opt=module=madkv/kvstore \
           proto/kvstore.proto
    # Tidy up dependencies
    go mod tidy
    echo "*******Dependencies installed and protobuf code generated*******"


build:
    cd kvstore && mkdir -p bin
    cd kvstore && go build -o bin/server ./server
    cd kvstore && go build -o bin/client ./client
    @echo "*******Built server and client binaries*******"

clean:
    rm -rf kvstore/bin
    rm -rf kvstore/gen
    @echo "*******Cleaned build artifacts*******"

# run your KV store server
server listen_addr="127.0.0.1:50051":
    cd kvstore && ./bin/server --listen {{listen_addr}}

# run you KV store client in stdin/out interface mode
client server="127.0.0.1:50051":
    cd kvstore && ./bin/client --server {{server}}

# run your KV store testcase 1 client
test1 server="127.0.0.1:50051":
    #!/usr/bin/env bash
    set -euo pipefail
    cd kvstore
    ./bin/client --server "{{server}}" <<'EOF'
    PUT a 1
    PUT b 2
    GET a
    SWAP a 3
    GET a
    SCAN a z
    DELETE b
    GET b
    STOP
    EOF

# run your KV store testcase 2 client
test2 server="127.0.0.1:50051":
    #!/usr/bin/env bash
    set -euo pipefail
    cd kvstore
    ./bin/client --server "{{server}}" <<'EOF'
    PUT x 9
    GET y
    SWAP y 1
    DELETE y
    SCAN z zz
    DELETE x
    PUT k 
    GET 
    SCAN
    STOP 55
    STOP
    EOF

# run your KV store testcase 3 client
test3 server="127.0.0.1:50051":
    # FIXME: your testcase 3 client run command here

# run your KV store testcase 4 client(s)
test4 server="127.0.0.1:50051":
    # FIXME: your testcase 4 client(s) run commands here
    #        for multi-client testcases, this recipe should run multiple
    #        concurrent client processes; you may write an additional runner
    #        utility or a scripting layer to manage multiple processes

# run your KV store testcase 5 client(s)
test5 server="127.0.0.1:50051":
    # FIXME: your testcase 5 client(s) run commands here

# kill all processes of your KV store system
kill:
    -pkill -f "kvstore/bin/server" 2>/dev/null
    -pkill -f "kvstore/bin/client" 2>/dev/null
    -pkill -f "bin/server" 2>/dev/null
    -pkill -f "bin/client" 2>/dev/null

# NOTE: feel free to add more recipes as you see fit...
#       also feel free to add extra parameters to the recipes as you see fit,
#       but don't change the existing parameters

python_run := if `which uv || true` != "" { "uv run" } else { "python3" }
tmpdir_prefix := "/tmp/madkv-p1"

# launch a long-running KV store service
service listen="0.0.0.0:50051":
    just p1::build
    just utils::build
    cargo run -p runner -r --bin service -- \
        --server-just-args p1::server "{{listen}}" \
        --manager-just-args none

# ensure a subdir under 'tmp/' exists
tmpdir subdir:
    mkdir -p "{{tmpdir_prefix}}/{{subdir}}"

# run a student-provided testcase
testcase num server="127.0.0.1:50051": (tmpdir "tests")
    just p1::build
    just "p1::test{{num}}" "{{server}}" \
        | tee "{{tmpdir_prefix}}/tests/test{{num}}.log"
    just p1::kill

# run a fuzz testing scenario
fuzz nclis conflict="no" server="127.0.0.1:50051": (tmpdir "fuzz")
    just p1::build
    just utils::build
    cargo run -p runner -r --bin fuzzer -- \
        --num-clis "{{nclis}}" \
        {{ if conflict == "no" { "" } else { "--conflict" } }} \
        --client-just-args p1::client "{{server}}" \
        | tee "{{tmpdir_prefix}}/fuzz/fuzz-{{nclis}}-{{conflict}}.log"
    just p1::kill

# run a YCSB benchmark workload
bench nclis wload server="127.0.0.1:50051": (tmpdir "bench")
    just p1::build
    just utils::build
    just utils::ycsb
    cargo run -p runner -r --bin bencher -- \
        --num-clis "{{nclis}}" \
        --workload "{{wload}}" \
        --client-just-args p1::client "{{server}}" \
        | tee "/tmp/madkv-p1/bench/bench-{{nclis}}-{{wload}}.log"
    just p1::kill

# generate .md report template from existing results (wip)
report:
    {{python_run}} sumgen/proj1.py
